# `dabazo` — build spec

## What it is
A cross-platform CLI tool, written in Go, for installing, running, and operating database engines during development and debugging. MVP supports PostgreSQL; the internal architecture is built around an **engine driver** interface so MySQL / MariaDB / SQLite / etc. can be added later without restructuring commands.

Primary use case: spin up a local database instance, create users, apply migrations, and snapshot data for debugging — with the same CLI shape on Linux, macOS, and Windows.

## High-level principles

1. **Go, single static binary.** No runtime dependency beyond the OS's native package manager and the installed DB's own client tools (`psql`, `pg_dump`, …).
2. **Cross-platform.** Must work on Debian/Ubuntu, RHEL/Fedora, macOS, and Windows. Everything OS-specific lives behind a `PackageManager` abstraction.
3. **Native package managers.** `install` shells out to `apt-get` / `dnf` / `brew` / `winget` / `choco`. **Always prints the exact command it is about to run and the source it will use, then asks for confirmation, before executing.** No silent installs.
4. **Engine driver abstraction.** A Go `Engine` interface with one implementation (`postgres`) for MVP. Adding `mysql` later = new driver, no command changes.
5. **Instance registry.** Every installed instance is named and tracked in a local registry. Commands target instances by `--name`.
6. **Consistent CLI shape.** Every command is `dabazo <action> [--db <engine[:version]>] [--port <port>] [--name <instance-name>] <command-specific args>`.

## Instance registry

Location (OS-specific user config dir):
- Linux/macOS: `~/.config/dabazo/instances.json`
- Windows: `%APPDATA%\dabazo\instances.json`

Schema (per instance):
```json
{
  "name": "prod",
  "engine": "postgres",
  "version": "16",
  "port": 5432,
  "host": "localhost",
  "installedAt": "2026-04-14T10:00:00Z",
  "packageManager": "apt",
  "serviceName": "postgresql@16-main"
}
```

### `--name` resolution rule
- `install` **requires** `--name`.
- Every other command accepts `--name`:
  - **0 instances** in registry → error: "no instances registered; run `dabazo install` first".
  - **1 instance** → `--name` is optional; the single instance is used.
  - **>1 instances** → `--name` is **required**; omitting it errors with a list of available instance names.

## Global flags

| Flag | Meaning | Default |
|---|---|---|
| `--db <engine[:version]>` | Engine + optional version. Engine required on `install`; version defaults to **latest** available via the host's package manager. Other commands read engine/version from the registry entry, so the flag is ignored there. | — |
| `--port <port>` | TCP port the instance listens on. Required on `install`. Other commands read port from the registry. | — |
| `--name <instance-name>` | Logical instance name (see resolution rule above). | — |
| `--yes` / `-y` | Skip confirmation prompts (still prints the plan). Use in scripts. | `false` |
| `--help` / `-h` | Show help for a command. | — |

## Commands

### `dabazo install --db <engine[:version]> --port <port> --name <instance-name>`
Install and register a new DB instance.

Flow:
1. Resolve the engine driver from `--db` (MVP: only `postgres`).
2. Detect the host OS and native package manager.
3. Ask the driver what it needs to install (package names, repo setup, initdb steps).
4. **Print the plan** — exact commands, package source (e.g. `apt-get install -y postgresql-16 from deb.debian.org`), target data directory, and listener port.
5. **Prompt for confirmation**: `Proceed? [y/N]` — unless `--yes` is passed.
6. On confirm: run the commands, initialize the cluster, configure it to listen on `--port`, and register the instance in the registry under `--name`.
7. Leaves the instance **stopped** — operator runs `dabazo start` next.

Errors if `--name` already exists in the registry.

### `dabazo start [--name <instance-name>]`
Start the instance's DB process/service (via the same package manager or service manager used at install time). Idempotent: starting an already-running instance prints "already running" and exits 0.

### `dabazo stop [--name <instance-name>]`
Stop the instance. Idempotent: stopping an already-stopped instance prints "already stopped" and exits 0.

### `dabazo config user <username> [--name <instance-name>]`
Create a DB role named `<username>` with a freshly generated random password (32-char base62, no special shell chars) and a database of the same name owned by it.

Side effect: writes a file in the **current working directory** literally named `<username>` (no extension) with:
```
DB_URL=jdbc:postgresql://localhost:<port>/<username>
DB_USER=<username>
DB_PASSWORD=<generated>
```
Permissions: `0600`, owned by the invoking user. Refuses to overwrite an existing file.

Fails loudly if the role already exists in the DB (no silent rotation).

### `dabazo migrate <filepath> [--name <instance-name>] [--user <username>]`
Apply the SQL file at `<filepath>` to the instance's database.

- If `--user` is passed, uses that role; otherwise loads credentials from a file named after the role in the current working directory (same format as `config user` output).
- Exits non-zero on any SQL error and prints the full client error output.

### `dabazo snapshot <db> <path> [--name <instance-name>]`
Dump the entire database `<db>` to `<path>` as plain SQL, importable on another instance.

- Prompts interactively for DB user (`read`) and password (hidden input).
- Runs `pg_dump -Fp` (plain SQL, schema + data) against `localhost:<port>/<db>` using the prompted credentials via `PGPASSWORD` for the child process only — never logged, never persisted.
- Refuses to overwrite an existing file at `<path>` unless `--force` is passed.
- Exits non-zero on failure and prints stderr.

### `dabazo registry add --db <engine[:version]> --port <port> --name <instance-name>`
Register an **already-existing** DB instance that was **not** installed by dabazo (e.g. a pre-existing local Postgres, a dev container, a remote dev DB). No packages are installed, no service is touched — only the registry is modified.

- `--db`, `--port`, `--name` are all required.
- Optional `--host <host>` (default `localhost`) for remote instances.
- Fails if `--name` already exists in the registry.
- The resulting registry entry has `packageManager: "external"` and no `serviceName`; `dabazo start` / `dabazo stop` / `dabazo uninstall` refuse to operate on external entries and print a clear message ("instance was added via `registry add`; dabazo does not manage its lifecycle").
- `config user`, `migrate`, and `snapshot` work normally against external entries (they only need host+port+credentials).

Example:
```
dabazo registry add --db postgres:16 --port 5432 --name legacy
```

### `dabazo registry remove --name <instance-name>`
Remove an entry from the registry **without** uninstalling anything or touching the DB itself. Safe to run on both dabazo-installed and externally-added entries.

- `--name` is always required, even when only one instance is registered (explicit, to avoid accidental removal).
- Errors if the named instance does not exist.
- Does not stop the DB, does not remove data, does not uninstall packages. For the full teardown of a dabazo-managed instance, use `dabazo uninstall`.

Example:
```
dabazo registry remove --name legacy
```

### `dabazo list`
Print the registry: one line per instance with `name`, `engine:version`, `port`, running/stopped status.

### `dabazo uninstall [--name <instance-name>]`
Stop the instance (if running), uninstall the packages via the same package manager (with confirmation and printed plan, same as `install`), and remove the registry entry. Does **not** delete the data directory unless `--purge` is passed.

### `dabazo help [command]`
Print help.

- `dabazo help` — top-level overview: list of commands, global flags, one-line description of each command.
- `dabazo help <command>` — detailed help for one command: synopsis, every flag with type and default, expected arguments, examples, exit codes.
- `dabazo --help` / `dabazo -h` — same as `dabazo help`.
- `dabazo <command> --help` — same as `dabazo help <command>`.

Each per-command help must include:
- Synopsis line (`dabazo install --db <engine[:version]> --port <port> --name <instance-name>`).
- Description.
- All flags with `<type>` and default.
- Positional argument list.
- 2–3 worked examples.
- Exit code legend.

## Engine driver interface (Go)

```go
type Engine interface {
    // Name returns the canonical engine name, e.g. "postgres".
    Name() string

    // Plan returns the install plan (commands + rationale) for the given
    // version/port/OS without executing anything. Used for the confirmation prompt.
    Plan(version string, port int, pm PackageManager) (InstallPlan, error)

    // Install executes a plan previously produced by Plan.
    Install(plan InstallPlan) error

    // Start / Stop control the running service.
    Start(inst Instance) error
    Stop(inst Instance) error

    // Operations
    CreateUser(inst Instance, username, password string) error
    ApplySQL(inst Instance, user, password, filepath string) error
    Dump(inst Instance, db, user, password, outPath string) error
}

type PackageManager interface {
    Name() string                          // "apt", "dnf", "brew", "winget", "choco"
    InstallCommand(pkgs []string) []string // exact argv
    UninstallCommand(pkgs []string) []string
    ServiceStart(svc string) []string
    ServiceStop(svc string) []string
}
```

MVP drivers:
- `engines/postgres` — implements `Engine`.
- `pkgmgr/{apt,dnf,brew,winget,choco}` — each implements `PackageManager`. Detection at startup via OS probe.

## Confirmation prompt UX

Any command that mutates the host system (`install`, `uninstall`) must print a block like this before prompting:

```
dabazo will run the following to install postgres:16 on port 5432:

  Package manager : apt
  Source          : deb.debian.org (main)
  Packages        : postgresql-16, postgresql-client-16
  Commands:
    sudo apt-get update
    sudo apt-get install -y postgresql-16 postgresql-client-16
  Post-install:
    initdb data directory: /var/lib/postgresql/16/main
    listening port:        5432
    registered as:         prod

Proceed? [y/N]
```

`--yes` skips the prompt but still prints the block for audit.

## Exit codes

| Code | Meaning |
|---|---|
| 0 | Success |
| 1 | Generic failure |
| 2 | Usage error (bad flags / missing `--name` when required) |
| 3 | Instance not found in registry |
| 4 | Instance already exists (on `install`) |
| 5 | Package manager operation failed |
| 6 | DB operation failed (SQL error, dump failed, auth failed) |
| 7 | User aborted confirmation |

## Project layout

```
dabazo/
├── cmd/
│   └── dabazo/
│       └── main.go          # entrypoint, cobra root command
├── internal/
│   ├── cli/                 # cobra command definitions (install, start, stop, ...)
│   ├── registry/            # instances.json load/save, --name resolution
│   ├── engines/
│   │   ├── engine.go        # Engine interface
│   │   └── postgres/        # MVP driver
│   ├── pkgmgr/
│   │   ├── pkgmgr.go        # PackageManager interface + OS detection
│   │   ├── apt/
│   │   ├── dnf/
│   │   ├── brew/
│   │   ├── winget/
│   │   └── choco/
│   ├── prompt/              # confirmation + hidden password input
│   └── secret/              # random password generation
├── go.mod
└── README.md
```

Dependencies:
- `github.com/spf13/cobra` for commands and help generation.
- `golang.org/x/term` for hidden password input.
- Standard library for everything else.

## Build

- `go build ./cmd/dabazo` produces a single static binary.
- CI matrix: linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64.
- Release artifacts: one binary per target, plus SHA256 sums.

## Acceptance criteria

- On a fresh Debian box: `dabazo install --db postgres:16 --port 5432 --name dev` prints the plan, asks for confirmation, installs via `apt-get`, and registers `dev`.
- `dabazo start` (with one instance) starts the service; `dabazo stop` stops it.
- `dabazo config user alice` creates the role + DB and writes `./alice` with `DB_URL` / `DB_USER` / `DB_PASSWORD` (mode `0600`).
- `dabazo migrate ./V1__setup.sql` applies the SQL using credentials from `./<user>`.
- `dabazo snapshot alice /tmp/alice.sql` prompts for user+password, writes an importable plain-SQL dump.
- A second instance (`dabazo install --db postgres:17 --port 5433 --name next`) makes every later command require `--name`.
- `dabazo help` and `dabazo help <command>` print complete, accurate usage.
- The same binary works on macOS (via `brew`) and Windows (via `winget` or `choco`) with the same CLI.

# dabazo

Cross-platform CLI for installing, running, and operating database engines during development.

dabazo is a cross-platform CLI tool for installing, running, and operating database engines during development and debugging. MVP supports PostgreSQL. Use the same CLI shape on Linux, macOS, and Windows to spin up a local database instance, create users, apply migrations, and snapshot data for debugging.

## Features

- **One-command install** -- install a database engine via your native package manager with a single command
- **Cross-platform** -- works on Linux (apt, dnf), macOS (Homebrew), and Windows (winget, Chocolatey) with automatic package manager detection
- **Instance registry** -- track multiple named database instances with a local JSON registry
- **User and credential management** -- create database roles with auto-generated passwords and `.env`-style credential files
- **SQL migrations** -- apply SQL files to any registered instance
- **Database snapshots** -- dump an entire database to a portable SQL file
- **External instance support** -- register pre-existing databases not installed by dabazo
- **Minimal dependencies** -- built with Go standard library plus `golang.org/x/term`

## Requirements

- **Go 1.26.2+** (for building from source)
- **Supported operating systems and package managers:**

| OS | Package Manager |
|---|---|
| Linux (Debian/Ubuntu) | apt-get |
| Linux (Fedora/RHEL) | dnf |
| macOS | Homebrew (brew) |
| Windows | winget, Chocolatey (choco) |

## Installation

### From source

Clone and build:

```bash
git clone https://github.com/your-org/dabazo.git
cd dabazo
go build -o dabazo ./cmd/dabazo
```

## Quick Start

```bash
# 1. Install PostgreSQL 16 on port 5432
dabazo install --db postgres:16 --port 5432 --name dev

# 2. Start the instance
dabazo start --name dev

# 3. Create a database user with auto-generated credentials
dabazo config user alice --name dev

# 4. Apply a migration
dabazo migrate ./V1__setup.sql --user alice --name dev
```

## Command Reference

### `dabazo install`

Install and register a new database instance via the native package manager. Prints the exact commands it will run and prompts for confirmation before executing. The instance is left stopped after installation.

**Usage:**

```
dabazo install [flags]
```

**Required flags:** `--db`, `--port`, `--name`

**Examples:**

```bash
dabazo install --db postgres:16 --port 5432 --name dev
dabazo install --db postgres:17 --port 5433 --name next -y
```

**Notes:**
- Version defaults to `16` if omitted from `--db` (e.g., `--db postgres`)
- The package manager is auto-detected based on your OS
- After install, the instance is registered in the local registry

---

### `dabazo start`

Start the database service for a registered instance. Idempotent: starting an already-running instance prints a message and exits successfully.

**Usage:**

```
dabazo start [flags]
```

**Examples:**

```bash
dabazo start
dabazo start --name dev
```

**Notes:** Cannot start instances added via `registry add` (external instances).

---

### `dabazo stop`

Stop the database service for a registered instance. Idempotent: stopping an already-stopped instance prints a message and exits successfully.

**Usage:**

```
dabazo stop [flags]
```

**Examples:**

```bash
dabazo stop
dabazo stop --name dev
```

**Notes:** Cannot stop instances added via `registry add` (external instances).

---

### `dabazo config user`

Create a database role with a randomly generated password and a database of the same name. Writes credentials to a file named after the username in the current directory (mode 0600). The credential file contains `DB_URL`, `DB_USER`, and `DB_PASSWORD` in `.env` format.

**Usage:**

```
dabazo config user <username> [flags]
```

**Examples:**

```bash
dabazo config user alice
dabazo config user bob --name dev
```

**Notes:**
- Refuses to overwrite an existing credential file
- The credential file is used by `dabazo migrate`

---

### `dabazo migrate`

Apply a SQL file to the instance's database using the credentials from a previously created user.

**Usage:**

```
dabazo migrate <filepath> [flags]
```

**Required flags:** `--user`

| Local Flag | Type | Default | Description |
|---|---|---|---|
| `--user` | string | `""` | Database role to use; must match an existing credential file in the current directory |

**Examples:**

```bash
dabazo migrate ./V1__setup.sql --user alice
dabazo migrate ./V2__data.sql --user bob --name dev
```

---

### `dabazo snapshot`

Dump an entire database to a file as plain SQL (schema and data). Prompts interactively for the database user and password. The dump is importable on another instance.

**Usage:**

```
dabazo snapshot <db> <path> [flags]
```

| Local Flag | Type | Default | Description |
|---|---|---|---|
| `--force` | bool | `false` | Overwrite existing output file |

**Examples:**

```bash
dabazo snapshot alice /tmp/alice.sql
dabazo snapshot mydb ./backup.sql --name dev --force
```

---

### `dabazo list`

Print the registry: one line per instance showing name, engine version, port, and running/stopped status. Status is determined by attempting a TCP connection to the instance port.

**Usage:**

```
dabazo list
```

**Example output:**

```
NAME            ENGINE          PORT     STATUS
dev             postgres:16     5432     running
next            postgres:17     5433     stopped
```

---

### `dabazo registry add`

Register an already-existing database instance that was not installed by dabazo. No packages are installed and no service is touched. The instance is recorded as external and cannot be started, stopped, or uninstalled by dabazo.

**Usage:**

```
dabazo registry add [flags]
```

**Required flags:** `--db`, `--port`, `--name`

| Local Flag | Type | Default | Description |
|---|---|---|---|
| `--host` | string | `"localhost"` | Host address for the instance |

**Examples:**

```bash
dabazo registry add --db postgres:16 --port 5432 --name legacy
dabazo registry add --db postgres:16 --port 5432 --name remote --host 10.0.0.5
```

---

### `dabazo registry remove`

Remove an entry from the registry without uninstalling anything or touching the database. Safe for both dabazo-installed and externally-added entries.

**Usage:**

```
dabazo registry remove [flags]
```

**Required flags:** `--name`

**Example:**

```bash
dabazo registry remove --name legacy
```

---

### `dabazo uninstall`

Stop the instance (if running), uninstall the packages via the same package manager used at install time, and remove the registry entry. Prompts for confirmation before executing.

**Usage:**

```
dabazo uninstall [flags]
```

| Local Flag | Type | Default | Description |
|---|---|---|---|
| `--purge` | bool | `false` | Also delete the data directory |

**Examples:**

```bash
dabazo uninstall --name dev
dabazo uninstall --name dev --purge -y
```

**Notes:** Cannot uninstall instances added via `registry add` (external instances).

## Global Flags

These flags are available on all commands:

| Flag | Type | Default | Description |
|---|---|---|---|
| `--name` | string | `""` | Logical instance name (optional when only one instance is registered) |
| `--db` | string | `""` | Engine and version in `engine[:version]` format (e.g., `postgres:16`) |
| `--port` | int | `0` | TCP port the instance listens on |
| `--yes` / `-y` | bool | `false` | Skip confirmation prompts |

## Registry

dabazo maintains a JSON registry of all known instances at:

```
<os-config-dir>/dabazo/instances.json
```

Where `<os-config-dir>` is the value returned by Go's `os.UserConfigDir()` (e.g., `~/.config` on Linux, `~/Library/Application Support` on macOS, `%AppData%` on Windows).

### Name resolution rule

The `--name` flag follows this resolution logic:

- **0 instances** in registry: error, run `dabazo install` first
- **1 instance** in registry: `--name` is optional, the single instance is used automatically
- **Multiple instances** in registry: `--name` is required

## Exit Codes

| Code | Constant | Meaning |
|---|---|---|
| 0 | ExitSuccess | Command completed successfully |
| 1 | ExitGeneric | Generic failure |
| 2 | ExitUsage | Usage error (bad flags, missing required args) |
| 3 | ExitNotFound | Instance not found in registry |
| 4 | ExitAlreadyExists | Instance name already taken |
| 5 | ExitPkgManager | Package manager operation failed |
| 6 | ExitDBOperation | Database operation failed |
| 7 | ExitAborted | User aborted a confirmation prompt |

## Supported Database Engines

| Engine | `--db` value | Status |
|---|---|---|
| PostgreSQL | `postgres[:version]` | MVP (fully implemented) |

## Contributing

Contributions are welcome. Please open an issue to discuss proposed changes before submitting a pull request.

To build and test locally:

```bash
go build ./...
go test ./...
```

## License

This project is not yet licensed. A license will be added in a future release.

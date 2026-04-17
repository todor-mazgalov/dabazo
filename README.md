<p>
  <img src="./assets/logo-h.png" alt="Dabazo" height="80">
</p>

# Dabazo

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
- **Delete operations** -- drop users, databases, and schemas with confirmation prompts
- **Interactive mode** -- use `--interactive` / `-it` to be prompted for missing required parameters
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

### With `go install`

```bash
go install github.com/todor-mazgalov/dabazo/cmd/dabazo@latest
```

The resulting binary lands in `$(go env GOBIN)` (or `$GOPATH/bin`).

### From source

Clone and build:

```bash
git clone https://github.com/todor-mazgalov/dabazo.git
cd dabazo
go build -o dabazo ./cmd/dabazo
```

## Quick Start

```bash
# 1. Install PostgreSQL 16 on port 5432
dabazo install --engine postgres:16 --port 5432 --name dev

# 2. Start the instance
dabazo start --name dev

# 3. Create a database user with auto-generated credentials
dabazo create user alice --name dev

# 4. Apply a migration
dabazo migrate ./V1__setup.sql --user alice --name dev

# 5. Delete the user when done
dabazo delete user alice --name dev
```

## Command Reference

### `dabazo install`

Install and register a new database instance via the native package manager. Prints the exact commands it will run and prompts for confirmation before executing. The instance is left stopped after installation.

**Usage:**

```
dabazo install [flags]
```

**Required flags:** `--engine`, `--port`, `--name`

**Examples:**

```bash
dabazo install --engine postgres:16 --port 5432 --name dev
dabazo install -e postgres:17 -p 5433 -n next -y
```

**Notes:**
- Version defaults to `16` if omitted from `--engine` (e.g., `--engine postgres`)
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

**Notes:**
- The resolved instance name is printed as the first output line
- Cannot start instances added via `registry add` (external instances)

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

**Notes:**
- The resolved instance name is printed as the first output line
- Cannot stop instances added via `registry add` (external instances)

---

### `dabazo create user`

Create a database role with a randomly generated password and a database of the same name. Writes credentials to a file named after the username in the current directory (mode 0600). The credential file contains `DB_URL`, `DB_USER`, and `DB_PASSWORD`.

**Usage:**

```
dabazo create user <username> [flags]
```

| Local Flag | Short | Type | Default | Description |
|---|---|---|---|---|
| `--output` | `-o` | string | `java` | Credential file format: `java`, `shell`/`bash`, `bat`/`cmd`, `pwsh`/`powershell` |
| `--url-format` | `-uf` | string | `jdbc` | URL style for `DB_URL`: `jdbc` (e.g. `jdbc:postgresql://...`) or `plain` (e.g. `postgresql://...`) |

**Output format examples:**

| Format | Example line |
|---|---|
| `java` (default) | `DB_USER=alice` |
| `shell` / `bash` | `export DB_USER=alice` |
| `bat` / `cmd` | `set DB_USER=alice` |
| `pwsh` / `powershell` | `$env:DB_USER = "alice"` |

**Examples:**

```bash
dabazo create user alice
dabazo create user bob --name dev
dabazo create user alice -o bash
dabazo create user alice -o pwsh -uf plain
```

**Notes:**
- Refuses to overwrite an existing credential file
- The credential file is used by `dabazo migrate`, `dabazo create schema`, and `dabazo snapshot`
- The JDBC URL prefix respects the engine (e.g. `jdbc:postgresql://`, `jdbc:mysql://`)

---

### `dabazo create database`

Create a database owned by an existing role. Runs as the PostgreSQL superuser and does not require a credential file. The owner role must already exist (created via `dabazo create user` or externally).

**Usage:**

```
dabazo create database <database-name> [flags]
```

**Required flags:** `--user`

| Local Flag | Short | Type | Default | Description |
|---|---|---|---|---|
| `--user` | `-u` | string | `""` | Owner role for the new database |

**Examples:**

```bash
dabazo create database app -u alice
dabazo create database reports -u alice --name dev
```

---

### `dabazo create schema`

Create a schema inside an existing database. Connects as the role identified by `--user`, reading the password from a credential file named after the user in the current directory.

**Usage:**

```
dabazo create schema <schema-name> [flags]
```

**Required flags:** `--user`, `--database`

| Local Flag | Short | Type | Default | Description |
|---|---|---|---|---|
| `--user` | `-u` | string | `""` | Role to connect as |
| `--database` | `-db` | string | `""` | Database to create the schema in |

**Examples:**

```bash
dabazo create schema audit -db app -u alice
dabazo create schema reporting -db app -u alice --name dev
```

---

### `dabazo delete user`

Drop a database role. Prompts for confirmation before executing. The resolved instance name is printed as the first output line.

**Usage:**

```
dabazo delete user <username> [flags]
```

**Examples:**

```bash
dabazo delete user alice
dabazo delete user bob --name dev -y
```

**Notes:**
- Will fail if the role owns databases or other objects; handle dependencies manually before dropping
- Use `--yes` / `-y` to skip the confirmation prompt

---

### `dabazo delete database`

Drop a database. Prompts for confirmation before executing. The resolved instance name is printed as the first output line.

**Usage:**

```
dabazo delete database <database-name> [flags]
```

**Examples:**

```bash
dabazo delete database app
dabazo delete database reports --name dev -y
```

**Notes:** Will fail if there are active connections to the database.

---

### `dabazo delete schema`

Drop a schema from a database. Connects as the role identified by `--user`, reading the password from a credential file named after the user in the current directory. The resolved instance name is printed as the first output line.

**Usage:**

```
dabazo delete schema <schema-name> [flags]
```

**Required flags:** `--user`, `--database`

| Local Flag | Short | Type | Default | Description |
|---|---|---|---|---|
| `--user` | `-u` | string | `""` | Role to connect as |
| `--database` | `-db` | string | `""` | Database containing the schema |

**Examples:**

```bash
dabazo delete schema audit -db app -u alice
dabazo delete schema public -db app -u alice --name dev -y
```

---

### `dabazo migrate`

Apply a SQL file to the instance's database using the credentials from a previously created user.

**Usage:**

```
dabazo migrate <filepath> [flags]
```

**Required flags:** `--user`

| Local Flag | Short | Type | Default | Description |
|---|---|---|---|---|
| `--user` | `-u` | string | `""` | Database role to use; must match an existing credential file in the current directory |
| `--database` | `-db` | string | `""` (falls back to `--user`) | Target database name |
| `--schema` | `-s` | string | `""` | Session `search_path` (applied via `PGOPTIONS`) |
| `--host` | `-h` | string | `""` (uses registry) | Override instance host from the registry |

`--port` / `-p` (global) also overrides the instance port when set.

**Examples:**

```bash
dabazo migrate ./V1__setup.sql --user alice
dabazo migrate ./V2__data.sql -u bob -n dev -db app -s public
dabazo migrate ./V3.sql -u alice -h 10.0.0.5 -p 5433
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

**Required flags:** `--engine`, `--port`, `--name`

| Local Flag | Type | Default | Description |
|---|---|---|---|
| `--host` | string | `"localhost"` | Host address for the instance |

**Examples:**

```bash
dabazo registry add --engine postgres:16 --port 5432 --name legacy
dabazo registry add -e postgres:16 -p 5432 -n remote --host 10.0.0.5
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

**Notes:**
- The resolved instance name is printed as the first output line
- Cannot uninstall instances added via `registry add` (external instances)

## Global Flags

These flags are available on all commands:

| Flag | Short | Type | Default | Description |
|---|---|---|---|---|
| `--name` | `-n` | string | `""` | Logical instance name (optional when only one instance is registered) |
| `--engine` | `-e` | string | `""` | Engine and version in `engine[:version]` format (e.g., `postgres:16`) |
| `--port` | `-p` | int | `0` | TCP port the instance listens on |
| `--yes` | `-y` | bool | `false` | Skip confirmation prompts |
| `--interactive` | `-it` | bool | `false` | Enable interactive prompting for missing required parameters |

**Help:** use `--help` on any command. The short `-h` is reserved for `--host` on commands that accept it (e.g. `dabazo migrate`); all other commands accept `-h` as a help alias.

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

| Engine | `--engine` value | Status |
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

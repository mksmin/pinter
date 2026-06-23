# Project Structure

Pinter MVP is CLI + TUI. One Go binary. No GUI frameworks.

## Top Level

```text
pinter/
  cmd/
    pinter/                 CLI entrypoint
  docs/
    PROJECT_STRUCTURE.md    This guide
  internal/                 Go application code
  justfile                  Local build/test recipes
  skiff/                    Local reference project, ignored by git
  go.mod
  README.md
```

## Runtime Flow

CLI path:

```text
cmd/pinter/main.go
  -> internal/config
  -> internal/app.Service
  -> internal/tui.Run when no args
  -> internal/hosts.Repository
  -> internal/db SQLite
```

Connect path:

```text
cmd/pinter connect <alias>
  -> internal/app.Service.Connect
  -> internal/hosts.Repository.GetByAlias
  -> internal/terminal.Launcher.Open
  -> external terminal launcher
  -> system ssh
  -> internal/hosts.Repository.MarkConnected
  -> internal/hosts.Repository.AddHistory
```

Import path:

```text
cmd/pinter import-ssh-config
  -> internal/app.Service.ImportSSHConfig
  -> internal/sshconfig.ParseFile
  -> internal/hosts.Repository.UpsertByAlias
  -> SQLite
```

## Packages

### `cmd/pinter`

CLI only. Parses args and flags. Prints output. Exits with non-zero code on error.

Commands:

```text
pinter
pinter add
pinter list
pinter connect
pinter history
pinter import-ssh-config
```

No args opens TUI. Subcommands stay script-friendly.

Keep handlers thin. Real behavior belongs in `internal/app`.

### `internal/app`

Use-case layer. Coordinates repository, SSH config parser, and terminal launcher.

Owns app operations:

- add host
- list hosts
- connect host
- import SSH config
- read history

### `internal/config`

Resolves local paths.

Env vars:

```text
PINTER_DB_PATH     Exact SQLite file path
PINTER_DATA_DIR    Directory containing pinter.sqlite
PINTER_TERMINAL    Windows terminal: auto, wt, pwsh, powershell, cmd
```

Default macOS DB:

```text
~/Library/Application Support/pinter/pinter.sqlite
```

### `internal/db`

SQLite open + schema migration.

Tables:

```text
hosts
connection_history
```

Schema lives in Go code for now. If schema changes, update:

- `internal/db/db.go`
- `internal/hosts/repository.go`
- `internal/model/model.go`

### `internal/hosts`

Repository layer. Knows SQL. Does not know CLI or Terminal.app.

Owns:

- create host
- update/upsert host
- list/search hosts
- get host by alias or ID
- mark last connected
- write/read history

### `internal/ids`

Random ID helper.

Current formats:

```text
hst_<random_hex>
hstlog_<random_hex>
```

### `internal/model`

Shared data structs:

- `Host`
- `HostInput`
- `ConnectionHistory`

Keep this package boring. Types here cross package boundaries.

### `internal/sshconfig`

OpenSSH config parser.

Supported directives:

- `Host`
- `HostName`
- `Port`
- `User`
- `IdentityFile`

`IdentityFile` expands:

- `~/`
- `~\`
- `$HOME`
- `%USERPROFILE%`
- other known `%VAR%` environment values

Unknown `%...%` values remain unchanged so OpenSSH tokens are not silently lost.

Skipped:

- `Host *`
- wildcard host patterns like `prod-*`
- `Include`

Tests:

```text
internal/sshconfig/parser_test.go
```

### `internal/terminal`

External terminal launcher.

macOS MVP:

```text
osascript -> Terminal.app -> ssh command
```

Windows launchers:

```text
PINTER_TERMINAL=auto -> wt -> pwsh -> powershell -> cmd
PINTER_TERMINAL=wt   -> Windows Terminal, then shell fallback
PINTER_TERMINAL=pwsh -> PowerShell 7
PINTER_TERMINAL=powershell -> Windows PowerShell
PINTER_TERMINAL=cmd  -> cmd.exe
```

SSH command arguments are built first, then rendered for the selected shell:

- POSIX shell quoting for macOS/Linux.
- cmd quoting for `cmd.exe`.
- PowerShell quoting for PowerShell launchers.

If launcher logic grows further, split:

```text
launcher_darwin.go
launcher_windows.go
launcher_unix.go
```

Tests:

```text
internal/terminal/terminal_test.go
```

### `internal/tui`

Terminal UI. No GUI framework.

Uses:

```text
golang.org/x/term
```

Owns:

- ASCII logo
- main menu
- hosts browser
- add host prompts
- import prompt
- history view
- status view

Controls:

```text
Up/Down or K/J
Enter
B Back
Q Quit
```

Keep this package UI-only. Business logic stays in `internal/app`.

## Data Model

### `hosts`

SSH inventory.

Fields:

```text
id
alias
hostname
port
username
identity_file
notes
favorite
last_connected_at
created_at
updated_at
```

### `connection_history`

Connection launch log.

Fields:

```text
id
host_id
alias_snapshot
command
started_at
exit_status
terminal_app
```

`exit_status` is nullable and currently not filled. Terminal session runs outside Pinter.

## Ignored Files

```text
.cache/
.idea/
build/
pinter
skiff/
```

`skiff/` stays on disk as reference, but new git history ignores it.

## Where To Add Next Features

Edit host:

```text
internal/hosts.Update
internal/app.UpdateHost
cmd/pinter edit
```

Delete host:

```text
internal/hosts.Delete
internal/app.DeleteHost
cmd/pinter delete
```

Better SSH config parser:

```text
internal/sshconfig
```

Windows terminal:

```text
internal/terminal
```

Build/test recipes:

```text
justfile
```

Encrypted credentials later:

```text
internal/vault
internal/db schema
internal/model credential refs
```

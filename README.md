# Pinter

Local SSH keeper MVP. CLI + TUI. Stores hosts in SQLite, imports `~/.ssh/config` on explicit command, opens SSH in macOS Terminal via system `ssh`.

No GUI frameworks. No Wails. No frontend.

## Requirements

- Go 1.26+
- macOS Terminal.app for `connect`

## Data

Default DB path:

```text
~/Library/Application Support/pinter/pinter.sqlite
```

Use temp DB for testing:

```bash
export PINTER_DB_PATH="/private/tmp/pinter-dev.sqlite"
```

Use project-local DB:

```bash
export PINTER_DATA_DIR="$PWD/.cache/pinter"
```

`GOCACHE="$PWD/.cache/go-build"` only needed inside restricted sandbox. Normal local run does not need it.

## Run

Open TUI:

```bash
go run ./cmd/pinter
```

Help:

```bash
go run ./cmd/pinter help
```

Add host:

```bash
go run ./cmd/pinter add \
  --alias local \
  --host 127.0.0.1 \
  --user "$USER" \
  --notes "Local smoke test"
```

List hosts:

```bash
go run ./cmd/pinter list
```

Search:

```bash
go run ./cmd/pinter list -q local
```

Import SSH config, explicit only:

```bash
go run ./cmd/pinter import-ssh-config
```

Use custom SSH config path:

```bash
go run ./cmd/pinter import-ssh-config --path ./my-ssh-config
```

Connect:

```bash
go run ./cmd/pinter connect local
```

This opens Terminal.app with system `ssh`.

History:

```bash
go run ./cmd/pinter history
```

## Build Binary

```bash
go build -o build/pinter ./cmd/pinter
```

Run built binary:

```bash
./build/pinter list
```

Open TUI from built binary:

```bash
./build/pinter
```

## Verify

Run tests:

```bash
go test ./...
```

Build all packages:

```bash
go build ./...
```

Build CLI binary:

```bash
go build -o build/pinter ./cmd/pinter
```

## Smoke Test

```bash
export PINTER_DB_PATH="/private/tmp/pinter-smoke.sqlite"
rm -f "$PINTER_DB_PATH"

go run ./cmd/pinter add \
  --alias smoke \
  --host 127.0.0.1 \
  --user "$USER" \
  --notes "smoke test"

go run ./cmd/pinter list
go run ./cmd/pinter history
```

Optional Terminal check:

```bash
go run ./cmd/pinter connect smoke
```

## Current Limits

- TUI only. No GUI.
- No passwords/passphrases stored.
- Uses local `ssh`, key files, and `ssh-agent`.
- SSH config import skips wildcard hosts like `Host *` and `Host prod-*`.
- SSH config `Include` is not supported yet.
- History records launch time, not remote shell exit code.

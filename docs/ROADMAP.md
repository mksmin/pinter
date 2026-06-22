# Development Plan

Pinter goal: local SSH keeper, CLI + TUI, clean Go code, no GUI frameworks.

## Principles

- One small Go binary.
- TUI first, subcommands always script-friendly.
- SQLite local storage.
- No passwords/passphrases until vault design exists.
- Use system `ssh` and `ssh-agent`.
- Keep business logic in `internal/app`, not in TUI.

## Phase 0: Current Baseline

Status: done.

- CLI commands:
  - `add`
  - `list`
  - `connect`
  - `history`
  - `import-ssh-config`
- TUI main menu.
- SQLite schema:
  - `hosts`
  - `connection_history`
- macOS Terminal.app launcher.
- SSH config parser for basic directives.
- Docs:
  - README
  - project structure

## Phase 1: TUI Usability

Goal: TUI usable for daily host browsing.

Tasks:

- Host details view:
  - alias
  - target
  - key path
  - notes
  - last connected
- Search/filter inside TUI.
- Add host form improvements:
  - field validation before save
  - defaults visible
  - cancel per form
- Edit host flow.
- Delete host flow with confirmation.
- Better empty states.
- Better error/status messages.
- Terminal resize handling.

Acceptance:

- User can manage hosts without using subcommands.
- CLI commands still work unchanged.

## Phase 2: CLI Completeness

Goal: all core actions scriptable.

Tasks:

- `pinter edit <alias>`
- `pinter delete <alias>`
- `pinter show <alias>`
- `pinter notes <alias>`
- `pinter favorite <alias>`
- `pinter list --json`
- `pinter history --json`

Acceptance:

- Every TUI action has CLI equivalent.
- JSON output stable enough for scripts.

## Phase 3: Data Model

Goal: better organization without cloud/sync.

Tasks:

- Tags.
- Favorites.
- Groups/folders.
- Last used sorting.
- Host aliases uniqueness rules.
- Migration version table.
- DB backup/export command.
- DB import/restore command.

Acceptance:

- SQLite migrations explicit and reversible enough for MVP.
- User can backup and move local inventory.

## Phase 4: SSH Config Import

Goal: reliable import from real-world `~/.ssh/config`.

Tasks:

- Parse `Include`.
- Parse multiple `Host` aliases.
- Expand `~` and env vars in `IdentityFile`.
- Preserve unsupported directives in notes.
- Preview import in TUI before apply.
- Conflict handling:
  - skip
  - update
  - rename

Acceptance:

- Import large SSH configs without silent loss.
- Wildcards remain skipped unless template model exists.

## Phase 5: Terminal Launching

Goal: smoother connection launching.

Tasks:

- macOS launcher options:
  - Terminal.app
  - iTerm2 if installed
  - custom command template
- Copy SSH command action.
- Dry-run connect output.
- Record failed launcher errors clearly.
- Windows launcher:
  - Windows Terminal
  - PowerShell fallback
- Linux launcher:
  - `$TERMINAL`
  - common terminal fallback list

Acceptance:

- macOS works by default.
- Windows/Linux have MVP launch path.

## Phase 6: Security And Secrets

Goal: add secrets only with clear vault model.

Tasks:

- Decide vault scope:
  - no secrets
  - key passphrases only
  - passwords + keys
- Master password flow.
- Encryption at rest.
- Auto-lock.
- Clear memory best effort.
- Threat model doc.

Acceptance:

- No secret stored before encryption design is implemented.
- Docs explain what is and is not protected.

## Phase 7: Embedded Terminal Investigation

Goal: decide if Pinter should embed SSH terminal.

Research:

- PTY handling in Go.
- SSH client library options.
- Terminal UI rendering constraints.
- Copy/paste behavior.
- Resize behavior.
- Host key verification.

Possible paths:

- Keep external terminal only.
- Embed terminal in TUI.
- Add separate terminal mode.

Acceptance:

- Written decision before implementation.

## Phase 8: Polish

Goal: pleasant daily tool.

Tasks:

- App version command.
- Update check disabled by default.
- Shell completions.
- Better README screenshots/asciinema.
- Release build script.
- Homebrew formula later.

Acceptance:

- Fresh install path documented.
- Binary easy to build and use.

## Immediate Next Tasks

1. Add TUI host details view.
2. Add TUI edit host.
3. Add CLI `show`, `edit`, `delete`.
4. Add JSON output for `list` and `history`.
5. Add migration version table.


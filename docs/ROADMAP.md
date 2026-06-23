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
- TUI host list, details, add, delete, history, status.
- TUI color theme and Russian keyboard shortcuts.
- SQLite schema:
  - `hosts`
  - `connection_history`
- macOS Terminal.app launcher.
- SSH config parser for basic directives.
- Docs:
  - README
  - project structure
  - changelog
- Release:
  - GitHub Actions release builds
  - install scripts for macOS, Linux, and Windows

## Phase 1: TUI Usability

Goal: TUI usable for daily host browsing.

Tasks:

- Done:
  - Host details view:
    - alias
    - target
    - key path
    - notes
    - last connected
  - Delete host flow with confirmation.
  - Better status screen.
  - Better README screenshots.
  - Russian keyboard shortcuts.
- Remaining:
  - Search/filter inside TUI.
  - Add host form improvements:
    - field validation before save
    - defaults visible
    - cancel per form
  - Edit host flow.
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

- Done:
  - Release build script.
  - Install scripts:
    - macOS/Linux shell installer
    - Windows PowerShell installer
  - Better README screenshots.
- Remaining:
  - App version command.
  - Update check disabled by default.
  - Shell completions.
  - Install script checksum verification.
  - Homebrew formula later.
  - Scoop manifest later.
  - Better README asciinema later.

Acceptance:

- Fresh install path documented.
- Binary easy to build and use.

## Next Release: v0.2.0

Goal: make Pinter useful for daily host management from TUI and scripts.

Scope:

- TUI edit host flow.
- TUI search/filter inside host list.
- CLI `show <alias>`.
- CLI `edit <alias>`.
- CLI `delete <alias>`.
- JSON output for `list` and `history`.
- Migration version table before next schema change.
- Install script checksum verification.

Acceptance:

- User can add, inspect, edit, delete, search, and connect hosts from TUI.
- Script users can inspect and delete hosts from CLI.
- JSON output is stable enough for simple scripts.
- Installer verifies downloaded release artifact.

## Immediate Next Tasks

1. Add TUI edit host.
2. Add TUI search/filter.
3. Add CLI `show`, `edit`, `delete`.
4. Add JSON output for `list` and `history`.
5. Add migration version table before next schema change.
6. Add install script checksum verification.

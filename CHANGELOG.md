# Changelog

## Unreleased

### Added

- Added Windows terminal selection with `PINTER_TERMINAL=auto|wt|pwsh|powershell|cmd`.
- Added Windows Terminal, PowerShell 7, Windows PowerShell, and cmd launch paths.
- Added shell-specific SSH command quoting for Windows launchers.
- Added `IdentityFile` expansion for Windows paths, env vars, and Unicode paths.
- Added `justfile` recipes for build, run, test, and clean.
- Added terminal and SSH config tests for Windows paths.

### Fixed

- Fixed Windows SSH key paths being passed with POSIX quoting.
- Fixed `wt` launch fallback when Windows Terminal is unavailable.

## v0.1.1 - 2026-06-23

### Added

- Added changelog.
- Added install scripts for macOS, Linux, and Windows.

## v0.1.0 - 2026-06-23

First MVP release of Pinter: local SSH keeper with CLI and TUI.

### Added

- CLI commands: `add`, `list`, `connect`, `history`, `import-ssh-config`.
- TUI main menu, host browsing, host details, add host, delete host, import, history, status.
- Colored TUI screens and README screenshots.
- Russian keyboard shortcuts in TUI.
- SQLite local storage for hosts and connection history.
- SSH config import for basic directives.
- macOS Terminal.app launcher.
- Cross-platform Makefile build/run helpers.
- GitHub Actions release workflow for macOS, Linux, and Windows binaries.

### Notes

- No passwords or passphrases stored.
- SSH config `Include` is not supported yet.
- Windows/Linux launch paths are still MVP-level.

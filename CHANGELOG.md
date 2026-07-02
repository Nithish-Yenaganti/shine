# Changelog

All notable changes to `shine` will be documented in this file.

## 0.1.0 - 2026-07-01

### Added

- Terminal-native Markdown preview and docs quality checking for files and stdin.
- Interactive TUI with scrolling, search, heading outline, help panel, and live reload.
- Markdown rendering for headings, paragraphs, inline styles, links, task lists, nested lists, block quotes, callouts, tables, code blocks, and image placeholders.
- Built-in themes: `tomorrow-night`, `github`, `mono`, `catppuccin-latte`, `catppuccin-mocha`, `claude`, `everforest`, `jellybeans`, and `gotham`.
- In-app theme picker with `t`.
- Non-interactive modes: `--print`, `--plain`, `--outline`, and `--check`.
- Docs checks for skipped heading levels, duplicate headings, multiple H1s, broken local links, raw URL link text, empty table headers, and long table cells.
- Shell completions for bash, zsh, fish, and PowerShell.
- `--version` output for release verification.
- Installation documentation for source builds and planned Go, GitHub release, npm, and Homebrew channels.
- GoReleaser configuration for macOS and Linux release archives with checksums.
- npm wrapper package that installs and runs the matching GitHub release binary.

### Changed

- Optimized TUI scrolling by avoiding per-frame row padding, width measurement, and line-count recomputation.
- Avoided unnecessary `--watch` reloads when no file changes have occurred.
- Aligned the release workflow with GoReleaser archive names and checksum assets used by the npm installer.

# Changelog

All notable changes to `shine` will be documented in this file.

## Unreleased

### Added

- Added optional Mermaid diagram previews through Mermaid CLI (`mmdc`) with cached image output and code fallback.

### Changed

- Added a 15-column left TUI padding with the existing right-side content reserve.

## 0.1.1 - 2026-07-03

### Changed

- Added an npm-specific README publishing flow.
- Simplified the GitHub README for a shorter project overview.

## 0.1.0 - 2026-07-01

### Added

- Terminal-native Markdown preview and docs quality checking for files and stdin.
- Interactive TUI with scrolling, search, heading outline, help panel, and live reload.
- Markdown rendering for headings, paragraphs, inline styles, links, task lists, nested lists, block quotes, callouts, tables, code blocks, local image previews in Kitty/Ghostty-compatible terminals, and image placeholders elsewhere.
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

- Changed the default theme to `mono`.
- Resolved local image paths relative to the Markdown source file and kept non-interactive output as text placeholders.
- Added left/right terminal padding to the interactive TUI while keeping wrapped content inside the usable terminal width.
- Preserved terminal image placeholder escapes during search highlighting so inline image previews continue to render.
- Optimized TUI scrolling by avoiding per-frame row padding, width measurement, and line-count recomputation.
- Avoided unnecessary `--watch` reloads when no file changes have occurred.
- Aligned the release workflow with GoReleaser archive names and checksum assets used by the npm installer.

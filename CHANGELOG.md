# Changelog

All notable changes to `shine` will be documented in this file.

## Unreleased

## 0.1.2 - 2026-07-10

### 0.1.2 Added

- Added optional Mermaid diagram previews through Mermaid CLI (`mmdc`) with cached image output and code fallback.
- Added `T` as an alternate theme-picker shortcut.
- Added explicit Bubble Tea mouse support for smoother wheel scrolling.
- Added cached viewport body redraws for repeated TUI renders after mouse scrolling.
- Added npm publication checks for duplicate versions and missing GitHub release assets.

### 0.1.2 Changed

- Changed the TUI layout to use responsive page gutters based on terminal width.
- Changed the default `mono` theme to use a black background, white document text, and highlighted dark code blocks.
- Tuned mouse wheel scrolling to move more lines per wheel tick.
- Updated theme display names, aliases, and README theme documentation.
- Changed README logo references to use `fixtures/LOGO.jpeg`.
- Changed the theme picker to avoid rebuilding the document viewport while the overlay is open.

### 0.1.2 Fixed

- Reduced image-heavy TUI redraw cost by transferring cached PNG files instead of embedding converted images in every frame.
- Filtered ignored mouse events before Bubble Tea redraws while preserving vertical wheel scrolling.
- Fixed the initial viewport rendering blank until the first terminal resize event.
- Fixed completed searches remaining focused and cancelled searches leaving stale highlights or matches.

## 0.1.1 - 2026-07-03

### 0.1.1 Changed

- Added an npm-specific README publishing flow.
- Simplified the GitHub README for a shorter project overview.

## 0.1.0 - 2026-07-01

### 0.1.0 Added

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

### 0.1.0 Changed

- Changed the default theme to `mono`.
- Resolved local image paths relative to the Markdown source file and kept non-interactive output as text placeholders.
- Added left/right terminal padding to the interactive TUI while keeping wrapped content inside the usable terminal width.
- Preserved terminal image placeholder escapes during search highlighting so inline image previews continue to render.
- Optimized TUI scrolling by avoiding per-frame row padding, width measurement, and line-count recomputation.
- Avoided unnecessary `--watch` reloads when no file changes have occurred.
- Aligned the release workflow with GoReleaser archive names and checksum assets used by the npm installer.

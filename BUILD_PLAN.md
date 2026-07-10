# Shine Build Plan

## Product Decision

Build `shine` as a Go + Charm terminal Markdown previewer and docs quality checker that feels visually polished and web-like while preserving the core idea: an owned Markdown render model and a terminal-native review experience.

Core pipeline:

```text
Markdown source
-> Goldmark parser
-> owned Shine block model
-> Lip Gloss layout renderer
-> Bubble Tea TUI / print output
```

## Stack

```text
Go
Bubble Tea      TUI app loop, keyboard handling, resize, live interaction
Bubbles         viewport and search input components
Lip Gloss       spacing, colors, borders, visual hierarchy
Goldmark        Markdown parsing into our own model
Cobra           CLI flags and shell completions
fsnotify        live reload
```

## Requirements

- Binary name: `shine`.
- Product name: `shine`.
- Parse Markdown into an owned block model, not a Glamour/Glow wrapper.
- Support headings, paragraphs, bold, italic, inline code, links, quotes, callouts, task lists, tables, code blocks, dividers, optional Mermaid previews, local image previews in Kitty/Ghostty-compatible TUIs, and text image placeholders elsewhere.
- Keep CLI behavior: file input, stdin input, `--print`, `--plain`, `--outline`, `--check`, `--watch`, themes, and shell completions.
- Make `--check` useful for README, changelog, and release-note review by catching publishing mistakes before docs land.
- Use Bubble Tea for interactive TUI behavior.
- Use Lip Gloss for visual presentation.
- Keep the interactive TUI readable with responsive page gutters while leaving non-interactive output unpadded.
- Support keyboard and mouse scrolling in the interactive TUI.
- Use Bubbles where useful for viewport and search input.
- Verify locally against `README.md` and `fixtures/basic.md`.

## Validation

```sh
go test ./...
go build -o bin/shine ./cmd/shine
./bin/shine --print fixtures/basic.md
./bin/shine --print README.md
./bin/shine fixtures/basic.md
```

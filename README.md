# shine

`shine` is a terminal-native Markdown previewer built with Go and the Charm ecosystem. It parses Markdown into its own render model, lays that model out with Lip Gloss, and shows a cleaner preview than a raw Markdown file.

## Scope

This project focuses only on Markdown preview/rendering:

- Markdown files and stdin
- owned document/block model
- terminal layout and rendering
- tables, syntax-highlighted code blocks, callouts, task lists, nested lists, quotes, images as placeholders
- mixed inline styles for bold, italic, strikethrough, inline code, and links
- interactive scrolling
- highlighted search matches
- heading outline
- agent-friendly plain, outline, and check modes
- in-app help panel
- live reload with `--watch`
- theme preset support

It does not include AI transcript rendering, Codex/Claude integration, a browser preview, a Markdown editor, collaboration, or remote docs hosting.

## Usage

```sh
shine README.md
shine --watch README.md
shine --print README.md
shine --plain README.md
shine --outline README.md
shine --check README.md
shine --theme daylight README.md
shine --theme catppuccin-mocha README.md
shine --show-keys README.md
shine --debug-keys README.md
cat README.md | shine
shine completions zsh > _shine
```

Agent-friendly modes:

```sh
shine --plain README.md      # rendered preview without ANSI styling
shine --outline README.md    # heading outline for quick navigation
shine --check README.md      # Markdown quality warnings; exits 2 when warnings exist
```

Use `--print` or `--plain` when a coding agent should show a preview directly in chat output. Use the full TUI only when a human is running `shine` in their own terminal.

Keyboard controls:

```text
q          quit
j/down     scroll down
k/up       scroll up
d/space    half-page down
u          half-page up
g          top
G          bottom
/          search
n          next search result
N          previous search result
o          heading outline
r          reload file
t          theme picker
h/H/F1     show help panel
?          toggle help panel
```

## Themes

Built-in theme presets are `midnight`, `daylight`, `mono`, `catppuccin-latte`, `catppuccin-mocha`, and `claude`. `midnight` is the default. Inside the TUI, press `t` to open the theme picker, use up/down or `j/k`, then press Enter to apply.

`cappuccino` is accepted as an alias for `catppuccin-latte`, and `mocha` is accepted as an alias for `catppuccin-mocha`.

`claude` uses an Anthropic-inspired warm ivory, ink, and clay palette.

```sh
shine --theme midnight README.md
shine --theme daylight README.md
shine --theme mono README.md
shine --theme catppuccin-latte README.md
shine --theme catppuccin-mocha README.md
shine --theme claude README.md
```

## Development

```sh
go test ./...
go build -o bin/shine ./cmd/shine
./bin/shine fixtures/basic.md
./bin/shine --print fixtures/basic.md
```

# mdview

`mdview` is a terminal-native Markdown previewer. It parses Markdown into its own render model, lays that model out for terminal cells, and shows a cleaner preview than a raw Markdown file.

## Scope

This project focuses only on Markdown preview/rendering:

- Markdown files and stdin
- owned document/block model
- terminal layout and rendering
- tables, code blocks, callouts, task lists, quotes, images as placeholders
- mixed inline styles for bold, italic, inline code, and links
- interactive scrolling
- search
- heading outline
- live reload with `--watch`
- theme/config support

It does not include AI transcript rendering, Codex/Claude integration, a browser preview, a Markdown editor, collaboration, or remote docs hosting.

## Usage

```sh
mdview README.md
mdview --watch README.md
mdview --print README.md
mdview --theme daylight README.md
cat README.md | mdview
mdview completions zsh > _mdview
```

Keyboard controls:

```text
q          quit
j/down     scroll down
k/up       scroll up
g          top
G          bottom
/          search
n          next search result
N          previous search result
o          heading outline
r          reload file
```

## Config

`mdview` looks for config at `~/.config/mdview/config.toml`, or you can pass a path with `--config`.
Built-in theme presets are `midnight`, `daylight`, and `mono`.

```toml
[theme]
name = "midnight"
heading = "#8bd5ff"
body = "#e7e7e7"
muted = "#8a8f98"
border = "#3a3f4b"
code = "#b7f7c8"
inline_code = "#ffb86c"
link = "#7dcfff"
table_header = "#ffffff"
code_line_number = "#6c7280"
quote = "#c6a0ff"
callout_note = "#8bd5ff"
callout_tip = "#7ee787"
callout_warning = "#ffd166"
selection = "#26364d"
match_highlight = "#ffd166"
block_padding = 1
code_line_numbers = true
```

## Development

```sh
cargo fmt
cargo test
cargo run -- fixtures/basic.md
cargo run -- --print fixtures/basic.md
```

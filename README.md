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
- in-app help panel
- live reload with `--watch`
- theme preset support

It does not include AI transcript rendering, Codex/Claude integration, a browser preview, a Markdown editor, collaboration, or remote docs hosting.

## Usage

```sh
shine README.md
shine --watch README.md
shine --print README.md
shine --theme daylight README.md
cat README.md | shine
shine completions zsh > _shine
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
?          help panel
```

## Themes

Built-in theme presets are `midnight`, `daylight`, and `mono`.

```sh
shine --theme midnight README.md
shine --theme daylight README.md
shine --theme mono README.md
```

## Development

```sh
go test ./...
go build -o bin/shine ./cmd/shine
./bin/shine fixtures/basic.md
./bin/shine --print fixtures/basic.md
```

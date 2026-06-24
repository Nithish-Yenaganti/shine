```ascii
         █████       ███
        ░░███       ░░░
  █████  ░███████   ████  ████████    ██████
 ███░░   ░███░░███ ░░███ ░░███░░███  ███░░███
░░█████  ░███ ░███  ░███  ░███ ░███ ░███████
 ░░░░███ ░███ ░███  ░███  ░███ ░███ ░███░░░
 ██████  ████ █████ █████ ████ █████░░██████
░░░░░░  ░░░░ ░░░░░ ░░░░░ ░░░░ ░░░░░  ░░░░░░
```

# shine

`shine` is a terminal-native Markdown previewer built with Go and the Charm ecosystem. It turns Markdown files into a readable TUI preview with themes, tables, callouts, code blocks, search, and agent-friendly output modes.

> Release status: `shine` is not published yet. Current local version is `0.1.0-dev`.

## Features

- Markdown files and stdin input
- Owned document/block render model
- Terminal layout and rendering with Lip Gloss
- Headings, paragraphs, block quotes, callouts, task lists, nested lists, tables, links, inline styles, code blocks, and image placeholders
- Interactive scrolling, search, and heading outline
- Live preview reload with `--watch`
- In-app help panel and theme picker
- Built-in themes: `midnight`, `daylight`, `mono`, `catppuccin-latte`, `catppuccin-mocha`, `claude`, `everforest`, `jellybeans`, and `gotham`
- Agent-friendly modes: `--plain`, `--outline`, and `--check`
- Shell completions for bash, zsh, fish, and PowerShell

## Non-Goals

`shine` is focused only on Markdown preview/rendering. It does not include AI transcript rendering, Codex/Claude integration, browser preview, Markdown editing, collaboration, or remote docs hosting.

## Local Usage

Build the local binary:

```sh
go build -o bin/shine ./cmd/shine
```

Run the TUI preview:

```sh
bin/shine README.md
bin/shine --watch README.md
cat README.md | bin/shine
```

Print a rendered preview and exit:

```sh
bin/shine --print README.md
bin/shine --plain README.md
```

Inspect a Markdown file for agents or scripts:

```sh
bin/shine --outline README.md
bin/shine --check README.md
```

Check the version:

```sh
bin/shine version
bin/shine --version
```

Generate shell completions:

```sh
bin/shine completions zsh > _shine
```

## Agent-Friendly Modes

Use `--print` or `--plain` when a coding agent should show a Markdown preview directly in chat output. Use the full TUI only when a human is running `shine` in their own terminal.

```sh
bin/shine --plain README.md      # rendered preview without ANSI styling
bin/shine --outline README.md    # heading outline for quick navigation
bin/shine --check README.md      # Markdown quality warnings; exits 2 when warnings exist
```

## Keyboard Controls

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

`midnight` is the default theme. Inside the TUI, press `t` to open the theme picker, use up/down or `j/k`, then press Enter to apply.

```sh
bin/shine --theme midnight README.md
bin/shine --theme daylight README.md
bin/shine --theme mono README.md
bin/shine --theme catppuccin-latte README.md
bin/shine --theme catppuccin-mocha README.md
bin/shine --theme claude README.md
bin/shine --theme everforest README.md
bin/shine --theme jellybeans README.md
bin/shine --theme gotham README.md
```

Aliases:

- `cappuccino` resolves to `catppuccin-latte`
- `mocha` resolves to `catppuccin-mocha`

The `claude` theme uses an Anthropic-inspired warm ivory, ink, and clay palette.
The `everforest`, `jellybeans`, and `gotham` themes are dark palettes based on their original editor themes.

## Development

```sh
go test ./...
go vet ./...
go build -o bin/shine ./cmd/shine
bin/shine fixtures/basic.md
bin/shine --print fixtures/basic.md
```

## Release Notes

See [CHANGELOG.md](CHANGELOG.md).

## License

MIT. See [LICENSE](LICENSE).

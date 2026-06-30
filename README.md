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

`shine` is a terminal Markdown previewer and docs quality checker for README, changelog, and release-note workflows. It renders Markdown in a TUI with themes, tables, callouts, code blocks, search, live reload, and checks that catch common publishing mistakes before they land.

Status: unreleased. Current local version: `0.1.0-dev`.

## Problem

Markdown is easy to write but awkward to review raw inside terminal-first workflows. `shine` gives developers a fast rendered view without leaving the terminal, then checks whether the document is ready to publish: heading structure, missing alt text, broken local links and images, duplicate headings, and hard-to-scan tables.

## Features

- Markdown files and stdin
- TUI preview with scrolling, search, heading outline, help panel, and theme picker
- Tables, callouts, task lists, nested lists, quotes, links, inline styles, code blocks, and image placeholders
- Docs checks for heading structure, duplicate headings, image alt text, local links/images, and long table cells
- Live reload with `--watch`
- Responsive scrolling for larger Markdown files
- Non-interactive commands: `--print`, `--plain`, `--outline`, `--check`
- Shell completions for bash, zsh, fish, and PowerShell

## Install

Until the first release is published, build from source:

```sh
git clone https://github.com/Nithish-Yenaganti/shine.git
cd shine
go build -o bin/shine ./cmd/shine
bin/shine version
```

After the first tagged release:

```sh
go install github.com/Nithish-Yenaganti/shine/cmd/shine@latest
npm install -g @nithish-yenaganti/shine
```

Homebrew support is planned but not configured yet.

## Usage

```sh
bin/shine README.md
bin/shine --watch README.md
cat README.md | bin/shine
```

Non-interactive output:

```sh
bin/shine --print README.md
bin/shine --plain README.md
bin/shine --outline README.md
bin/shine --check README.md
```

## Docs Review

Use `--check` before publishing README files, changelogs, release notes, or docs pages:

```sh
bin/shine --check README.md
```

`shine` reports issues that are easy to miss in raw Markdown:

- missing or misplaced H1 headings
- skipped heading levels
- duplicate heading text
- missing image alt text
- broken local image and link targets
- raw URL link text
- uneven or hard-to-scan tables

Other commands:

```sh
bin/shine version
bin/shine --version
bin/shine completions zsh > _shine
```

## Keyboard

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

`midnight` is the default. Press `t` in the TUI to change themes.

Available themes:

- `midnight`
- `daylight`
- `mono`
- `catppuccin-latte`
- `catppuccin-mocha`
- `claude`
- `everforest`
- `jellybeans`
- `gotham`

Aliases:

- `cappuccino` -> `catppuccin-latte`
- `mocha` -> `catppuccin-mocha`

## Release

Tagged releases are built by the GitHub release workflow through GoReleaser. To verify the release configuration locally:

```sh
goreleaser check
goreleaser release --snapshot --clean
```

Archive format:

```text
shine_<version>_<os>_<arch>.tar.gz
```

Publish after tagging:

```sh
git tag v0.1.0
git push origin v0.1.0
npm publish --access public
```

The npm package is a thin wrapper that downloads the matching GitHub release binary during `postinstall`.

## Development

```sh
go test ./...
go vet ./...
go build -o bin/shine ./cmd/shine
npm run test:npm
```

## License

MIT. See [LICENSE](LICENSE).

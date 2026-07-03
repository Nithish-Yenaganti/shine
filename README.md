![Shine logo](fixtures/LOGO.png)

# shine

`shine` is a terminal Markdown previewer and docs checker for README, changelog, and release-note workflows.

Preview Markdown without leaving the terminal, then run quick checks for common publishing issues before your docs land.

Current version: `0.1.0`.

## Features

- Markdown files, stdin, and non-interactive output
- TUI preview with scrolling, search, outline, help, and themes
- Tables, callouts, task lists, code blocks, links, and inline styles
- Local image previews in Kitty/Ghostty-compatible terminals
- Docs checks for headings, duplicate titles, image alt text, links, images, and table readability
- Non-interactive commands: `--print`, `--plain`, `--outline`, `--check`
- Shell completions for bash, zsh, fish, and PowerShell

## Install

```sh
go install github.com/Nithish-Yenaganti/shine/cmd/shine@latest
npm install -g @nk02/shine
```

Build from source:

```sh
git clone https://github.com/Nithish-Yenaganti/shine.git
cd shine
go build -o bin/shine ./cmd/shine
bin/shine version
```

## Usage

```sh
shine README.md
shine --watch README.md
cat README.md | shine
shine --print README.md
shine --plain README.md
shine --outline README.md
shine --check README.md
```

## Image Previews

Local images render inline in the interactive TUI on Kitty-compatible terminals, currently Kitty and Ghostty. Image paths resolve relative to the Markdown file, so `![Logo](fixtures/LOGO.png)` works from `README.md`.

Unsupported terminals, including Apple's default macOS Terminal.app, show a text placeholder instead. `--print`, `--plain`, remote images, and missing files also use placeholders.

## Docs Review

Use `--check` before publishing docs:

```sh
shine --check README.md
```

Checks include:

- heading structure
- duplicate headings
- missing image alt text
- broken local links and images
- raw URL link text
- hard-to-scan tables

Example output:

```text
3 markdown warning(s):
- heading "Install" jumps from H1 to H3
- block 5 link file not found: ./missing.md
- block 7 table row 2 column 3 is very long
```

Other commands:

```sh
shine version
shine --version
shine completions zsh > _shine
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

`mono` is the default. Press `t` in the TUI to switch themes.

- `tomorrow-night`
- `github`
- `mono`
- `catppuccin-latte`
- `catppuccin-mocha`
- `claude`
- `everforest`
- `jellybeans`
- `gotham`

Aliases: `cappuccino`, `daylight`, `midnight`, `mocha`.

## Release

Tagged releases are built by GitHub Actions through GoReleaser.

```sh
goreleaser check
goreleaser release --snapshot --clean
```

Publish:

```sh
git tag v0.1.0
git push origin v0.1.0
npm run publish:npm -- --access public
```

Checklist:

- `go test ./...`
- `go vet ./...`
- `go build -o bin/shine ./cmd/shine`
- `npm run test:npm`
- `bin/shine --check README.md`
- `goreleaser check`

## Development

```sh
go test ./...
go vet ./...
go build -o bin/shine ./cmd/shine
npm run test:npm
```

## Community

- Read [CONTRIBUTING.md](CONTRIBUTING.md) before opening a pull request.
- Use [SUPPORT.md](SUPPORT.md) for help and issue guidance.
- Report vulnerabilities through [SECURITY.md](SECURITY.md).
- Follow the project [Code of Conduct](CODE_OF_CONDUCT.md).

## License

MIT. See [LICENSE](LICENSE).

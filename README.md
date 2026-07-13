 ![Shine logo](fixtures/demo/LOGO.jpeg)

# shine

`shine` is a terminal Markdown previewer and docs checker for README, changelog, and release-note workflows.

Preview Markdown without leaving the terminal, then run quick checks for common publishing issues before your docs land.

Current version: `0.1.2`.

## Demo

![Demo video](fictures/demo/demo.mov)

## Features

- Markdown files, stdin, and non-interactive output
- TUI preview with keyboard and mouse scrolling, search, outline, help, themes, responsive page gutters, and optimized redraws
- Tables, callouts, task lists, code blocks, links, and inline styles
- Local image previews in Kitty/Ghostty-compatible terminals
- Optional Mermaid previews through Mermaid CLI (`mmdc`)
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
# Open README.md in the interactive preview
shine README.md

# Preview README.md and reload when it changes
shine --watch README.md

# Preview Markdown from stdin
cat README.md | shine

# Print the styled Markdown output once
shine --print README.md

# Print plain text output without styling
shine --plain README.md

# Show the document heading outline
shine --outline README.md

# Check README.md for common docs issues
shine --check README.md
```

## Image Previews

Local images render inline in the interactive TUI on Kitty-compatible terminals, currently Kitty and Ghostty. Image paths resolve relative to the Markdown file, so `![Logo](fixtures/LOGO.jpeg)` works from `README.md`.

JPEG and GIF previews are cached as local PNG files to keep image-heavy scrolling and theme changes responsive.

Unsupported terminals, including Apple's default macOS Terminal.app, show a text placeholder instead. `--print`, `--plain`, remote images, and missing files also use placeholders.

Mermaid code blocks can render as inline diagrams when `mmdc` from Mermaid CLI is installed. Without `mmdc`, or outside supported image terminals, Mermaid blocks stay readable as code with a short fallback note.

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
mouse      scroll
d/space    half-page down
u          half-page up
g          top
G          bottom
/          search
n          next search result
N          previous search result
o          heading outline
r          reload file
t/T        theme picker
h/H/F1     show help panel
?          toggle help panel
```

Mouse wheel scrolling is tuned for terminal use. Non-wheel mouse input is filtered before redraws, and overlays such as help, outline, search, and the theme picker block document scrolling behind them.

## Themes

`mono` is the default black-background theme. Press `t` or `T` in the TUI to switch themes.

- Tomorrow Night: `tomorrow-night`
- GitHub Light: `github`
- Mono: `mono`
- Catppuccin Latte: `catppuccin-latte`
- Catppuccin Mocha: `catppuccin-mocha`
- Claude: `claude`
- Everforest Dark: `everforest`
- Jellybeans: `jellybeans`
- Gotham: `gotham`

Aliases: `daylight`, `latte`, `cappuccino`, `mocha`, `midnight`.

## Release

Tagged releases are built by GitHub Actions through GoReleaser.

```sh
goreleaser check
goreleaser release --snapshot --clean
```

Publish:

```sh
git tag v0.1.2
git push origin v0.1.2
# Wait for the GitHub release workflow and all release assets
npm run publish:npm -- --access public
```

The npm publish command checks that `0.1.2` is still available and that the published GitHub release contains every required asset.

Checklist:

- `go test ./...`
- `go vet ./...`
- `go build -o bin/shine ./cmd/shine`
- `npm run test:npm`
- `npm run test:publish`
- `bin/shine --check README.md`
- `goreleaser check`

## Development

```sh
go test ./...
go vet ./...
go build -o bin/shine ./cmd/shine
npm run test:npm
npm run test:publish
```

## Community

- Read [CONTRIBUTING.md](CONTRIBUTING.md) before opening a pull request.
- Use [SUPPORT.md](SUPPORT.md) for help and issue guidance.
- Report vulnerabilities through [SECURITY.md](SECURITY.md).
- Follow the project [Code of Conduct](CODE_OF_CONDUCT.md).

## License

MIT. See [LICENSE](LICENSE).

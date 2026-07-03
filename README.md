![Shine logo](fixtures/LOGO.png)

# shine

`shine` is a terminal Markdown previewer and docs quality checker for README, changelog, and release-note workflows. It renders Markdown in a TUI with themes, tables, callouts, code blocks, search, live reload, and checks that catch common publishing mistakes before they land.

Current version: `0.1.0`.

## Problem

Markdown is easy to write but awkward to review raw inside terminal-first workflows. `shine` gives developers a fast rendered view without leaving the terminal, then checks whether the document is ready to publish: heading structure, missing alt text, broken local links and images, duplicate headings, and hard-to-scan tables.

## Features

- Markdown files and stdin
- TUI preview with scrolling, search, heading outline, help panel, theme picker, and right-side terminal padding
- Tables, callouts, task lists, nested lists, quotes, links, inline styles, code blocks, and local image previews in Kitty/Ghostty-compatible terminals
- Text image placeholders for unsupported terminals, remote images, missing files, and non-interactive output
- Docs checks for heading structure, duplicate headings, image alt text, local links/images, and long table cells
- Live reload with `--watch`
- Responsive scrolling for larger Markdown files
- Non-interactive commands: `--print`, `--plain`, `--outline`, `--check`
- Shell completions for bash, zsh, fish, and PowerShell

## Why Shine?

Most Markdown tools answer one question: "what does this look like?" `shine` answers that and the next one: "is this ready to publish?"

Developers already have Markdown viewers, but many still want a faster terminal-native way to read, preview, and verify docs without jumping into an editor or browser.

Use it when you want a terminal-native review pass for README files, changelogs, release notes, or docs pages before they reach GitHub, npm, or a release.

That matters because project docs are often the first thing users see. Broken headings, missing image alt text, dead local links, and hard-to-read tables can make a solid project feel unfinished. `shine` keeps the review loop close to where developers already work: preview the Markdown, inspect the outline, and catch publish-blocking issues without opening a browser or editor preview.

In practice, `shine` helps you:

- preview Markdown without leaving the terminal
- catch docs problems before publishing
- review README, changelog, and release-note changes faster
- use `--check` in CI to stop broken docs from landing
- keep terminal-first workflows lightweight

## Install

Install the latest release with Go or npm:

```sh
go install github.com/Nithish-Yenaganti/shine/cmd/shine@latest
npm install -g @nk02/shine
```

Or build from source:

```sh
git clone https://github.com/Nithish-Yenaganti/shine.git
cd shine
go build -o bin/shine ./cmd/shine
bin/shine version
```

Homebrew support is planned but not configured yet.

## Usage

```sh
shine README.md
shine --watch README.md
cat README.md | shine
```

When building from source in this repository, use `bin/shine`:

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

## Image Previews

In the interactive TUI, local Markdown images render inline when the terminal supports Kitty-compatible graphics. This currently targets Kitty and Ghostty. Image paths are resolved relative to the Markdown file, so `![Logo](fixtures/LOGO.png)` works from `README.md`.

Note: inline image rendering only works in terminals with Kitty-compatible image support, such as Kitty and Ghostty. It does not work in Apple's default macOS Terminal.app.

`--print`, `--plain`, unsupported terminals, remote images, and missing files keep a text placeholder instead of emitting graphics escapes.

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

Example output:

```text
3 markdown warning(s):
- heading "Install" jumps from H1 to H3
- block 5 link file not found: ./missing.md
- block 7 table row 2 column 3 is very long
```

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

`mono` is the default. Press `t` in the TUI to change themes.

Available themes:

- `tomorrow-night`
- `github`
- `mono`
- `catppuccin-latte`
- `catppuccin-mocha`
- `claude`
- `everforest`
- `jellybeans`
- `gotham`

Aliases:

- `cappuccino` -> `catppuccin-latte`
- `daylight` -> `github`
- `midnight` -> `tomorrow-night`
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

Release checklist:

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

# shine

Terminal Markdown preview and docs checks from your shell.

```sh
npm install -g @nk02/shine
shine README.md
```

## What It Does

- renders Markdown in an interactive terminal preview
- checks README, changelog, and release-note files before publishing
- supports `--print`, `--plain`, `--outline`, and `--check`
- includes themes, search, scrolling, outline view, and live reload
- optionally renders Mermaid diagrams when Mermaid CLI (`mmdc`) is installed

## Image Support

Inline image previews work only in Kitty-compatible terminals, currently Kitty and Ghostty.

Apple's default macOS Terminal.app and unsupported terminals show text placeholders instead. `--print`, `--plain`, remote images, and missing files also use placeholders.

Mermaid diagrams use the same image path. If `mmdc` is unavailable, Mermaid blocks remain visible as code.

## Commands

```sh
shine README.md
shine --watch README.md
shine --check README.md
shine --print README.md
shine --outline README.md
```

## Notes

The npm package downloads the matching `shine` binary from the GitHub release during install.

Supported platforms: macOS and Linux on x64 or arm64.

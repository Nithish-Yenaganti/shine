# shine

Terminal Markdown preview and docs checks from your shell.

Current version: `0.1.2`.

```sh
# Install the CLI globally from npm
npm install -g @nk02/shine

# Open README.md in the interactive preview
shine README.md
```

## What It Does

- renders Markdown in an interactive terminal preview
- checks README, changelog, and release-note files before publishing
- supports `--print`, `--plain`, `--outline`, and `--check`
- includes themes, responsive page gutters, keyboard and mouse scrolling, outline view, live reload, and optimized redraws
- optionally renders Mermaid diagrams when Mermaid CLI (`mmdc`) is installed

## Image Support

Inline image previews work only in Kitty-compatible terminals, currently Kitty and Ghostty.

JPEG and GIF previews use cached local PNG files for responsive scrolling.

Apple's default macOS Terminal.app and unsupported terminals show text placeholders instead. `--print`, `--plain`, remote images, and missing files also use placeholders.

Mermaid diagrams use the same image path. If `mmdc` is unavailable, Mermaid blocks remain visible as code.

## Terminal Controls

Use `j/k`, arrow keys, page keys, `g/G`, or the mouse wheel to scroll. Use `/` to search, `o` for the heading outline, and `t` or `T` for the theme picker.

Mouse motion and clicks that are not wheel events are ignored. Help, outline, search, and theme overlays block document scrolling behind them.

## Commands

```sh
# Open README.md in the interactive preview
shine README.md

# Preview README.md and reload when it changes
shine --watch README.md

# Check README.md for common docs issues
shine --check README.md

# Print styled Markdown once
shine --print README.md

# Show the document heading outline
shine --outline README.md
```

## Notes

The npm package downloads the matching `shine` binary from the GitHub release during install.

Supported platforms: macOS and Linux on x64 or arm64.

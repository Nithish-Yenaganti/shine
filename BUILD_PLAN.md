# Terminal Markdown Previewer Build Plan

## Product Decision

Build a Rust terminal Markdown previewer that owns its Markdown render model.

This project is focused only on Markdown preview/rendering. It is not an AI transcript renderer, Codex/Claude integration, browser renderer, Markdown editor, or docs platform.

Core pipeline:

```text
Markdown source
-> parser
-> our block model
-> layout engine
-> terminal renderer
-> TUI preview
```

## Architecture Direction

Recommended stack:

```text
Rust
pulldown-cmark   Markdown parsing
ratatui          terminal UI layout
crossterm        terminal events and rendering backend
syntect          syntax highlighting
notify           live reload
clap             CLI arguments
serde/toml       config and themes
insta            snapshot tests
```

Important technical decision: the app owns its internal Markdown render model instead of directly piping Markdown into an existing renderer. That keeps control over spacing, wrapping, tables, callouts, themes, search, folding, and future layout behavior.

Example internal model:

```text
Document
  Heading
  Paragraph
  CodeBlock
  Table
  Quote
  Callout
  List
  TaskListItem
  Image
  Link
  Divider
```

## Phase 1: Proof Renderer

### Goal

Prove we can parse Markdown into our own model and render a readable terminal preview.

### Tasks

- Create Rust project.
- Add CLI command: `mdview README.md`.
- Parse Markdown with `pulldown-cmark`.
- Convert parsed events into our own block model.
- Render basic blocks:
  - Heading
  - Paragraph
  - Code block
  - Quote
  - List
  - Divider
- Print static terminal output.

### Dependencies

```text
clap
pulldown-cmark
textwrap
```

### Validation Check

A real `README.md` renders clearly in the terminal without raw Markdown syntax showing everywhere.

### Tests

- Parser unit tests: Markdown source to block model.
- Snapshot tests: block model to rendered text.
- Golden sample: `fixtures/basic.md`.

### Risks Carried Forward

Tables, nested lists, and syntax highlighting are not solved yet.

### Rollback Or Course Correction

If `pulldown-cmark` makes block modeling painful, test `markdown-rs` before building more.

## Phase 2: Real TUI Preview

### Goal

Turn static output into an interactive previewer.

### Tasks

- Add `ratatui` and `crossterm`.
- Create a scrollable viewport.
- Handle terminal resize.
- Add keyboard navigation:
  - `q`: quit
  - `j` / down: scroll down
  - `k` / up: scroll up
  - `g`: top
  - `G`: bottom
- Render document inside TUI.
- Preserve our render model as the source of truth.

### Dependencies

```text
ratatui
crossterm
```

### Validation Check

User can open a long Markdown file, scroll smoothly, resize the terminal, and still read it cleanly.

### Tests

- Layout tests for viewport width and height.
- Renderer tests for wrapping and scroll boundaries.
- Manual test in at least two terminals.

### Risks Carried Forward

Terminal rendering may look different across terminals.

### Rollback Or Course Correction

If `ratatui` fights the custom document layout too much, keep `crossterm` and build a simpler custom viewport renderer.

## Phase 3: Web-Like Markdown Blocks

### Goal

Make the preview feel meaningfully better than a styled text dump.

### Tasks

- Improve spacing and hierarchy.
- Add theme tokens:
  - heading colors
  - body color
  - muted color
  - border color
  - code background
  - quote/callout accent
- Add table rendering.
- Add task checkboxes.
- Add links with visible labels.
- Add callout detection:
  - `> [!NOTE]`
  - `> [!WARNING]`
  - `> [!TIP]`
- Improve list indentation.

### Dependencies

```text
unicode-width
serde
toml
```

### Validation Check

A Markdown document with headings, lists, code, tables, quotes, and callouts feels like a designed preview.

### Tests

- Fixture tests for each block type.
- Snapshot tests for narrow and wide terminal widths.
- Theme token tests.

### Risks Carried Forward

Tables and wrapping can become complex, especially on narrow terminals.

### Rollback Or Course Correction

If tables delay progress, ship simple readable tables first and defer smart column resizing.

## Phase 4: Live Preview

### Goal

Make it useful while editing Markdown.

### Tasks

- Add `--watch` mode.
- Reload file on save.
- Preserve scroll position when possible.
- Show parse/render errors without crashing.
- Add status bar:
  - filename
  - modified time
  - line/block count
  - watch status

### Dependencies

```text
notify
```

### Validation Check

User edits a Markdown file in their editor and sees the terminal preview update immediately.

### Tests

- File watcher integration test.
- Reload behavior test.
- Invalid Markdown/input resilience test.
- Manual editor workflow test.

### Risks Carried Forward

File watch behavior differs by OS.

### Rollback Or Course Correction

If live watch is flaky, add a manual reload key first and keep watch behind `--watch`.

## Phase 5: Search And Navigation

### Goal

Make large Markdown files practical.

### Tasks

- Add search.
- Highlight matches.
- Add heading outline.
- Jump between headings.
- Add block-level anchors internally.
- Add keyboard shortcuts:
  - `/`: search
  - `n`: next match
  - `N`: previous match
  - `o`: outline

### Dependencies

No major new dependency required.

### Validation Check

User can navigate a long README or docs page without endless scrolling.

### Tests

- Search index tests.
- Heading extraction tests.
- Match navigation tests.

### Risks Carried Forward

Search highlighting can conflict with syntax highlighting and theme colors.

### Rollback Or Course Correction

If outline UI is too much, ship heading jump first.

## Phase 6: Polish And Distribution

### Goal

Make it installable and credible.

### Tasks

- Choose final binary name.
- Add config file support.
- Add theme presets.
- Add shell completions.
- Add README with screenshots or GIF.
- Set up CI.
- Build macOS, Linux, and Windows binaries.
- Create GitHub release flow.
- Add Homebrew formula later.

### Dependencies

```text
serde
toml
insta
GitHub Actions
```

### Validation Check

A developer can install it, run it on a Markdown file, and understand the product in under one minute.

### Tests

- CI test suite.
- Cross-platform build check.
- Snapshot render tests.
- Manual terminal compatibility pass.

### Risks Carried Forward

Windows terminal behavior may need dedicated fixes.

### Rollback Or Course Correction

If cross-platform release slows launch, ship macOS/Linux first and mark Windows experimental.

## Deferred Work

Do not build these until the Markdown previewer is strong:

- AI transcript rendering
- Codex/Claude integration
- Browser/web renderer
- Markdown editor
- Plugin system
- Remote docs hosting
- Collaboration
- Custom Markdown parser from scratch
- Image rendering beyond simple terminal-supported experiments
- Diagram rendering

## Readiness Verdict

Ready to build.

Start with Phase 1 only: Rust project, CLI input, parser, owned block model, and static terminal renderer. If that foundation feels clean, move to the TUI.

# Contributing to shine

Thanks for helping improve `shine`. This project is a terminal Markdown previewer and docs quality checker, so the best contributions either improve terminal rendering, make docs checks more useful, reduce release risk, or clarify the user workflow.

## Good First Contributions

- Improve Markdown rendering for a common syntax case.
- Add or refine themes; new theme contributions are welcome.
- Add or refine `--check` warnings for README, changelog, and release-note workflows.
- Improve tests around parser, renderer, CLI, or TUI behavior.
- Clarify README examples or release documentation.
- Report a Markdown file that renders poorly, ideally with a small reproducible fixture.

## Development Setup

```sh
git clone https://github.com/Nithish-Yenaganti/shine.git
cd shine
go test ./...
go vet ./...
go build -o bin/shine ./cmd/shine
npm run test:npm
```

If `npm run test:npm` fails because `bin/shine` does not exist, build the Go binary first.

## Before Opening a Pull Request

Run:

```sh
test -z "$(gofmt -l cmd internal)"
go test ./...
go vet ./...
go build -o bin/shine ./cmd/shine
npm run test:npm
bin/shine --check README.md
```

## Pull Request Guidelines

- Keep changes focused. Prefer small PRs over broad rewrites.
- Add tests for behavior changes.
- Update README or CHANGELOG when user-facing behavior changes.
- For theme changes, include the display name, aliases if needed, docs updates, and palette tests.
- For TUI interaction changes, cover keyboard, mouse wheel, overlay blocking, and redraw/cache invalidation behavior where relevant.
- Avoid unrelated formatting or dependency churn.
- Explain why the change matters for terminal Markdown review.

## Reporting Bugs

Please include:

- operating system and terminal
- `shine version`
- command you ran
- minimal Markdown input that reproduces the issue
- expected output versus actual output

## Proposing Features

Feature requests are strongest when they include:

- the Markdown workflow being improved
- why existing behavior is insufficient
- a sample input file or command
- whether the feature belongs in preview, `--check`, or both

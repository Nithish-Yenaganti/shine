package render

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"

	"github.com/Nithish-Yenaganti/shine/internal/config"
	"github.com/Nithish-Yenaganti/shine/internal/parser"
)

func TestRenderCalloutTableAndCode(t *testing.T) {
	doc, err := parser.Parse([]byte(`# shine

> [!WARNING]
> Pay attention.

| Feature | Description |
| --- | --- |
| Tables | wrap long content into readable terminal rows |

`+"```go\nfmt.Println(\"hi\")\n```\n"+`
`), "test.md")
	if err != nil {
		t.Fatal(err)
	}
	out := New(48, config.ThemeByName("mono")).Render(doc)
	if !strings.Contains(out, "┃ WARNING") {
		t.Fatalf("missing callout accent:\n%s", out)
	}
	if !strings.Contains(out, "┌") || !strings.Contains(out, "┼") {
		t.Fatalf("missing table borders:\n%s", out)
	}
	if !strings.Contains(out, "┌ go") || !strings.Contains(out, "fmt.Println") {
		t.Fatalf("missing code block:\n%s", out)
	}
}

func TestRenderWidthConstrainedTable(t *testing.T) {
	doc, err := parser.Parse([]byte(`| Feature | Description |
| --- | --- |
| Tables | wrap long content into readable terminal rows |
`), "test.md")
	if err != nil {
		t.Fatal(err)
	}
	out := New(36, config.ThemeByName("mono")).Render(doc)
	for _, line := range strings.Split(out, "\n") {
		if lipgloss.Width(line) > 80 {
			t.Fatalf("line too wide (%d): %q", lipgloss.Width(line), line)
		}
	}
	if !strings.Contains(out, "readable") {
		t.Fatalf("expected wrapped content:\n%s", out)
	}
}

func TestRenderNestedList(t *testing.T) {
	doc, err := parser.Parse([]byte("- Parent item wraps into a continuation line when the terminal is narrow\n  - Child item\n"), "test.md")
	if err != nil {
		t.Fatal(err)
	}
	out := New(34, config.ThemeByName("mono")).Render(doc)
	if !strings.Contains(out, "• Parent item") || !strings.Contains(out, "  • Child item") {
		t.Fatalf("missing nested list layout:\n%s", out)
	}
	if strings.Contains(out, "Parent item Child item") {
		t.Fatalf("nested text was flattened into parent:\n%s", out)
	}
}

func TestAsciiCodeBlockDoesNotShowLineNumbers(t *testing.T) {
	doc, err := parser.Parse([]byte("```ascii\n █████\n░░███\n```\n"), "test.md")
	if err != nil {
		t.Fatal(err)
	}
	out := New(48, config.ThemeByName("mono")).Render(doc)
	if strings.Contains(out, "│ 1 ") || strings.Contains(out, "│ 2 ") {
		t.Fatalf("ascii block should not show line numbers:\n%s", out)
	}
	if !strings.Contains(out, "█████") || !strings.Contains(out, "░░███") {
		t.Fatalf("missing ascii art:\n%s", out)
	}
}

func TestDaylightCodeUsesReadableRawLines(t *testing.T) {
	r := New(48, config.ThemeByName("daylight"))
	lines := r.codeLines("fn main() {\n    fmt.Println(\"hi\")\n}", "go")
	got := strings.Join(lines, "\n")
	if strings.Contains(got, "\x1b[") {
		t.Fatalf("daylight code should not contain nested ANSI highlighting: %q", got)
	}
	if !strings.Contains(got, "fmt.Println") {
		t.Fatalf("missing code content: %q", got)
	}
}

func TestClaudeCodeUsesThemeReadableRawLines(t *testing.T) {
	r := New(48, config.ThemeByName("claude"))
	lines := r.codeLines("fn main() {\n    fmt.Println(\"hi\")\n}", "go")
	got := strings.Join(lines, "\n")
	if strings.Contains(got, "\x1b[") {
		t.Fatalf("claude code should not contain nested ANSI highlighting: %q", got)
	}
	if !strings.Contains(got, "fmt.Println") {
		t.Fatalf("missing code content: %q", got)
	}
}

func TestDaylightDoesNotPaintTokenBackgrounds(t *testing.T) {
	r := New(48, config.ThemeByName("daylight"))
	bg := r.style().GetBackground()
	if _, ok := bg.(lipgloss.NoColor); !ok {
		t.Fatalf("daylight renderer should not paint token backgrounds, got %v", bg)
	}
}

func TestInlineRelativePathKeepsSpace(t *testing.T) {
	doc, err := parser.Parse([]byte("- `go test ./...`\n- `go build ./cmd/shine`\n"), "test.md")
	if err != nil {
		t.Fatal(err)
	}
	out := New(64, config.ThemeByName("mono")).Render(doc)
	if strings.Contains(out, "test./...") || strings.Contains(out, "build./cmd") {
		t.Fatalf("relative paths should not stick to previous words:\n%s", out)
	}
	if !strings.Contains(out, "go test ./...") || !strings.Contains(out, "go build ./cmd/shine") {
		t.Fatalf("missing spaced relative path commands:\n%s", out)
	}
}

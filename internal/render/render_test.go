package render

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"

	"shine/internal/config"
	"shine/internal/parser"
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

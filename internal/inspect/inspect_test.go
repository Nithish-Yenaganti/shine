package inspect

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Nithish-Yenaganti/shine/internal/parser"
)

func TestOutline(t *testing.T) {
	doc, err := parser.Parse([]byte("# Title\n\n## Section\n"), "test.md")
	if err != nil {
		t.Fatal(err)
	}
	got := Outline(doc)
	if got != "# Title\n## Section\n" {
		t.Fatalf("outline = %q", got)
	}
}

func TestCheckWarnings(t *testing.T) {
	doc, err := parser.Parse([]byte("plain text\n\n![missing](./missing.png)\n"), "test.md")
	if err != nil {
		t.Fatal(err)
	}
	warnings := Check(doc, "")
	if len(warnings) == 0 {
		t.Fatal("expected warnings")
	}
	out := FormatWarnings(warnings)
	if !strings.Contains(out, "document has no headings") {
		t.Fatalf("missing heading warning:\n%s", out)
	}
}

func TestCheckDocsQualityWarnings(t *testing.T) {
	dir := t.TempDir()
	sourcePath := filepath.Join(dir, "README.md")
	if err := os.WriteFile(sourcePath, []byte("placeholder"), 0o644); err != nil {
		t.Fatal(err)
	}
	doc, err := parser.Parse([]byte(`# Title

### Skipped

## Skipped

# Another Title

## Skipped

[https://example.com](https://example.com)

[Missing local](./missing.md)

| Feature | Description |
| --- | --- |
| Check | This value is intentionally long enough to trigger a table readability warning because it would be hard to scan in a terminal preview |
`), sourcePath)
	if err != nil {
		t.Fatal(err)
	}

	out := FormatWarnings(Check(doc, sourcePath))
	for _, want := range []string{
		`heading "Skipped" jumps from H1 to H3`,
		`duplicate heading "Skipped"`,
		"document has multiple H1 headings",
		"link text is a raw URL",
		"link file not found: ./missing.md",
		"table row 1 column 2 is very long",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing warning %q:\n%s", want, out)
		}
	}
}

func TestStripANSI(t *testing.T) {
	got := StripANSI("\x1b[31mred\x1b[0m text")
	if got != "red text" {
		t.Fatalf("StripANSI = %q", got)
	}
}

package inspect

import (
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

func TestStripANSI(t *testing.T) {
	got := StripANSI("\x1b[31mred\x1b[0m text")
	if got != "red text" {
		t.Fatalf("StripANSI = %q", got)
	}
}

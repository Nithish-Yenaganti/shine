package parser

import (
	"testing"

	"shine/internal/model"
)

func TestParseCoreBlocks(t *testing.T) {
	source := "# shine\n\n" +
		"Text with **bold**, *italic*, `code`, and [link](https://example.com).\n\n" +
		"> [!NOTE]\n> Callout body.\n\n" +
		"- [x] Done\n\n" +
		"| Feature | Status |\n| --- | --- |\n| Tables | Good |\n\n" +
		"```go\nfmt.Println(\"hi\")\n```\n\n" +
		"---\n"
	doc, err := Parse([]byte(source), "test.md")
	if err != nil {
		t.Fatal(err)
	}

	want := []model.BlockKind{
		model.BlockHeading,
		model.BlockParagraph,
		model.BlockCallout,
		model.BlockList,
		model.BlockTable,
		model.BlockCode,
		model.BlockDivider,
	}
	for i, kind := range want {
		if doc.Blocks[i].Kind != kind {
			t.Fatalf("block %d kind = %v, want %v", i, doc.Blocks[i].Kind, kind)
		}
	}
	if len(doc.Headings) != 1 || doc.Headings[0].Text != "shine" {
		t.Fatalf("unexpected headings: %#v", doc.Headings)
	}
}

func TestParseInlineStyles(t *testing.T) {
	doc, err := Parse([]byte("Text with **bold**, *italic*, `code`, ~~old~~, and [link](https://example.com)."), "test.md")
	if err != nil {
		t.Fatal(err)
	}
	spans := doc.Blocks[0].Content.Spans
	var bold, italic, code, strike, link bool
	for _, span := range spans {
		bold = bold || span.Bold
		italic = italic || span.Italic
		code = code || span.Code
		strike = strike || span.Strike
		link = link || span.Link != ""
	}
	if !bold || !italic || !code || !strike || !link {
		t.Fatalf("missing inline styles in %#v", spans)
	}
}

func TestParseNestedLists(t *testing.T) {
	doc, err := Parse([]byte("- Parent\n  - Child\n  - Second child\n- Sibling\n"), "test.md")
	if err != nil {
		t.Fatal(err)
	}
	list := doc.Blocks[0]
	if len(list.Items) != 2 {
		t.Fatalf("top-level item count = %d, want 2", len(list.Items))
	}
	if got := list.Items[0].Content.Plain(); got != "Parent" {
		t.Fatalf("parent content = %q, want Parent", got)
	}
	if len(list.Items[0].Children) != 2 {
		t.Fatalf("nested item count = %d, want 2", len(list.Items[0].Children))
	}
	if got := list.Items[0].Children[0].Content.Plain(); got != "Child" {
		t.Fatalf("child content = %q, want Child", got)
	}
}

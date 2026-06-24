package parser

import (
	"bytes"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	east "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"

	"shine/internal/model"
)

var markdown = goldmark.New(
	goldmark.WithExtensions(extension.GFM),
	goldmark.WithParserOptions(parser.WithAutoHeadingID()),
)

type inlineStyle struct {
	bold   bool
	italic bool
	code   bool
	strike bool
	link   string
}

func Parse(source []byte, sourceName string) (model.Document, error) {
	root := markdown.Parser().Parse(text.NewReader(source))
	var blocks []model.Block

	for node := root.FirstChild(); node != nil; node = node.NextSibling() {
		block, ok := parseBlock(node, source)
		if ok {
			blocks = append(blocks, block)
		}
	}

	return model.NewDocument(sourceName, blocks), nil
}

func parseBlock(node ast.Node, source []byte) (model.Block, bool) {
	switch n := node.(type) {
	case *ast.Heading:
		return model.Block{
			Kind:    model.BlockHeading,
			Level:   n.Level,
			Content: collectInline(n, source, inlineStyle{}),
		}, true
	case *ast.Paragraph:
		if image, ok := paragraphImage(n, source); ok {
			return image, true
		}
		return model.Block{Kind: model.BlockParagraph, Content: collectInline(n, source, inlineStyle{})}, true
	case *ast.FencedCodeBlock:
		return model.Block{
			Kind:     model.BlockCode,
			Language: string(n.Language(source)),
			Code:     codeText(n, source),
		}, true
	case *ast.CodeBlock:
		return model.Block{Kind: model.BlockCode, Code: codeText(n, source)}, true
	case *ast.Blockquote:
		return quoteOrCallout(n, source), true
	case *ast.List:
		return parseList(n, source), true
	case *east.Table:
		return parseTable(n, source), true
	case *ast.ThematicBreak:
		return model.Block{Kind: model.BlockDivider}, true
	default:
		if node.Kind() == ast.KindImage {
			return parseImage(node, source), true
		}
		return model.Block{}, false
	}
}

func collectInline(node ast.Node, source []byte, style inlineStyle) model.RichText {
	var spans []model.InlineSpan
	var walk func(ast.Node, inlineStyle)
	walk = func(n ast.Node, style inlineStyle) {
		switch n := n.(type) {
		case *ast.Text:
			text := string(n.Segment.Value(source))
			if n.SoftLineBreak() || n.HardLineBreak() {
				text += " "
			}
			spans = appendSpan(spans, model.InlineSpan{
				Text:   text,
				Bold:   style.bold,
				Italic: style.italic,
				Code:   style.code,
				Strike: style.strike,
				Link:   style.link,
			})
		case *ast.CodeSpan:
			codeStyle := style
			codeStyle.code = true
			for child := n.FirstChild(); child != nil; child = child.NextSibling() {
				walk(child, codeStyle)
			}
			return
		case *ast.Emphasis:
			next := style
			if n.Level == 2 {
				next.bold = true
			} else {
				next.italic = true
			}
			for child := n.FirstChild(); child != nil; child = child.NextSibling() {
				walk(child, next)
			}
			return
		case *ast.Link:
			next := style
			next.link = string(n.Destination)
			for child := n.FirstChild(); child != nil; child = child.NextSibling() {
				walk(child, next)
			}
			return
		case *east.Strikethrough:
			next := style
			next.strike = true
			for child := n.FirstChild(); child != nil; child = child.NextSibling() {
				walk(child, next)
			}
			return
		case *ast.Image:
			alt := imageAlt(n, source)
			spans = appendSpan(spans, model.InlineSpan{Text: alt, Link: string(n.Destination)})
			return
		case *east.TaskCheckBox:
			return
		}

		for child := n.FirstChild(); child != nil; child = child.NextSibling() {
			walk(child, style)
		}
	}
	walk(node, style)
	return normalizeRich(model.RichText{Spans: spans})
}

func appendSpan(spans []model.InlineSpan, span model.InlineSpan) []model.InlineSpan {
	if span.Text == "" {
		return spans
	}
	if len(spans) > 0 {
		last := &spans[len(spans)-1]
		if last.Bold == span.Bold && last.Italic == span.Italic && last.Code == span.Code && last.Strike == span.Strike && last.Link == span.Link {
			last.Text += span.Text
			return spans
		}
	}
	return append(spans, span)
}

func normalizeRich(r model.RichText) model.RichText {
	var out []model.InlineSpan
	for _, span := range r.Spans {
		parts := strings.Fields(span.Text)
		if len(parts) == 0 {
			continue
		}
		span.Text = strings.Join(parts, " ")
		out = appendSpan(out, span)
	}
	return model.RichText{Spans: out}
}

func codeText(node ast.Node, source []byte) string {
	var buf bytes.Buffer
	lines := node.Lines()
	for i := 0; i < lines.Len(); i++ {
		segment := lines.At(i)
		buf.Write(segment.Value(source))
	}
	return strings.TrimRight(buf.String(), "\n")
}

func quoteOrCallout(node *ast.Blockquote, source []byte) model.Block {
	var parts []model.RichText
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		if child.Kind() == ast.KindParagraph {
			parts = append(parts, collectInline(child, source, inlineStyle{}))
		}
	}
	plain := ""
	if len(parts) > 0 {
		plain = parts[0].Plain()
	}
	if strings.HasPrefix(plain, "[!") {
		end := strings.Index(plain, "]")
		if end > 2 {
			if kind, ok := model.ParseCalloutKind(plain[2:end]); ok {
				title := strings.TrimSpace(plain[end+1:])
				if title == "" {
					title = string(kind)
				}
				body := joinRich(parts[1:])
				return model.Block{Kind: model.BlockCallout, Callout: kind, Title: title, Content: body}
			}
		}
	}
	return model.Block{Kind: model.BlockQuote, Content: joinRich(parts)}
}

func joinRich(parts []model.RichText) model.RichText {
	var spans []model.InlineSpan
	for i, part := range parts {
		if i > 0 {
			spans = appendSpan(spans, model.InlineSpan{Text: " "})
		}
		for _, span := range part.Spans {
			spans = appendSpan(spans, span)
		}
	}
	return model.RichText{Spans: spans}
}

func parseList(n *ast.List, source []byte) model.Block {
	return model.Block{Kind: model.BlockList, Ordered: n.IsOrdered(), Items: parseListItems(n, source)}
}

func parseListItems(n *ast.List, source []byte) []model.ListItem {
	var items []model.ListItem
	for itemNode := n.FirstChild(); itemNode != nil; itemNode = itemNode.NextSibling() {
		content := collectListItemInline(itemNode, source)
		plain := content.Plain()
		checked := findTaskCheckBox(itemNode)
		if strings.HasPrefix(plain, "[x] ") || strings.HasPrefix(plain, "[X] ") {
			v := true
			checked = &v
			content = model.Plain(strings.TrimSpace(plain[4:]))
		} else if strings.HasPrefix(plain, "[ ] ") {
			v := false
			checked = &v
			content = model.Plain(strings.TrimSpace(plain[4:]))
		}
		childItems, childOrdered := nestedList(itemNode, source)
		items = append(items, model.ListItem{
			Content:         content,
			Checked:         checked,
			Children:        childItems,
			ChildrenOrdered: childOrdered,
		})
	}
	return items
}

func collectListItemInline(itemNode ast.Node, source []byte) model.RichText {
	var parts []model.RichText
	for child := itemNode.FirstChild(); child != nil; child = child.NextSibling() {
		if _, ok := child.(*ast.List); ok {
			continue
		}
		parts = append(parts, collectInline(child, source, inlineStyle{}))
	}
	return joinRich(parts)
}

func nestedList(itemNode ast.Node, source []byte) ([]model.ListItem, bool) {
	for child := itemNode.FirstChild(); child != nil; child = child.NextSibling() {
		if list, ok := child.(*ast.List); ok {
			return parseListItems(list, source), list.IsOrdered()
		}
	}
	return nil, false
}

func findTaskCheckBox(node ast.Node) *bool {
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		if checkbox, ok := child.(*east.TaskCheckBox); ok {
			value := checkbox.IsChecked
			return &value
		}
		if value := findTaskCheckBox(child); value != nil {
			return value
		}
	}
	return nil
}

func parseTable(n *east.Table, source []byte) model.Block {
	var headers []string
	var rows [][]string
	for rowNode := n.FirstChild(); rowNode != nil; rowNode = rowNode.NextSibling() {
		cells := tableRowCells(rowNode, source)
		if len(cells) == 0 {
			continue
		}
		if headers == nil {
			headers = cells
		} else {
			rows = append(rows, cells)
		}
	}
	return model.Block{Kind: model.BlockTable, Headers: headers, Rows: rows}
}

func tableRowCells(rowNode ast.Node, source []byte) []string {
	var cells []string
	for cellNode := rowNode.FirstChild(); cellNode != nil; cellNode = cellNode.NextSibling() {
		cells = append(cells, collectInline(cellNode, source, inlineStyle{}).Plain())
	}
	return cells
}

func parseImage(node ast.Node, source []byte) model.Block {
	img, ok := node.(*ast.Image)
	if !ok {
		return model.Block{Kind: model.BlockParagraph, Content: model.Plain("")}
	}
	return model.Block{
		Kind: model.BlockImage,
		Alt:  collectInline(img, source, inlineStyle{}).Plain(),
		URL:  string(img.Destination),
	}
}

func paragraphImage(n *ast.Paragraph, source []byte) (model.Block, bool) {
	if n.FirstChild() == nil || n.FirstChild() != n.LastChild() {
		return model.Block{}, false
	}
	img, ok := n.FirstChild().(*ast.Image)
	if !ok {
		return model.Block{}, false
	}
	return model.Block{
		Kind: model.BlockImage,
		Alt:  imageAlt(img, source),
		URL:  string(img.Destination),
	}, true
}

func imageAlt(img *ast.Image, source []byte) string {
	var parts []string
	for child := img.FirstChild(); child != nil; child = child.NextSibling() {
		if text, ok := child.(*ast.Text); ok {
			parts = append(parts, string(text.Segment.Value(source)))
		}
	}
	return model.Normalize(strings.Join(parts, " "))
}

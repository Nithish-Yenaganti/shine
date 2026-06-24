package model

import "strings"

type Document struct {
	SourceName string
	Blocks     []Block
	Headings   []HeadingRef
}

type HeadingRef struct {
	Level      int
	Text       string
	BlockIndex int
}

type BlockKind int

const (
	BlockHeading BlockKind = iota
	BlockParagraph
	BlockCode
	BlockQuote
	BlockCallout
	BlockList
	BlockTable
	BlockDivider
	BlockImage
)

type Block struct {
	Kind     BlockKind
	Level    int
	Content  RichText
	Language string
	Code     string
	Callout  CalloutKind
	Title    string
	Ordered  bool
	Items    []ListItem
	Headers  []string
	Rows     [][]string
	Alt      string
	URL      string
}

type ListItem struct {
	Content         RichText
	Checked         *bool
	Children        []ListItem
	ChildrenOrdered bool
}

type CalloutKind string

const (
	CalloutNote      CalloutKind = "NOTE"
	CalloutTip       CalloutKind = "TIP"
	CalloutWarning   CalloutKind = "WARNING"
	CalloutImportant CalloutKind = "IMPORTANT"
	CalloutCaution   CalloutKind = "CAUTION"
)

type RichText struct {
	Spans []InlineSpan
}

type InlineSpan struct {
	Text   string
	Bold   bool
	Italic bool
	Code   bool
	Strike bool
	Link   string
}

func NewDocument(sourceName string, blocks []Block) Document {
	doc := Document{SourceName: sourceName, Blocks: blocks}
	for i, block := range blocks {
		if block.Kind == BlockHeading {
			doc.Headings = append(doc.Headings, HeadingRef{
				Level:      block.Level,
				Text:       block.Content.Plain(),
				BlockIndex: i,
			})
		}
	}
	return doc
}

func Plain(text string) RichText {
	text = Normalize(text)
	if text == "" {
		return RichText{}
	}
	return RichText{Spans: []InlineSpan{{Text: text}}}
}

func (r RichText) Plain() string {
	var out strings.Builder
	for _, span := range r.Spans {
		out.WriteString(span.Text)
	}
	return Normalize(out.String())
}

func Normalize(text string) string {
	return strings.Join(strings.Fields(text), " ")
}

func ParseCalloutKind(value string) (CalloutKind, bool) {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case "NOTE":
		return CalloutNote, true
	case "TIP":
		return CalloutTip, true
	case "WARNING", "WARN":
		return CalloutWarning, true
	case "IMPORTANT":
		return CalloutImportant, true
	case "CAUTION", "DANGER":
		return CalloutCaution, true
	default:
		return "", false
	}
}

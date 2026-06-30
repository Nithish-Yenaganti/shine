package inspect

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Nithish-Yenaganti/shine/internal/model"
)

type Warning struct {
	Message string
}

func Outline(doc model.Document) string {
	if len(doc.Headings) == 0 {
		return "no headings\n"
	}
	var lines []string
	for _, heading := range doc.Headings {
		marker := strings.Repeat("#", heading.Level)
		lines = append(lines, fmt.Sprintf("%s %s", marker, heading.Text))
	}
	return strings.Join(lines, "\n") + "\n"
}

func Check(doc model.Document, sourcePath string) []Warning {
	var warnings []Warning
	if len(doc.Blocks) == 0 {
		warnings = append(warnings, Warning{Message: "document is empty"})
		return warnings
	}
	if len(doc.Headings) == 0 {
		warnings = append(warnings, Warning{Message: "document has no headings"})
	} else if doc.Headings[0].Level != 1 {
		warnings = append(warnings, Warning{Message: "first heading is not an H1"})
	}
	warnings = append(warnings, headingWarnings(doc.Headings)...)
	for i, block := range doc.Blocks {
		blockNumber := i + 1
		switch block.Kind {
		case model.BlockHeading:
			if strings.TrimSpace(block.Content.Plain()) == "" {
				warnings = append(warnings, Warning{Message: fmt.Sprintf("block %d has an empty heading", blockNumber)})
			}
		case model.BlockTable:
			warnings = append(warnings, tableWarnings(blockNumber, block)...)
		case model.BlockImage:
			if block.Alt == "" {
				warnings = append(warnings, Warning{Message: fmt.Sprintf("block %d image has no alt text", blockNumber)})
			}
			if missingLocalImage(sourcePath, block.URL) {
				warnings = append(warnings, Warning{Message: fmt.Sprintf("block %d image file not found: %s", blockNumber, block.URL)})
			}
		}
		warnings = append(warnings, linkWarnings(blockNumber, block, sourcePath)...)
	}
	return warnings
}

func FormatWarnings(warnings []Warning) string {
	if len(warnings) == 0 {
		return "OK: no markdown issues found\n"
	}
	var lines []string
	lines = append(lines, fmt.Sprintf("%d markdown warning(s):", len(warnings)))
	for _, warning := range warnings {
		lines = append(lines, "- "+warning.Message)
	}
	return strings.Join(lines, "\n") + "\n"
}

func StripANSI(value string) string {
	var out strings.Builder
	inEscape := false
	for _, r := range value {
		if r == '\x1b' {
			inEscape = true
			continue
		}
		if inEscape {
			if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
				inEscape = false
			}
			continue
		}
		out.WriteRune(r)
	}
	return out.String()
}

func headingWarnings(headings []model.HeadingRef) []Warning {
	var warnings []Warning
	seen := map[string]bool{}
	h1Count := 0
	previousLevel := 0
	for _, heading := range headings {
		text := strings.TrimSpace(heading.Text)
		if heading.Level == 1 {
			h1Count++
		}
		if previousLevel > 0 && heading.Level > previousLevel+1 {
			warnings = append(warnings, Warning{Message: fmt.Sprintf("heading %q jumps from H%d to H%d", text, previousLevel, heading.Level)})
		}
		key := strings.ToLower(text)
		if key != "" {
			if seen[key] {
				warnings = append(warnings, Warning{Message: fmt.Sprintf("duplicate heading %q", text)})
			}
			seen[key] = true
		}
		previousLevel = heading.Level
	}
	if h1Count > 1 {
		warnings = append(warnings, Warning{Message: "document has multiple H1 headings"})
	}
	return warnings
}

func tableWarnings(blockNumber int, block model.Block) []Warning {
	var warnings []Warning
	width := len(block.Headers)
	if width == 0 {
		return append(warnings, Warning{Message: fmt.Sprintf("block %d table has no header", blockNumber)})
	}
	for i, header := range block.Headers {
		if strings.TrimSpace(header) == "" {
			warnings = append(warnings, Warning{Message: fmt.Sprintf("block %d table header column %d is empty", blockNumber, i+1)})
		}
		if len([]rune(header)) > 40 {
			warnings = append(warnings, Warning{Message: fmt.Sprintf("block %d table header column %d is very long", blockNumber, i+1)})
		}
	}
	for i, row := range block.Rows {
		if len(row) != width {
			warnings = append(warnings, Warning{Message: fmt.Sprintf("block %d table row %d has %d cells, expected %d", blockNumber, i+1, len(row), width)})
		}
		for j, cell := range row {
			if len([]rune(cell)) > 80 {
				warnings = append(warnings, Warning{Message: fmt.Sprintf("block %d table row %d column %d is very long", blockNumber, i+1, j+1)})
			}
		}
	}
	return warnings
}

func linkWarnings(blockNumber int, block model.Block, sourcePath string) []Warning {
	var warnings []Warning
	for _, link := range blockLinks(block) {
		text := strings.TrimSpace(link.Text)
		if text == "" {
			warnings = append(warnings, Warning{Message: fmt.Sprintf("block %d link has no text: %s", blockNumber, link.Link)})
		}
		if isRawURLText(text, link.Link) {
			warnings = append(warnings, Warning{Message: fmt.Sprintf("block %d link text is a raw URL: %s", blockNumber, link.Link)})
		}
		if missingLocalLink(sourcePath, link.Link) {
			warnings = append(warnings, Warning{Message: fmt.Sprintf("block %d link file not found: %s", blockNumber, link.Link)})
		}
	}
	return warnings
}

func blockLinks(block model.Block) []model.InlineSpan {
	var links []model.InlineSpan
	addRichTextLinks := func(content model.RichText) {
		for _, span := range content.Spans {
			if span.Link != "" {
				links = append(links, span)
			}
		}
	}
	addListLinks := func(items []model.ListItem) {}
	addListLinks = func(items []model.ListItem) {
		for _, item := range items {
			addRichTextLinks(item.Content)
			addListLinks(item.Children)
		}
	}
	addRichTextLinks(block.Content)
	addListLinks(block.Items)
	return links
}

func isRawURLText(text string, link string) bool {
	normalizedText := strings.TrimRight(text, "/")
	normalizedLink := strings.TrimRight(link, "/")
	return (strings.HasPrefix(normalizedText, "http://") || strings.HasPrefix(normalizedText, "https://")) && normalizedText == normalizedLink
}

func missingLocalLink(sourcePath string, link string) bool {
	path, ok := localFileTarget(sourcePath, link)
	if !ok {
		return false
	}
	if _, err := os.Stat(path); err != nil {
		return os.IsNotExist(err)
	}
	return false
}

func missingLocalImage(sourcePath string, imagePath string) bool {
	path, ok := localFileTarget(sourcePath, imagePath)
	if !ok {
		return false
	}
	if _, err := os.Stat(path); err != nil {
		return os.IsNotExist(err)
	}
	return false
}

func localFileTarget(sourcePath string, target string) (string, bool) {
	if target == "" || sourcePath == "" || isExternalTarget(target) || strings.HasPrefix(target, "#") {
		return "", false
	}
	path := strings.SplitN(target, "#", 2)[0]
	if path == "" {
		return "", false
	}
	path = filepath.Clean(path)
	if !filepath.IsAbs(path) {
		path = filepath.Join(filepath.Dir(sourcePath), path)
	}
	return path, true
}

func isExternalTarget(target string) bool {
	lower := strings.ToLower(target)
	return strings.Contains(lower, "://") ||
		strings.HasPrefix(lower, "mailto:") ||
		strings.HasPrefix(lower, "tel:")
}

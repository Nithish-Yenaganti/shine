package inspect

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"shine/internal/model"
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
	for i, block := range doc.Blocks {
		switch block.Kind {
		case model.BlockHeading:
			if strings.TrimSpace(block.Content.Plain()) == "" {
				warnings = append(warnings, Warning{Message: fmt.Sprintf("block %d has an empty heading", i+1)})
			}
		case model.BlockTable:
			warnings = append(warnings, tableWarnings(i+1, block)...)
		case model.BlockImage:
			if block.Alt == "" {
				warnings = append(warnings, Warning{Message: fmt.Sprintf("block %d image has no alt text", i+1)})
			}
			if missingLocalImage(sourcePath, block.URL) {
				warnings = append(warnings, Warning{Message: fmt.Sprintf("block %d image file not found: %s", i+1, block.URL)})
			}
		}
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

func tableWarnings(blockNumber int, block model.Block) []Warning {
	var warnings []Warning
	width := len(block.Headers)
	if width == 0 {
		return append(warnings, Warning{Message: fmt.Sprintf("block %d table has no header", blockNumber)})
	}
	for i, row := range block.Rows {
		if len(row) != width {
			warnings = append(warnings, Warning{Message: fmt.Sprintf("block %d table row %d has %d cells, expected %d", blockNumber, i+1, len(row), width)})
		}
	}
	return warnings
}

func missingLocalImage(sourcePath string, imagePath string) bool {
	if imagePath == "" || strings.Contains(imagePath, "://") || sourcePath == "" {
		return false
	}
	path := filepath.Clean(imagePath)
	if !filepath.IsAbs(path) {
		path = filepath.Join(filepath.Dir(sourcePath), path)
	}
	if _, err := os.Stat(path); err != nil {
		return os.IsNotExist(err)
	}
	return false
}

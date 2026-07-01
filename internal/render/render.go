package render

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/charmbracelet/lipgloss"

	"github.com/Nithish-Yenaganti/shine/internal/config"
	"github.com/Nithish-Yenaganti/shine/internal/model"
)

type Renderer struct {
	Width int
	Theme config.Theme
}

type richWord struct {
	text  string
	style lipgloss.Style
}

func New(width int, theme config.Theme) Renderer {
	if width < 32 {
		width = 32
	}
	return Renderer{Width: width, Theme: theme}
}

func (r Renderer) Render(doc model.Document) string {
	var blocks []string
	for _, block := range doc.Blocks {
		rendered := r.renderBlock(block)
		if rendered != "" {
			blocks = append(blocks, rendered)
		}
	}
	gap := "\n\n"
	if r.Theme.BlockGap == 0 {
		gap = "\n"
	}
	return strings.TrimRight(strings.Join(blocks, gap), "\n") + "\n"
}

func (r Renderer) renderBlock(block model.Block) string {
	switch block.Kind {
	case model.BlockHeading:
		return r.heading(block.Level, block.Content.Plain())
	case model.BlockParagraph:
		return r.rich(block.Content, r.Width, "")
	case model.BlockQuote:
		return r.accent(block.Content, r.quoteStyle(), r.color(r.Theme.Quote), "")
	case model.BlockCallout:
		return r.callout(block)
	case model.BlockList:
		return r.list(block)
	case model.BlockTable:
		return r.table(block)
	case model.BlockCode:
		return r.code(block)
	case model.BlockDivider:
		return r.rule()
	case model.BlockImage:
		return r.image(block)
	default:
		return ""
	}
}

func (r Renderer) heading(level int, text string) string {
	switch level {
	case 1:
		title := r.headingStyle().Bold(true).Render(text)
		return title + "\n" + r.borderStyle().Render(strings.Repeat("━", clamp(lipgloss.Width(text), 3, r.Width/2)))
	case 2:
		title := r.headingStyle().Render(text)
		return title + "\n" + r.borderStyle().Render(strings.Repeat("─", clamp(lipgloss.Width(text), 3, r.Width/2)))
	default:
		return r.mutedStyle().Render(strings.Repeat("·", max(0, level-2))+" ") + r.headingStyle().Render(text)
	}
}

func (r Renderer) callout(block model.Block) string {
	accent := r.color(r.Theme.CalloutNote)
	if block.Callout == model.CalloutTip {
		accent = r.color(r.Theme.CalloutTip)
	}
	if block.Callout == model.CalloutWarning || block.Callout == model.CalloutCaution {
		accent = r.color(r.Theme.CalloutWarning)
	}
	title := r.style().Foreground(accent).Bold(true).Render(string(block.Callout))
	if block.Title != "" && block.Title != string(block.Callout) {
		title += r.style().Foreground(r.color(r.Theme.Body)).Render("  " + block.Title)
	}
	body := r.rich(block.Content, r.Width-3, "")
	lines := []string{"┃ " + title}
	if body != "" {
		for _, line := range strings.Split(body, "\n") {
			lines = append(lines, r.style().Foreground(accent).Render("┃ ")+line)
		}
	}
	return strings.Join(lines, "\n")
}

func (r Renderer) accent(content model.RichText, style lipgloss.Style, accent lipgloss.Color, title string) string {
	body := r.rich(content, r.Width-3, "")
	var lines []string
	if title != "" {
		lines = append(lines, r.style().Foreground(accent).Bold(true).Render("┃ "+title))
	}
	for _, line := range strings.Split(body, "\n") {
		lines = append(lines, r.style().Foreground(accent).Render("┃ ")+style.Render(line))
	}
	return strings.Join(lines, "\n")
}

func (r Renderer) list(block model.Block) string {
	return strings.Join(r.listItems(block.Items, block.Ordered, 0), "\n")
}

func (r Renderer) listItems(items []model.ListItem, ordered bool, depth int) []string {
	var lines []string
	for i, item := range items {
		marker := "•"
		if ordered {
			marker = strconv.Itoa(i+1) + "."
		}
		if item.Checked != nil {
			if *item.Checked {
				marker = "✓"
			} else {
				marker = "☐"
			}
		}
		indent := strings.Repeat("  ", depth)
		firstPrefix := r.mutedStyle().Render(indent + marker + " ")
		restPrefix := r.mutedStyle().Render(indent + strings.Repeat(" ", lipgloss.Width(marker)+1))
		width := r.Width - lipgloss.Width(restPrefix)
		lines = append(lines, r.richWithPrefix(item.Content, width, firstPrefix, restPrefix))
		if len(item.Children) > 0 {
			lines = append(lines, r.listItems(item.Children, item.ChildrenOrdered, depth+1)...)
		}
	}
	return lines
}

func (r Renderer) table(block model.Block) string {
	if len(block.Headers) == 0 {
		return ""
	}
	cols := len(block.Headers)
	for _, row := range block.Rows {
		if len(row) > cols {
			cols = len(row)
		}
	}
	widths := make([]int, cols)
	for i := 0; i < cols; i++ {
		widths[i] = 4
	}
	for i, header := range block.Headers {
		widths[i] = max(widths[i], min(lipgloss.Width(header), 28))
	}
	for _, row := range block.Rows {
		for i, cell := range row {
			widths[i] = max(widths[i], min(lipgloss.Width(cell), 28))
		}
	}
	total := cols + 1
	for _, w := range widths {
		total += w + 2
	}
	for total > r.Width && largest(widths) > 6 {
		i := largestIndex(widths)
		widths[i]--
		total--
	}

	var out []string
	out = append(out, r.tableBorder(widths, "top"))
	out = append(out, r.tableRow(block.Headers, widths, true)...)
	out = append(out, r.tableBorder(widths, "mid"))
	for _, row := range block.Rows {
		out = append(out, r.tableRow(row, widths, false)...)
	}
	out = append(out, r.tableBorder(widths, "bottom"))
	return strings.Join(out, "\n")
}

func (r Renderer) tableRow(cells []string, widths []int, header bool) []string {
	wrapped := make([][]string, len(widths))
	height := 1
	for i, width := range widths {
		cell := ""
		if i < len(cells) {
			cell = cells[i]
		}
		wrapped[i] = wrapPlain(cell, width)
		height = max(height, len(wrapped[i]))
	}
	var lines []string
	for line := 0; line < height; line++ {
		var b strings.Builder
		b.WriteString(r.borderStyle().Render("│"))
		for i, width := range widths {
			value := ""
			if line < len(wrapped[i]) {
				value = wrapped[i][line]
			}
			style := r.bodyStyle()
			if header {
				style = r.style().Foreground(r.color(r.Theme.TableHeader)).Bold(true)
			}
			b.WriteString(" ")
			b.WriteString(style.Render(padRight(value, width)))
			b.WriteString(" ")
			b.WriteString(r.borderStyle().Render("│"))
		}
		lines = append(lines, b.String())
	}
	return lines
}

func (r Renderer) tableBorder(widths []int, pos string) string {
	left, join, right := "┌", "┬", "┐"
	if pos == "mid" {
		left, join, right = "├", "┼", "┤"
	}
	if pos == "bottom" {
		left, join, right = "└", "┴", "┘"
	}
	var b strings.Builder
	b.WriteString(left)
	for i, width := range widths {
		if i > 0 {
			b.WriteString(join)
		}
		b.WriteString(strings.Repeat("─", width+2))
	}
	b.WriteString(right)
	return r.borderStyle().Render(b.String())
}

func (r Renderer) code(block model.Block) string {
	label := block.Language
	if label == "" {
		label = "text"
	}
	codeLines := r.codeLines(block.Code, label)
	numberWidth := len(strconv.Itoa(len(codeLines)))
	contentWidth := max(18, r.Width-2)
	showLineNumbers := r.showCodeLineNumbers(label)
	if showLineNumbers {
		contentWidth -= numberWidth + 3
	}
	topLabel := " " + label + " "
	topFill := strings.Repeat("─", max(0, r.Width-lipgloss.Width(topLabel)-2))
	lines := []string{r.borderStyle().Render("┌" + topLabel + topFill)}
	for i, line := range codeLines {
		line = lipgloss.NewStyle().MaxWidth(contentWidth).Render(line)
		prefix := r.mutedStyle().Render("│ ")
		if showLineNumbers {
			prefix = r.mutedStyle().Render(fmt.Sprintf("│ %*d ", numberWidth, i+1))
		}
		lines = append(lines, prefix+r.codeBackgroundStyle().Render(padRight(line, contentWidth)))
	}
	lines = append(lines, r.borderStyle().Render("└"+strings.Repeat("─", max(1, r.Width-1))))
	return strings.Join(lines, "\n")
}

func (r Renderer) showCodeLineNumbers(language string) bool {
	if !r.Theme.CodeLineNumbers {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(language)) {
	case "ascii", "art", "banner":
		return false
	default:
		return true
	}
}

func (r Renderer) codeLines(code string, language string) []string {
	raw := splitCode(code)
	if r.Theme.Name == "mono" || r.Theme.Name == "daylight" || r.Theme.Name == "claude" {
		return raw
	}
	lexer := lexers.Get(language)
	if lexer == nil {
		lexer = lexers.Analyse(code)
	}
	if lexer == nil {
		return raw
	}
	formatter := formatters.Get("terminal16m")
	if formatter == nil {
		return raw
	}
	styleName := "github-dark"
	if r.Theme.Name == "daylight" {
		styleName = "github"
	}
	style := styles.Get(styleName)
	if style == nil {
		return raw
	}
	var out bytes.Buffer
	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		return raw
	}
	if err := formatter.Format(&out, style, iterator); err != nil {
		return raw
	}
	return splitCode(strings.TrimRight(out.String(), "\n"))
}

func (r Renderer) image(block model.Block) string {
	status := "image preview unavailable in this terminal"
	if terminalImageProtocol() != "" && localFileExists(block.URL) {
		status = "local image can render in " + terminalImageProtocol()
	}
	return strings.Join([]string{
		r.mutedStyle().Render("image  ") + r.bodyStyle().Bold(true).Render(block.Alt),
		r.mutedStyle().Render("       ") + r.linkStyle().Render(block.URL),
		r.mutedStyle().Render("       " + status),
	}, "\n")
}

func (r Renderer) rule() string {
	return r.borderStyle().Render(strings.Repeat("─", min(r.Width, 72)))
}

func (r Renderer) rich(content model.RichText, width int, prefix string) string {
	return r.richWithPrefix(content, width, prefix, prefix)
}

func (r Renderer) richWithPrefix(content model.RichText, width int, firstPrefix string, restPrefix string) string {
	if width < 10 {
		width = 10
	}
	var words []richWord
	for _, span := range content.Spans {
		style := r.bodyStyle()
		if span.Bold {
			style = style.Bold(true)
		}
		if span.Italic {
			style = style.Italic(true)
		}
		if span.Code {
			style = r.inlineCodeStyle()
		}
		if span.Strike {
			style = style.Strikethrough(true)
		}
		if span.Link != "" {
			style = r.linkStyle()
		}
		for _, part := range strings.Fields(span.Text) {
			if span.Link != "" {
				part += "↗"
			}
			words = append(words, richWord{text: part, style: style})
		}
	}
	if len(words) == 0 {
		return ""
	}
	var lines []string
	var current []richWord
	currentWidth := 0
	prefix := firstPrefix
	for _, w := range words {
		sep := 0
		if currentWidth > 0 && !sticks(w.text) {
			sep = 1
		}
		wordWidth := lipgloss.Width(w.text)
		if currentWidth > 0 && currentWidth+sep+wordWidth > width {
			lines = append(lines, renderWords(current, prefix))
			prefix = restPrefix
			current = nil
			currentWidth = 0
		}
		if currentWidth > 0 && !sticks(w.text) {
			current = append(current, richWord{text: " ", style: r.bodyStyle()})
			currentWidth++
		}
		current = append(current, w)
		currentWidth += wordWidth
	}
	if len(current) > 0 {
		lines = append(lines, renderWords(current, prefix))
	}
	return strings.Join(lines, "\n")
}

func renderWords(words []richWord, prefix string) string {
	var b strings.Builder
	b.WriteString(prefix)
	for _, w := range words {
		b.WriteString(w.style.Render(w.text))
	}
	return b.String()
}

func (r Renderer) style() lipgloss.Style {
	return lipgloss.NewStyle()
}

func (r Renderer) color(value string) lipgloss.Color {
	return lipgloss.Color(value)
}

func (r Renderer) bodyStyle() lipgloss.Style {
	return r.style().Foreground(r.color(r.Theme.Body))
}

func (r Renderer) headingStyle() lipgloss.Style {
	return r.style().Foreground(r.color(r.Theme.Heading)).Bold(true)
}

func (r Renderer) mutedStyle() lipgloss.Style {
	return r.style().Foreground(r.color(r.Theme.Muted))
}

func (r Renderer) borderStyle() lipgloss.Style {
	return r.style().Foreground(r.color(r.Theme.Border))
}

func (r Renderer) codeStyle() lipgloss.Style {
	return r.style().Foreground(r.color(r.Theme.Code))
}

func (r Renderer) codeBackgroundStyle() lipgloss.Style {
	return r.style().Foreground(r.color(r.Theme.Code)).Background(r.color(r.Theme.CodeBackground))
}

func (r Renderer) inlineCodeStyle() lipgloss.Style {
	return r.style().Foreground(r.color(r.Theme.InlineCode)).Bold(true)
}

func (r Renderer) linkStyle() lipgloss.Style {
	return r.style().Foreground(r.color(r.Theme.Link)).Underline(true)
}

func (r Renderer) quoteStyle() lipgloss.Style {
	return r.style().Foreground(r.color(r.Theme.Body))
}

func wrapPlain(text string, width int) []string {
	var lines []string
	var current []string
	currentWidth := 0
	for _, word := range strings.Fields(text) {
		w := lipgloss.Width(word)
		if currentWidth > 0 && currentWidth+1+w > width {
			lines = append(lines, strings.Join(current, " "))
			current = nil
			currentWidth = 0
		}
		current = append(current, word)
		if currentWidth == 0 {
			currentWidth = w
		} else {
			currentWidth += 1 + w
		}
	}
	if len(current) > 0 {
		lines = append(lines, strings.Join(current, " "))
	}
	if len(lines) == 0 {
		return []string{""}
	}
	return lines
}

func splitCode(code string) []string {
	if code == "" {
		return []string{""}
	}
	return strings.Split(strings.TrimRight(code, "\n"), "\n")
}

func padRight(text string, width int) string {
	return text + strings.Repeat(" ", max(0, width-lipgloss.Width(text)))
}

func largest(values []int) int {
	best := 0
	for _, value := range values {
		best = max(best, value)
	}
	return best
}

func largestIndex(values []int) int {
	best := 0
	for i, value := range values {
		if value > values[best] {
			best = i
		}
	}
	return best
}

func sticks(text string) bool {
	if text == "" {
		return false
	}
	if strings.HasPrefix(text, "./") || strings.HasPrefix(text, "../") {
		return false
	}
	switch []rune(text)[0] {
	case '.', ',', ';', ':', '!', '?', ')', ']', '}':
		return true
	default:
		return false
	}
}

func terminalImageProtocol() string {
	term := strings.ToLower(os.Getenv("TERM"))
	program := strings.ToLower(os.Getenv("TERM_PROGRAM"))
	if strings.Contains(term, "kitty") {
		return "kitty"
	}
	if strings.Contains(program, "iterm") {
		return "iterm2"
	}
	if strings.Contains(term, "sixel") {
		return "sixel"
	}
	return ""
}

func localFileExists(path string) bool {
	if path == "" || strings.Contains(path, "://") {
		return false
	}
	if _, err := os.Stat(filepath.Clean(path)); err == nil {
		return true
	}
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func clamp(value, low, high int) int {
	return max(low, min(value, high))
}

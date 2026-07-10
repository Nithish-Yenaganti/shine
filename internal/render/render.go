package render

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"hash/fnv"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/charmbracelet/lipgloss"

	"github.com/Nithish-Yenaganti/shine/internal/config"
	"github.com/Nithish-Yenaganti/shine/internal/model"
)

type Renderer struct {
	Width          int
	Theme          config.Theme
	SourcePath     string
	RenderImages   bool
	MermaidCommand string
	MermaidCache   string
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

func (r Renderer) WithSourcePath(path string) Renderer {
	r.SourcePath = path
	return r
}

func (r Renderer) WithImages(enabled bool) Renderer {
	r.RenderImages = enabled
	return r
}

func (r Renderer) WithMermaidCommand(command string) Renderer {
	r.MermaidCommand = command
	return r
}

func (r Renderer) WithMermaidCache(path string) Renderer {
	r.MermaidCache = path
	return r
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
		rule := strings.Repeat("━", clamp(lipgloss.Width(text), 3, r.Width/2))
		title := r.headingStyle().Bold(true).Render(text)
		return title + "\n" + r.borderStyle().Render(rule)
	case 2:
		rule := strings.Repeat("─", clamp(lipgloss.Width(text), 3, r.Width/2))
		title := r.headingStyle().Render(text)
		return title + "\n" + r.borderStyle().Render(rule)
	default:
		prefix := strings.Repeat("·", max(0, level-2)) + " "
		return r.mutedStyle().Render(prefix) + r.headingStyle().Render(text)
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
	if isMermaidLanguage(label) {
		return r.mermaid(block)
	}
	return r.codeBlock(block)
}

func (r Renderer) codeBlock(block model.Block) string {
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

func (r Renderer) codeFallback(block model.Block, note string) string {
	out := r.codeBlock(block)
	if note == "" {
		return out
	}
	return out + "\n" + r.mutedStyle().Render("mermaid  "+note)
}

func (r Renderer) mermaid(block model.Block) string {
	protocol := terminalImageProtocol()
	if !r.RenderImages {
		return r.codeFallback(block, "preview disabled in text output")
	}
	if !supportsKittyGraphics(protocol) {
		return r.codeFallback(block, "preview requires kitty-compatible graphics")
	}
	path, err := r.mermaidImage(block.Code)
	if err != nil {
		return r.codeFallback(block, err.Error())
	}
	preview, err := r.kittyImage(path)
	if err != nil {
		return r.codeFallback(block, "preview failed")
	}
	return strings.Join([]string{
		r.mutedStyle().Render("diagram  ") + r.bodyStyle().Bold(true).Render("Mermaid"),
		preview,
	}, "\n")
}

func (r Renderer) mermaidImage(code string) (string, error) {
	command := r.mermaidCommand()
	if _, err := exec.LookPath(command); err != nil {
		return "", fmt.Errorf("preview requires Mermaid CLI (mmdc)")
	}
	cacheDir, err := r.mermaidCacheDir()
	if err != nil {
		return "", fmt.Errorf("preview failed")
	}
	key := mermaidCacheKey(code)
	inputPath := filepath.Join(cacheDir, key+".mmd")
	outputPath := filepath.Join(cacheDir, key+".png")
	if stat, err := os.Stat(outputPath); err == nil && stat.Mode().IsRegular() {
		return outputPath, nil
	}
	if err := os.WriteFile(inputPath, []byte(code), 0o600); err != nil {
		return "", fmt.Errorf("preview failed")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, command, "-i", inputPath, "-o", outputPath)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("preview failed")
	}
	if stat, err := os.Stat(outputPath); err != nil || !stat.Mode().IsRegular() {
		return "", fmt.Errorf("preview failed")
	}
	return outputPath, nil
}

func (r Renderer) mermaidCommand() string {
	if r.MermaidCommand != "" {
		return r.MermaidCommand
	}
	if command := os.Getenv("SHINE_MERMAID_CMD"); command != "" {
		return command
	}
	return "mmdc"
}

func (r Renderer) mermaidCacheDir() (string, error) {
	dir := r.MermaidCache
	if dir == "" {
		base, err := os.UserCacheDir()
		if err != nil || base == "" {
			base = os.TempDir()
		}
		dir = filepath.Join(base, "shine", "mermaid")
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	return dir, nil
}

func mermaidCacheKey(code string) string {
	sum := sha256.Sum256([]byte(code))
	return hex.EncodeToString(sum[:])
}

func isMermaidLanguage(language string) bool {
	return strings.EqualFold(strings.TrimSpace(language), "mermaid")
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
	if r.Theme.Name == "github" || r.Theme.Name == "claude" {
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
	style := styles.Get(r.syntaxStyleName())
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

func (r Renderer) syntaxStyleName() string {
	switch r.Theme.Name {
	case "github", "catppuccin-latte", "claude":
		return "github"
	case "mono":
		return "monokai"
	default:
		return "github-dark"
	}
}

func (r Renderer) image(block model.Block) string {
	path, ok := r.localImagePath(block.URL)
	protocol := terminalImageProtocol()
	status := "image preview requires kitty-compatible graphics"
	if isExternalTarget(block.URL) {
		status = "image preview unavailable for remote images"
	} else if !ok {
		status = "local image file not found"
	} else if r.RenderImages && supportsKittyGraphics(protocol) {
		if preview, err := r.kittyImage(path); err == nil {
			return strings.Join([]string{
				r.mutedStyle().Render("image  ") + r.bodyStyle().Bold(true).Render(block.Alt),
				preview,
				r.mutedStyle().Render("       ") + r.linkStyle().Render(block.URL),
			}, "\n")
		}
		status = "image preview failed"
	} else if supportsKittyGraphics(protocol) {
		status = protocol + " image preview disabled in text output"
	}
	return strings.Join([]string{
		r.mutedStyle().Render("image  ") + r.bodyStyle().Bold(true).Render(block.Alt),
		r.mutedStyle().Render("       ") + r.linkStyle().Render(block.URL),
		r.mutedStyle().Render("       " + status),
	}, "\n")
}

func (r Renderer) kittyImage(path string) (string, error) {
	cols, rows, format, err := r.imageCellSize(path)
	if err != nil {
		return "", err
	}
	transferPath := path
	if format != "png" {
		transferPath, err = cachedImagePNG(path)
		if err != nil {
			return "", err
		}
	}
	imageID := kittyImageID(transferPath)
	payload := base64.StdEncoding.EncodeToString([]byte(transferPath))
	escape := fmt.Sprintf("\x1b_Ga=T,q=2,f=100,t=f,U=1,i=%d,c=%d,r=%d;%s\x1b\\", imageID, cols, rows, payload)
	return escape + kittyPlaceholder(imageID, cols, rows), nil
}

func cachedImagePNG(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	cacheDir, err := imageCacheDir()
	if err != nil {
		return "", err
	}
	cachePath := filepath.Join(cacheDir, hex.EncodeToString(hash.Sum(nil))+".png")
	if isRegularFile(cachePath) {
		return cachePath, nil
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", err
	}
	img, _, err := image.Decode(file)
	if err != nil {
		return "", err
	}
	if err := writeCachedPNG(cacheDir, cachePath, img); err != nil {
		return "", err
	}
	return cachePath, nil
}

func imageCacheDir() (string, error) {
	base, err := os.UserCacheDir()
	if err != nil || base == "" {
		return "", fmt.Errorf("image cache unavailable")
	}
	dir := filepath.Join(base, "shine", "images")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	info, err := os.Lstat(dir)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("image cache path is not a directory")
	}
	if info.Mode().Perm() != 0o700 {
		if err := os.Chmod(dir, 0o700); err != nil {
			return "", err
		}
	}
	return dir, nil
}

func writeCachedPNG(cacheDir string, cachePath string, img image.Image) error {
	temp, err := os.CreateTemp(cacheDir, ".image-*.png")
	if err != nil {
		return err
	}
	tempPath := temp.Name()
	defer os.Remove(tempPath)
	if err := png.Encode(temp, img); err != nil {
		_ = temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tempPath, cachePath); err != nil {
		if isRegularFile(cachePath) {
			return nil
		}
		return err
	}
	return nil
}

func isRegularFile(path string) bool {
	info, err := os.Lstat(path)
	return err == nil && info.Mode().IsRegular() && info.Size() > 0
}

func kittyImageID(path string) int {
	h := fnv.New32a()
	_, _ = h.Write([]byte(path))
	id := int(h.Sum32() & 0x00ffffff)
	if id == 0 {
		return 1
	}
	return id
}

func kittyPlaceholder(imageID int, cols int, rows int) string {
	if cols > len(kittyDiacritics) || rows > len(kittyDiacritics) {
		cols = min(cols, len(kittyDiacritics))
		rows = min(rows, len(kittyDiacritics))
	}
	var b strings.Builder
	red := (imageID >> 16) & 0xff
	green := (imageID >> 8) & 0xff
	blue := imageID & 0xff
	for row := 0; row < rows; row++ {
		if row > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(fmt.Sprintf("\x1b[38;2;%d;%d;%dm", red, green, blue))
		for col := 0; col < cols; col++ {
			b.WriteRune('\U0010eeee')
			b.WriteRune(kittyDiacritics[row])
			b.WriteRune(kittyDiacritics[col])
		}
		b.WriteString("\x1b[39m")
	}
	return b.String()
}

var kittyDiacritics = []rune{
	'\u0305', '\u030d', '\u030e', '\u0310', '\u0312', '\u033d', '\u033e', '\u033f',
	'\u0346', '\u034a', '\u034b', '\u034c', '\u0350', '\u0351', '\u0352', '\u0357',
	'\u035b', '\u0363', '\u0364', '\u0365', '\u0366', '\u0367', '\u0368', '\u0369',
	'\u036a', '\u036b', '\u036c', '\u036d', '\u036e', '\u036f', '\u0483', '\u0484',
	'\u0485', '\u0486', '\u0487', '\u0592', '\u0593', '\u0594', '\u0595', '\u0597',
	'\u0598', '\u0599', '\u059c', '\u059d', '\u059e', '\u059f', '\u05a0', '\u05a1',
	'\u05a8', '\u05a9', '\u05ab', '\u05ac', '\u05af', '\u05c4', '\u0610', '\u0611',
	'\u0612', '\u0613', '\u0614', '\u0615', '\u0616', '\u0617', '\u0657', '\u0658',
	'\u0659', '\u065a', '\u065b', '\u065d', '\u065e', '\u06d6', '\u06d7', '\u06d8',
}

func (r Renderer) imageCellSize(path string) (int, int, string, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, 0, "", err
	}
	defer file.Close()

	cfg, format, err := image.DecodeConfig(file)
	if err != nil {
		return 0, 0, "", err
	}
	cols := clamp(r.Width-7, 16, 72)
	if cfg.Width > 0 && cfg.Width < cols {
		cols = max(1, cfg.Width)
	}
	rows := 8
	if cfg.Width > 0 && cfg.Height > 0 {
		rows = clamp((cfg.Height*cols)/(cfg.Width*2), 1, 18)
	}
	return cols, rows, format, nil
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
	if strings.Contains(term, "ghostty") || strings.Contains(program, "ghostty") {
		return "ghostty"
	}
	if strings.Contains(term, "kitty") || os.Getenv("KITTY_WINDOW_ID") != "" {
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

func supportsKittyGraphics(protocol string) bool {
	return protocol == "kitty" || protocol == "ghostty"
}

func (r Renderer) localImagePath(target string) (string, bool) {
	path, ok := localFileTarget(r.SourcePath, target)
	if !ok {
		return "", false
	}
	stat, err := os.Stat(path)
	if err != nil || !stat.Mode().IsRegular() {
		return "", false
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return path, true
	}
	return abs, true
}

func localFileTarget(sourcePath string, target string) (string, bool) {
	if target == "" || isExternalTarget(target) || strings.HasPrefix(target, "#") {
		return "", false
	}
	path := strings.SplitN(target, "#", 2)[0]
	if path == "" {
		return "", false
	}
	path = filepath.Clean(path)
	if !filepath.IsAbs(path) && sourcePath != "" {
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

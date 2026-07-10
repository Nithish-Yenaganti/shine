package render

import (
	"encoding/base64"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/Nithish-Yenaganti/shine/internal/config"
	"github.com/Nithish-Yenaganti/shine/internal/parser"
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
	stripped := ansi.Strip(out)
	if !strings.Contains(stripped, "┌ go") || !strings.Contains(stripped, "fmt.Println") {
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

func TestAsciiCodeBlockDoesNotShowLineNumbers(t *testing.T) {
	doc, err := parser.Parse([]byte("```ascii\n █████\n░░███\n```\n"), "test.md")
	if err != nil {
		t.Fatal(err)
	}
	out := New(48, config.ThemeByName("mono")).Render(doc)
	if strings.Contains(out, "│ 1 ") || strings.Contains(out, "│ 2 ") {
		t.Fatalf("ascii block should not show line numbers:\n%s", out)
	}
	if !strings.Contains(out, "█████") || !strings.Contains(out, "░░███") {
		t.Fatalf("missing ascii art:\n%s", out)
	}
}

func TestGitHubCodeUsesReadableRawLines(t *testing.T) {
	r := New(48, config.ThemeByName("github"))
	lines := r.codeLines("fn main() {\n    fmt.Println(\"hi\")\n}", "go")
	got := strings.Join(lines, "\n")
	if strings.Contains(got, "\x1b[") {
		t.Fatalf("github code should not contain nested ANSI highlighting: %q", got)
	}
	if !strings.Contains(got, "fmt.Println") {
		t.Fatalf("missing code content: %q", got)
	}
}

func TestClaudeCodeUsesThemeReadableRawLines(t *testing.T) {
	r := New(48, config.ThemeByName("claude"))
	lines := r.codeLines("fn main() {\n    fmt.Println(\"hi\")\n}", "go")
	got := strings.Join(lines, "\n")
	if strings.Contains(got, "\x1b[") {
		t.Fatalf("claude code should not contain nested ANSI highlighting: %q", got)
	}
	if !strings.Contains(got, "fmt.Println") {
		t.Fatalf("missing code content: %q", got)
	}
}

func TestLightThemesUseLightSyntaxStyle(t *testing.T) {
	for _, name := range []string{"github", "catppuccin-latte", "claude"} {
		r := New(48, config.ThemeByName(name))
		if got := r.syntaxStyleName(); got != "github" {
			t.Fatalf("%s syntax style = %q, want github", name, got)
		}
	}
}

func TestDarkThemesUseDarkSyntaxStyle(t *testing.T) {
	for _, name := range []string{"tomorrow-night", "catppuccin-mocha", "everforest", "jellybeans", "gotham"} {
		r := New(48, config.ThemeByName(name))
		if got := r.syntaxStyleName(); got != "github-dark" {
			t.Fatalf("%s syntax style = %q, want github-dark", name, got)
		}
	}
}

func TestMonoUsesMonokaiSyntaxStyle(t *testing.T) {
	r := New(48, config.ThemeByName("mono"))
	if got := r.syntaxStyleName(); got != "monokai" {
		t.Fatalf("mono syntax style = %q, want monokai", got)
	}
	lines := r.codeLines("go install github.com/Nithish-Yenaganti/shine/cmd/shine@latest", "sh")
	if got := strings.Join(lines, "\n"); !strings.Contains(got, "\x1b[") {
		t.Fatalf("mono code should use syntax highlighting, got %q", got)
	}
}

func TestGitHubDoesNotPaintTokenBackgrounds(t *testing.T) {
	r := New(48, config.ThemeByName("github"))
	bg := r.style().GetBackground()
	if _, ok := bg.(lipgloss.NoColor); !ok {
		t.Fatalf("github renderer should not paint token backgrounds, got %v", bg)
	}
}

func TestInlineRelativePathKeepsSpace(t *testing.T) {
	doc, err := parser.Parse([]byte("- `go test ./...`\n- `go build ./cmd/shine`\n"), "test.md")
	if err != nil {
		t.Fatal(err)
	}
	out := New(64, config.ThemeByName("mono")).Render(doc)
	if strings.Contains(out, "test./...") || strings.Contains(out, "build./cmd") {
		t.Fatalf("relative paths should not stick to previous words:\n%s", out)
	}
	if !strings.Contains(out, "go test ./...") || !strings.Contains(out, "go build ./cmd/shine") {
		t.Fatalf("missing spaced relative path commands:\n%s", out)
	}
}

func TestMermaidFallsBackWhenImagesDisabled(t *testing.T) {
	doc, err := parser.Parse([]byte("```mermaid\ngraph TD\n  A --> B\n```\n"), "test.md")
	if err != nil {
		t.Fatal(err)
	}

	out := New(48, config.ThemeByName("mono")).Render(doc)

	if !strings.Contains(out, "┌ mermaid") || !strings.Contains(out, "A --> B") {
		t.Fatalf("expected mermaid source code fallback:\n%s", out)
	}
	if !strings.Contains(out, "mermaid  preview disabled in text output") {
		t.Fatalf("missing disabled mermaid fallback note:\n%s", out)
	}
	if strings.Contains(out, "\x1b_G") {
		t.Fatalf("disabled mermaid preview should not emit image escapes:\n%q", out)
	}
}

func TestMermaidFallsBackWhenCommandMissing(t *testing.T) {
	t.Setenv("TERM", "xterm-kitty")
	t.Setenv("TERM_PROGRAM", "")
	doc, err := parser.Parse([]byte("```mermaid\ngraph TD\n  A --> B\n```\n"), "test.md")
	if err != nil {
		t.Fatal(err)
	}

	out := New(48, config.ThemeByName("mono")).
		WithImages(true).
		WithMermaidCommand(filepath.Join(t.TempDir(), "missing-mmdc")).
		Render(doc)

	if !strings.Contains(out, "preview requires Mermaid CLI (mmdc)") {
		t.Fatalf("missing mermaid CLI fallback note:\n%s", out)
	}
	if !strings.Contains(out, "A --> B") {
		t.Fatalf("expected mermaid source to remain visible:\n%s", out)
	}
}

func TestMermaidRendersCachedImageWithKittyGraphics(t *testing.T) {
	t.Setenv("TERM", "xterm-kitty")
	t.Setenv("TERM_PROGRAM", "")
	dir := t.TempDir()
	cacheDir := filepath.Join(dir, "cache")
	imagePath := filepath.Join(dir, "rendered.png")
	writeTestPNG(t, imagePath)
	command := fakeMermaidCommand(t, imagePath)
	doc, err := parser.Parse([]byte("```mermaid\ngraph TD\n  A --> B\n```\n"), "test.md")
	if err != nil {
		t.Fatal(err)
	}

	out := New(48, config.ThemeByName("mono")).
		WithImages(true).
		WithMermaidCommand(command).
		WithMermaidCache(cacheDir).
		Render(doc)

	if !strings.Contains(out, "diagram") || !strings.Contains(out, "Mermaid") {
		t.Fatalf("missing mermaid diagram label:\n%q", out)
	}
	if !strings.Contains(out, "\x1b_Ga=T,q=2,f=100,t=f,U=1,i=") {
		t.Fatalf("missing kitty image escape for rendered mermaid:\n%q", out)
	}
	if strings.Contains(out, "A --> B") || strings.Contains(out, "preview failed") {
		t.Fatalf("successful mermaid render should not show source fallback:\n%q", out)
	}
	cached := filepath.Join(cacheDir, mermaidCacheKey("graph TD\n  A --> B")+".png")
	if _, err := os.Stat(cached); err != nil {
		t.Fatalf("expected cached mermaid image at %s: %v", cached, err)
	}
}

func TestLocalImagePathResolvesRelativeToSourcePath(t *testing.T) {
	dir := t.TempDir()
	docsDir := filepath.Join(dir, "docs")
	assetsDir := filepath.Join(docsDir, "assets")
	if err := os.MkdirAll(assetsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	imagePath := filepath.Join(assetsDir, "logo.png")
	writeTestPNG(t, imagePath)

	got, ok := New(48, config.ThemeByName("mono")).
		WithSourcePath(filepath.Join(docsDir, "README.md")).
		localImagePath("assets/logo.png")
	if !ok {
		t.Fatalf("expected image path to resolve")
	}
	if got != imagePath {
		t.Fatalf("resolved path = %q, want %q", got, imagePath)
	}
}

func TestKittyImageRenderingEmitsGraphicsEscape(t *testing.T) {
	t.Setenv("TERM", "xterm-kitty")
	t.Setenv("TERM_PROGRAM", "")
	dir := t.TempDir()
	imagePath := filepath.Join(dir, "logo.png")
	writeTestPNG(t, imagePath)
	doc, err := parser.Parse([]byte("![Logo](logo.png)\n"), filepath.Join(dir, "README.md"))
	if err != nil {
		t.Fatal(err)
	}

	out := New(48, config.ThemeByName("mono")).
		WithSourcePath(filepath.Join(dir, "README.md")).
		WithImages(true).
		Render(doc)

	if !strings.Contains(out, "\x1b_Ga=T,q=2,f=100,t=f,U=1,i=") {
		t.Fatalf("missing kitty virtual placement escape:\n%q", out)
	}
	if !strings.Contains(out, base64.StdEncoding.EncodeToString([]byte(imagePath))) {
		t.Fatalf("kitty escape does not include encoded image path:\n%q", out)
	}
	if !strings.Contains(out, "\U0010eeee") {
		t.Fatalf("kitty output does not include unicode placeholder cells:\n%q", out)
	}
	if !strings.Contains(out, "\U0010eeee\u0305\u0305") || !strings.Contains(out, "\U0010eeee\u0305\u030d") {
		t.Fatalf("kitty placeholders should encode row and column diacritics:\n%q", out)
	}
	assertPlaceholderColorMatchesImageID(t, out)
	if strings.Contains(out, "image preview unavailable") || strings.Contains(out, "kitty image preview disabled") {
		t.Fatalf("expected real image rendering, got fallback:\n%q", out)
	}
}

func TestGhosttyUsesKittyGraphicsProtocol(t *testing.T) {
	t.Setenv("TERM", "xterm-ghostty")
	t.Setenv("TERM_PROGRAM", "ghostty")
	if got := terminalImageProtocol(); got != "ghostty" {
		t.Fatalf("expected ghostty protocol, got %q", got)
	}
}

func TestGhosttyImageRenderingUsesUnicodePlaceholderPlacement(t *testing.T) {
	t.Setenv("TERM", "xterm-ghostty")
	t.Setenv("TERM_PROGRAM", "ghostty")
	dir := t.TempDir()
	imagePath := filepath.Join(dir, "logo.png")
	writeTestPNG(t, imagePath)
	doc, err := parser.Parse([]byte("![Logo](logo.png)\n"), filepath.Join(dir, "README.md"))
	if err != nil {
		t.Fatal(err)
	}

	out := New(48, config.ThemeByName("mono")).
		WithSourcePath(filepath.Join(dir, "README.md")).
		WithImages(true).
		Render(doc)

	if !strings.Contains(out, "\x1b_Ga=T,q=2,f=100,t=f,U=1,i=") {
		t.Fatalf("missing virtual kitty graphics escape for ghostty:\n%q", out)
	}
	if !strings.Contains(out, "\U0010eeee\u0305\u0305") {
		t.Fatalf("ghostty should use unicode placeholders so images scroll with the viewport:\n%q", out)
	}
	assertPlaceholderColorMatchesImageID(t, out)
	if strings.Contains(out, "image preview unavailable") || strings.Contains(out, "image preview failed") {
		t.Fatalf("expected ghostty image rendering, got fallback:\n%q", out)
	}
}

func TestKittyImageRenderingSupportsDecodedLocalImages(t *testing.T) {
	t.Setenv("TERM", "xterm-kitty")
	t.Setenv("TERM_PROGRAM", "")
	dir := t.TempDir()
	imagePath := filepath.Join(dir, "photo.jpg")
	writeTestJPEG(t, imagePath)
	doc, err := parser.Parse([]byte("![Photo](photo.jpg)\n"), filepath.Join(dir, "README.md"))
	if err != nil {
		t.Fatal(err)
	}

	out := New(48, config.ThemeByName("mono")).
		WithSourcePath(filepath.Join(dir, "README.md")).
		WithImages(true).
		Render(doc)

	if !strings.Contains(out, "\x1b_Ga=T,q=2,f=100,U=1,i=") {
		t.Fatalf("missing kitty graphics payload escape:\n%q", out)
	}
	if strings.Contains(out, "t=f") {
		t.Fatalf("decoded non-PNG image should use direct payload transfer, got file transfer:\n%q", out)
	}
	if strings.Contains(out, "image preview unavailable") || strings.Contains(out, "image preview failed") {
		t.Fatalf("expected decoded image rendering, got fallback:\n%q", out)
	}
}

func TestKittyPlaceholderSurvivesViewportRendering(t *testing.T) {
	t.Setenv("TERM", "xterm-kitty")
	t.Setenv("TERM_PROGRAM", "")
	dir := t.TempDir()
	imagePath := filepath.Join(dir, "logo.png")
	writeTestPNG(t, imagePath)
	doc, err := parser.Parse([]byte("![Logo](logo.png)\n"), filepath.Join(dir, "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	out := New(48, config.ThemeByName("mono")).
		WithSourcePath(filepath.Join(dir, "README.md")).
		WithImages(true).
		Render(doc)

	vp := viewport.New(48, 8)
	vp.SetContent(out)
	view := vp.View()
	if !strings.Contains(view, "\U0010eeee") {
		t.Fatalf("viewport stripped kitty placeholder cells:\n%q", view)
	}
	if !strings.Contains(view, "\U0010eeee\u0305\u0305") {
		t.Fatalf("viewport stripped kitty row/column diacritics:\n%q", view)
	}
	if !strings.Contains(view, "\x1b_G") {
		t.Fatalf("viewport stripped kitty image transmission escape:\n%q", view)
	}
}

func TestImageRenderingFallsBackWhenDisabled(t *testing.T) {
	t.Setenv("TERM", "xterm-kitty")
	t.Setenv("TERM_PROGRAM", "")
	dir := t.TempDir()
	imagePath := filepath.Join(dir, "logo.png")
	writeTestPNG(t, imagePath)
	doc, err := parser.Parse([]byte("![Logo](logo.png)\n"), filepath.Join(dir, "README.md"))
	if err != nil {
		t.Fatal(err)
	}

	out := New(48, config.ThemeByName("mono")).
		WithSourcePath(filepath.Join(dir, "README.md")).
		Render(doc)

	if strings.Contains(out, "\x1b_G") {
		t.Fatalf("default renderer should not emit kitty graphics escapes:\n%q", out)
	}
	if !strings.Contains(out, "kitty image preview disabled in text output") {
		t.Fatalf("missing disabled fallback message:\n%s", out)
	}
}

func TestImageRenderingFallsBackInUnsupportedTerminal(t *testing.T) {
	t.Setenv("TERM", "xterm-256color")
	t.Setenv("TERM_PROGRAM", "")
	dir := t.TempDir()
	imagePath := filepath.Join(dir, "logo.png")
	writeTestPNG(t, imagePath)
	doc, err := parser.Parse([]byte("![Logo](logo.png)\n"), filepath.Join(dir, "README.md"))
	if err != nil {
		t.Fatal(err)
	}

	out := New(48, config.ThemeByName("mono")).
		WithSourcePath(filepath.Join(dir, "README.md")).
		WithImages(true).
		Render(doc)

	if strings.Contains(out, "\x1b_G") {
		t.Fatalf("unsupported terminal should not emit kitty graphics escapes:\n%q", out)
	}
	if !strings.Contains(out, "image preview requires kitty-compatible graphics") {
		t.Fatalf("missing unsupported terminal fallback:\n%s", out)
	}
}

func TestKittyImageIDUsesFullHashRange(t *testing.T) {
	idA := kittyImageID("/tmp/a.png")
	idB := kittyImageID("/tmp/b.png")
	if idA == idB {
		t.Fatalf("expected different paths to produce different ids")
	}
	if idA <= 255 || idB <= 255 {
		t.Fatalf("expected image ids to use more than an 8-bit color range, got %d and %d", idA, idB)
	}
	if idA > 0x00ffffff || idB > 0x00ffffff {
		t.Fatalf("expected image ids to fit in truecolor placeholder range, got %d and %d", idA, idB)
	}
}

func TestImageRenderingFallsBackForRemoteAndMissingImages(t *testing.T) {
	t.Setenv("TERM", "xterm-kitty")
	doc, err := parser.Parse([]byte("![Remote](https://example.com/logo.png)\n\n![Missing](missing.png)\n"), "README.md")
	if err != nil {
		t.Fatal(err)
	}

	out := New(48, config.ThemeByName("mono")).
		WithSourcePath("README.md").
		WithImages(true).
		Render(doc)

	if strings.Contains(out, "\x1b_G") {
		t.Fatalf("remote or missing images should not emit kitty graphics escapes:\n%q", out)
	}
	if !strings.Contains(out, "image preview unavailable for remote images") {
		t.Fatalf("missing remote image fallback:\n%s", out)
	}
	if !strings.Contains(out, "local image file not found") {
		t.Fatalf("missing local missing-image fallback:\n%s", out)
	}
}

func writeTestPNG(t *testing.T, path string) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 4, 2))
	for y := 0; y < 2; y++ {
		for x := 0; x < 4; x++ {
			img.Set(x, y, color.RGBA{R: 220, G: 80, B: 40, A: 255})
		}
	}
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	if err := png.Encode(file, img); err != nil {
		t.Fatal(err)
	}
}

func assertPlaceholderColorMatchesImageID(t *testing.T, out string) {
	t.Helper()
	idMatch := regexp.MustCompile(`i=(\d+)`).FindStringSubmatch(out)
	if len(idMatch) != 2 {
		t.Fatalf("missing image id in output:\n%q", out)
	}
	id, err := strconv.Atoi(idMatch[1])
	if err != nil {
		t.Fatal(err)
	}
	colorMatch := regexp.MustCompile(`\x1b\[38;2;(\d+);(\d+);(\d+)m`).FindStringSubmatch(out)
	if len(colorMatch) != 4 {
		t.Fatalf("missing truecolor placeholder foreground:\n%q", out)
	}
	red, _ := strconv.Atoi(colorMatch[1])
	green, _ := strconv.Atoi(colorMatch[2])
	blue, _ := strconv.Atoi(colorMatch[3])
	got := red<<16 | green<<8 | blue
	if got != id {
		t.Fatalf("placeholder color encodes image id %d, want %d", got, id)
	}
}

func writeTestJPEG(t *testing.T, path string) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 4, 2))
	for y := 0; y < 2; y++ {
		for x := 0; x < 4; x++ {
			img.Set(x, y, color.RGBA{R: 80, G: 120, B: 220, A: 255})
		}
	}
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	if err := jpeg.Encode(file, img, nil); err != nil {
		t.Fatal(err)
	}
}

func fakeMermaidCommand(t *testing.T, renderedImage string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "mmdc")
	script := `#!/bin/sh
out=""
while [ "$#" -gt 0 ]; do
  if [ "$1" = "-o" ]; then
    shift
    out="$1"
  fi
  shift
done
cp "$SHINE_TEST_MERMAID_IMAGE" "$out"
`
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("SHINE_TEST_MERMAID_IMAGE", renderedImage)
	return path
}

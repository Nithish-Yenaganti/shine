package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/Nithish-Yenaganti/shine/internal/config"
	"github.com/Nithish-Yenaganti/shine/internal/source"
)

func TestBackgroundScreenDoesNotPadRows(t *testing.T) {
	m := model{
		theme:  config.ThemeByName("github"),
		width:  80,
		height: 4,
	}

	screen := m.screen("hello", "status")
	lines := strings.Split(screen, "\n")
	if len(lines) != 2 {
		t.Fatalf("expected screen to render body and status only, got %d rows", len(lines))
	}
	if got := lipgloss.Width(lines[0]); got != 5 {
		t.Fatalf("expected body row to avoid padding, got width %d for %q", got, lines[0])
	}
	if strings.Contains(screen, "\x1b[") {
		t.Fatalf("background screen should not paint ANSI backgrounds: %q", screen)
	}
}

func TestContentWidthReservesLeftAndRightPadding(t *testing.T) {
	tests := []struct {
		width int
		want  int
	}{
		{60, 57},
		{88, 54},
		{140, 84},
	}
	for _, tt := range tests {
		m := model{width: tt.width}
		if got := m.contentWidth(); got != tt.want {
			t.Fatalf("contentWidth(%d) = %d, want %d", tt.width, got, tt.want)
		}
	}
}

func TestPaddedBodyAddsLeftPaddingOnly(t *testing.T) {
	m := model{width: 88}
	got := m.paddedBody("one\ntwo")
	prefix := strings.Repeat(" ", 17)
	if got != prefix+"one\n"+prefix+"two" {
		t.Fatalf("unexpected padded body: %q", got)
	}
}

func TestPaddedBodyDoesNotCenterStructuralLines(t *testing.T) {
	m := model{width: 88}
	input := strings.Join([]string{
		"┌ go ───",
		"│ fmt.Println",
		"• item",
		"┃ quote",
	}, "\n")
	got := m.paddedBody(input)
	prefix := strings.Repeat(" ", 17)
	want := prefix + strings.ReplaceAll(input, "\n", "\n"+prefix)
	if got != want {
		t.Fatalf("unexpected structural padding:\n%q\nwant:\n%q", got, want)
	}
}

func TestReloadRenderLeavesHeadingsAndBodyLeftAligned(t *testing.T) {
	m := model{
		width: 88,
		theme: config.ThemeByName("mono"),
		source: source.Source{
			Name:    "test.md",
			Content: "# Shine\n\nBody text\n\n- item",
		},
	}

	m.reloadRender()
	stripped := ansi.Strip(m.content)
	lines := strings.Split(stripped, "\n")
	if len(lines) < 6 {
		t.Fatalf("expected rendered heading, paragraph, and list, got:\n%q", stripped)
	}
	if lines[0] != "Shine" {
		t.Fatalf("expected heading title to stay left aligned, got %q", lines[0])
	}
	if lines[1] != "━━━━━" {
		t.Fatalf("expected heading rule to stay left aligned, got %q", lines[1])
	}
	if lines[3] != "Body text" {
		t.Fatalf("expected paragraph to stay left aligned, got %q", lines[3])
	}
	if lines[5] != "• item" {
		t.Fatalf("expected list item to stay left aligned, got %q", lines[5])
	}
}

func TestStatusLineUsesCachedTotalLines(t *testing.T) {
	m := model{
		theme:      config.ThemeByName("mono"),
		source:     sourceForTest("test.md"),
		viewport:   viewport.New(20, 4),
		content:    "one\ntwo",
		totalLines: 42,
	}
	m.viewport.YOffset = 2

	status := m.statusLine()
	if !strings.Contains(status, "line:3/42") {
		t.Fatalf("expected cached line count in status, got %q", status)
	}
}

func TestKeyboardShortcutsScrollViewport(t *testing.T) {
	m := model{
		theme:    config.ThemeByName("mono"),
		viewport: newViewport(20, 4),
		content:  strings.Join([]string{"one", "two", "three", "four", "five", "six", "seven", "eight"}, "\n"),
	}
	m.viewport.SetContent(m.content)

	m = updateKey(t, m, "j")
	if m.viewport.YOffset != 1 {
		t.Fatalf("expected j to scroll down to 1, got %d", m.viewport.YOffset)
	}

	m = updateKey(t, m, "k")
	if m.viewport.YOffset != 0 {
		t.Fatalf("expected k to scroll up to 0, got %d", m.viewport.YOffset)
	}

	m = updateKey(t, m, "d")
	if m.viewport.YOffset == 0 {
		t.Fatalf("expected d to scroll down by a half page")
	}

	m = updateKey(t, m, "u")
	if m.viewport.YOffset != 0 {
		t.Fatalf("expected u to scroll back up to 0, got %d", m.viewport.YOffset)
	}
}

func TestThemeShortcutWorksAfterScrolling(t *testing.T) {
	m := model{
		theme:    config.ThemeByName("mono"),
		viewport: newViewport(20, 4),
		content:  strings.Join([]string{"one", "two", "three", "four", "five", "six", "seven", "eight"}, "\n"),
	}
	m.viewport.SetContent(m.content)

	m = updateKey(t, m, "j")
	m = updateKey(t, m, "T")
	if !m.themeMenu {
		t.Fatalf("expected uppercase T to open theme menu after scrolling")
	}
}

func TestMouseWheelScrollsViewportWithTunedDelta(t *testing.T) {
	m := model{
		theme:    config.ThemeByName("mono"),
		viewport: newViewport(20, 4),
		content:  strings.Join([]string{"one", "two", "three", "four", "five", "six", "seven", "eight", "nine", "ten", "eleven", "twelve"}, "\n"),
	}
	m.viewport.SetContent(m.content)

	m = updateMouse(t, m, tea.MouseMsg{
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonWheelDown,
		Type:   tea.MouseWheelDown,
	})
	if m.viewport.YOffset < 5 {
		t.Fatalf("expected wheel down to scroll by tuned delta, got offset %d", m.viewport.YOffset)
	}

	m = updateMouse(t, m, tea.MouseMsg{
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonWheelUp,
		Type:   tea.MouseWheelUp,
	})
	if m.viewport.YOffset != 0 {
		t.Fatalf("expected wheel up to return to top, got offset %d", m.viewport.YOffset)
	}
}

func TestKeyboardShortcutsJumpTopBottom(t *testing.T) {
	m := model{
		theme:    config.ThemeByName("mono"),
		viewport: viewport.New(20, 4),
		content:  strings.Join([]string{"one", "two", "three", "four", "five", "six", "seven", "eight"}, "\n"),
	}
	m.viewport.SetContent(m.content)

	m = updateKey(t, m, "G")
	if m.viewport.YOffset == 0 {
		t.Fatalf("expected G to jump to bottom")
	}

	m = updateKey(t, m, "g")
	if m.viewport.YOffset != 0 {
		t.Fatalf("expected g to jump to top, got %d", m.viewport.YOffset)
	}
}

func TestKeyboardShortcutsToggleHelp(t *testing.T) {
	m := model{
		theme:    config.ThemeByName("mono"),
		viewport: viewport.New(20, 4),
	}

	m = updateKey(t, m, "?")
	if !m.help {
		t.Fatalf("expected ? to open help")
	}

	m = updateKey(t, m, "?")
	if m.help {
		t.Fatalf("expected ? to close help")
	}

	m = updateKey(t, m, "H")
	if !m.help {
		t.Fatalf("expected H to open help")
	}

	m = updateKey(t, m, "h")
	if !m.help {
		t.Fatalf("expected h to keep help open")
	}
}

func TestShowKeysStartsWithHelpOpen(t *testing.T) {
	m := model{
		theme:    config.ThemeByName("mono"),
		viewport: viewport.New(20, 4),
		help:     true,
	}

	if !strings.Contains(m.View(), "keys") {
		t.Fatalf("expected help panel to render when help starts open")
	}
}

func TestDebugKeysShowsLastKey(t *testing.T) {
	m := model{
		theme:     config.ThemeByName("mono"),
		viewport:  viewport.New(20, 4),
		debugKeys: true,
	}

	m = updateKey(t, m, "j")
	if !strings.Contains(m.statusLine(), "key:j") {
		t.Fatalf("expected status line to show last key, got %q", m.statusLine())
	}
}

func TestSearchHighlightSkipsImageProtocolLines(t *testing.T) {
	imageLine := "\x1b_Ga=T,q=2;payload\x1b\\\x1b[38;5;42m\U0010eeee\u0305\u0305\x1b[39m"
	m := model{
		theme:   config.ThemeByName("mono"),
		content: "title\n" + imageLine + "\nlogo.png",
	}

	got := m.highlightedContent("payload")
	if !strings.Contains(got, imageLine) {
		t.Fatalf("image protocol line should remain unchanged:\n%q", got)
	}
	if strings.Contains(got, "\x1b[48;5") || strings.Contains(got, "#1f2328") {
		t.Fatalf("highlight style should not be applied to image protocol line:\n%q", got)
	}
}

func TestThemeMenuAppliesSelectedTheme(t *testing.T) {
	m := model{
		theme:    config.ThemeByName("tomorrow-night"),
		viewport: viewport.New(40, 8),
		content:  "# title\n",
	}
	m.viewport.SetContent(m.content)

	m = updateKey(t, m, "t")
	if !m.themeMenu {
		t.Fatalf("expected t to open theme menu")
	}
	view := m.View()
	if !strings.Contains(view, "themes") || !strings.Contains(view, "Tomorrow Night") || !strings.Contains(view, "GitHub Light") {
		t.Fatalf("expected theme menu view to render display names, got:\n%s", view)
	}

	m = updateKey(t, m, "j")
	m = updateKeyMsg(t, m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.themeMenu {
		t.Fatalf("expected enter to close theme menu")
	}
	if m.theme.Name != "github" {
		t.Fatalf("expected selected theme github, got %q", m.theme.Name)
	}
	if !strings.Contains(m.statusLine(), "theme:GitHub Light") {
		t.Fatalf("expected status line to show display name, got %q", m.statusLine())
	}
}

func updateKey(t *testing.T, m model, key string) model {
	t.Helper()
	return updateKeyMsg(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
}

func updateKeyMsg(t *testing.T, m model, msg tea.KeyMsg) model {
	t.Helper()
	updated, _ := m.Update(msg)
	got, ok := updated.(model)
	if !ok {
		t.Fatalf("expected model update, got %T", updated)
	}
	return got
}

func updateMouse(t *testing.T, m model, msg tea.MouseMsg) model {
	t.Helper()
	updated, _ := m.Update(msg)
	got, ok := updated.(model)
	if !ok {
		t.Fatalf("expected model update, got %T", updated)
	}
	return got
}

func sourceForTest(name string) source.Source {
	return source.Source{Name: name}
}

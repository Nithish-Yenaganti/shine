package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/Nithish-Yenaganti/shine/internal/config"
	"github.com/Nithish-Yenaganti/shine/internal/source"
)

func TestBackgroundScreenDoesNotPadRows(t *testing.T) {
	m := model{
		theme:  config.ThemeByName("daylight"),
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
		viewport: viewport.New(20, 4),
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

func TestThemeMenuAppliesSelectedTheme(t *testing.T) {
	m := model{
		theme:    config.ThemeByName("midnight"),
		viewport: viewport.New(40, 8),
		content:  "# title\n",
	}
	m.viewport.SetContent(m.content)

	m = updateKey(t, m, "t")
	if !m.themeMenu {
		t.Fatalf("expected t to open theme menu")
	}
	if !strings.Contains(m.View(), "themes") {
		t.Fatalf("expected theme menu view to render")
	}

	m = updateKey(t, m, "j")
	m = updateKeyMsg(t, m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.themeMenu {
		t.Fatalf("expected enter to close theme menu")
	}
	if m.theme.Name != "daylight" {
		t.Fatalf("expected selected theme daylight, got %q", m.theme.Name)
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

func sourceForTest(name string) source.Source {
	return source.Source{Name: name}
}

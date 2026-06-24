package tui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fsnotify/fsnotify"

	"shine/internal/config"
	shinemodel "shine/internal/model"
	"shine/internal/parser"
	"shine/internal/render"
	"shine/internal/source"
)

type Options struct {
	Source       source.Source
	Theme        config.Theme
	Watch        bool
	UseAltScreen bool
}

type model struct {
	source     source.Source
	theme      config.Theme
	watch      bool
	viewport   viewport.Model
	search     textinput.Model
	searching  bool
	outline    bool
	help       bool
	width      int
	height     int
	content    string
	headings   []shinemodel.HeadingRef
	matches    []int
	matchIndex int
	status     string
	err        error
}

type fileChanged struct{}

func Run(opts Options) error {
	doc, err := parser.Parse([]byte(opts.Source.Content), opts.Source.Name)
	if err != nil {
		return err
	}
	input := textinput.New()
	input.Prompt = "/"
	input.CharLimit = 128

	m := model{
		source:   opts.Source,
		theme:    opts.Theme,
		watch:    opts.Watch,
		viewport: viewport.New(88, 24),
		search:   input,
		content:  render.New(88, opts.Theme).Render(doc),
		status:   "shine",
	}
	m.headings = doc.Headings
	programOptions := []tea.ProgramOption{}
	if opts.UseAltScreen {
		programOptions = append(programOptions, tea.WithAltScreen())
	}
	_, err = tea.NewProgram(m, programOptions...).Run()
	return err
}

func (m model) Init() tea.Cmd {
	m.viewport.SetContent(m.content)
	if m.watch && m.source.Path != "" {
		return watchFile(m.source.Path)
	}
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width
		m.viewport.Height = max(1, msg.Height-1)
		m.reloadRender()
	case fileChanged:
		m.reloadFile()
		if m.watch && m.source.Path != "" {
			return m, watchFile(m.source.Path)
		}
	case tea.KeyMsg:
		if m.searching {
			switch msg.String() {
			case "enter":
				m.searching = false
				m.applySearch()
				return m, nil
			case "esc":
				m.searching = false
				m.search.SetValue("")
				return m, nil
			}
			var cmd tea.Cmd
			m.search, cmd = m.search.Update(msg)
			return m, cmd
		}
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "/":
			m.searching = true
			m.search.Focus()
			return m, nil
		case "?":
			m.help = !m.help
			return m, nil
		case "n":
			m.nextMatch()
			return m, nil
		case "N":
			m.previousMatch()
			return m, nil
		case "o":
			m.outline = !m.outline
			return m, nil
		case "r":
			m.reloadFile()
			return m, nil
		case "g":
			m.viewport.GotoTop()
			return m, nil
		case "G":
			m.viewport.GotoBottom()
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m model) View() string {
	body := m.viewport.View()
	if m.outline {
		body = m.outlineView() + "\n\n" + body
	}
	if m.help {
		body = m.helpView() + "\n\n" + body
	}
	status := m.statusLine()
	if m.searching {
		status = m.search.View()
	}
	return body + "\n" + status
}

func (m *model) reloadFile() {
	if m.source.Path == "" {
		return
	}
	content, err := os.ReadFile(m.source.Path)
	if err != nil {
		m.err = err
		return
	}
	m.source.Content = string(content)
	m.reloadRender()
	m.status = "reloaded " + m.source.Name
}

func (m *model) reloadRender() {
	doc, err := parser.Parse([]byte(m.source.Content), m.source.Name)
	if err != nil {
		m.err = err
		return
	}
	width := m.width
	if width == 0 {
		width = 88
	}
	m.content = render.New(width, m.theme).Render(doc)
	m.headings = doc.Headings
	m.applySearch()
}

func (m *model) applySearch() {
	query := strings.ToLower(strings.TrimSpace(m.search.Value()))
	m.matches = nil
	if query == "" {
		m.viewport.SetContent(m.content)
		return
	}
	for i, line := range strings.Split(m.content, "\n") {
		if strings.Contains(strings.ToLower(line), query) {
			m.matches = append(m.matches, i)
		}
	}
	if len(m.matches) > 0 {
		m.matchIndex = 0
		m.viewport.SetYOffset(m.matches[0])
	}
	m.viewport.SetContent(m.highlightedContent(query))
}

func (m *model) nextMatch() {
	if len(m.matches) == 0 {
		m.applySearch()
	}
	if len(m.matches) == 0 {
		return
	}
	m.matchIndex = (m.matchIndex + 1) % len(m.matches)
	m.viewport.SetYOffset(m.matches[m.matchIndex])
}

func (m *model) previousMatch() {
	if len(m.matches) == 0 {
		m.applySearch()
	}
	if len(m.matches) == 0 {
		return
	}
	m.matchIndex--
	if m.matchIndex < 0 {
		m.matchIndex = len(m.matches) - 1
	}
	m.viewport.SetYOffset(m.matches[m.matchIndex])
}

func (m model) outlineView() string {
	var lines []string
	title := lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.Heading)).Bold(true).Render("outline")
	lines = append(lines, title)
	for _, h := range m.headings {
		lines = append(lines, strings.Repeat("  ", max(0, h.Level-1))+h.Text)
	}
	return m.panelStyle().Render(strings.Join(lines, "\n"))
}

func (m model) helpView() string {
	help := strings.Join([]string{
		lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.Heading)).Bold(true).Render("keys"),
		"q quit    j/k scroll    g/G top/bottom",
		"/ search  n/N matches   o outline",
		"r reload  ? help",
	}, "\n")
	return m.panelStyle().Render(help)
}

func (m model) statusLine() string {
	if m.err != nil {
		return lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.CalloutWarning)).Render("error: " + m.err.Error())
	}
	watch := "watch:off"
	if m.watch {
		watch = "watch:on"
	}
	totalLines := max(1, len(strings.Split(m.content, "\n")))
	currentLine := min(totalLines, m.viewport.YOffset+1)
	search := "matches:0"
	if len(m.matches) > 0 {
		search = fmt.Sprintf("match:%d/%d", m.matchIndex+1, len(m.matches))
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.Muted)).Render(
		fmt.Sprintf("%s | %s | line:%d/%d | %s | ? keys", m.source.Name, watch, currentLine, totalLines, search),
	)
}

func (m model) highlightedContent(query string) string {
	lines := strings.Split(m.content, "\n")
	style := lipgloss.NewStyle().
		Background(lipgloss.Color(m.theme.MatchHighlight)).
		Foreground(lipgloss.Color("#1f2328"))
	for i, line := range lines {
		if strings.Contains(strings.ToLower(line), query) {
			lines[i] = style.Render(line)
		}
	}
	return strings.Join(lines, "\n")
}

func (m model) panelStyle() lipgloss.Style {
	width := m.width
	if width == 0 {
		width = 88
	}
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(m.theme.Border)).
		Padding(0, 1).
		Width(max(20, min(width-2, 42)))
}

func watchFile(path string) tea.Cmd {
	return func() tea.Msg {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			return fileChanged{}
		}
		defer watcher.Close()
		_ = watcher.Add(path)
		timer := time.NewTimer(30 * time.Second)
		defer timer.Stop()
		for {
			select {
			case event := <-watcher.Events:
				if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
					return fileChanged{}
				}
			case <-watcher.Errors:
				return fileChanged{}
			case <-timer.C:
				return fileChanged{}
			}
		}
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

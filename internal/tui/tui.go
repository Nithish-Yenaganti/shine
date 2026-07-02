package tui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/fsnotify/fsnotify"

	"github.com/Nithish-Yenaganti/shine/internal/config"
	shinemodel "github.com/Nithish-Yenaganti/shine/internal/model"
	"github.com/Nithish-Yenaganti/shine/internal/parser"
	"github.com/Nithish-Yenaganti/shine/internal/render"
	"github.com/Nithish-Yenaganti/shine/internal/source"
)

type Options struct {
	Source       source.Source
	Theme        config.Theme
	Watch        bool
	UseAltScreen bool
	ShowKeys     bool
	DebugKeys    bool
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
	themeMenu  bool
	themeIndex int
	width      int
	height     int
	content    string
	totalLines int
	headings   []shinemodel.HeadingRef
	matches    []int
	matchIndex int
	status     string
	err        error
	debugKeys  bool
	lastKey    string
}

type fileChanged struct{}
type watchFailed struct {
	err error
}
type terminalBackgroundMsg struct{}

func Run(opts Options) error {
	doc, err := parser.Parse([]byte(opts.Source.Content), opts.Source.Name)
	if err != nil {
		return err
	}
	input := textinput.New()
	input.Prompt = "/"
	input.CharLimit = 128

	content := render.New(88, opts.Theme).
		WithSourcePath(opts.Source.Path).
		WithImages(true).
		Render(doc)
	m := model{
		source:     opts.Source,
		theme:      opts.Theme,
		watch:      opts.Watch,
		viewport:   viewport.New(88, 24),
		search:     input,
		help:       opts.ShowKeys,
		themeIndex: themeIndex(opts.Theme.Name),
		debugKeys:  opts.DebugKeys,
		content:    content,
		totalLines: lineCount(content),
		status:     "shine",
	}
	m.headings = doc.Headings
	programOptions := []tea.ProgramOption{
		tea.WithInput(os.Stdin),
		tea.WithOutput(os.Stdout),
	}
	if opts.UseAltScreen {
		programOptions = append(programOptions, tea.WithAltScreen())
	}
	_, err = tea.NewProgram(m, programOptions...).Run()
	return err
}

func (m model) Init() tea.Cmd {
	m.viewport.SetContent(m.content)
	var cmds []tea.Cmd
	if m.theme.Background != "" {
		cmds = append(cmds, tea.Sequence(terminalBackgroundCmd(m.theme.Background), tea.ClearScreen))
	}
	if m.watch && m.source.Path != "" {
		cmds = append(cmds, watchFile(m.source.Path))
	}
	return tea.Batch(cmds...)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = m.contentWidth()
		m.viewport.Height = max(1, msg.Height-1)
		m.reloadRender()
	case fileChanged:
		m.reloadFile()
		if m.watch && m.source.Path != "" {
			return m, watchFile(m.source.Path)
		}
	case watchFailed:
		m.err = msg.err
	case tea.KeyMsg:
		m.lastKey = keyLabel(msg)
		if m.themeMenu {
			return m.updateThemeMenu(msg)
		}
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
		if opensHelp(msg) {
			m.help = true
			return m, nil
		}
		if togglesHelp(msg) {
			m.help = !m.help
			return m, nil
		}
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Sequence(terminalBackgroundCmd(""), tea.Quit)
		case "/":
			m.searching = true
			m.search.Focus()
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
		case "t":
			m.themeIndex = themeIndex(m.theme.Name)
			m.themeMenu = true
			return m, nil
		case "j", "down":
			m.viewport.LineDown(1)
			return m, nil
		case "k", "up":
			m.viewport.LineUp(1)
			return m, nil
		case "d", "ctrl+d", "pgdown", "pagedown", " ":
			m.viewport.HalfViewDown()
			return m, nil
		case "u", "ctrl+u", "pgup", "pageup":
			m.viewport.HalfViewUp()
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
	body := m.paddedBody(m.viewport.View())
	if m.themeMenu {
		body = m.centeredView(m.themeMenuView())
	} else if m.outline {
		body = m.paddedBody(m.outlineView()) + "\n\n" + body
	}
	if m.help {
		body = m.paddedBody(m.helpView()) + "\n\n" + body
	}
	status := m.statusLine()
	if m.searching {
		status = m.search.View()
	}
	return m.screen(body, status)
}

func (m model) updateThemeMenu(msg tea.KeyMsg) (model, tea.Cmd) {
	names := config.ThemeNames()
	switch msg.String() {
	case "esc", "q", "t":
		m.themeMenu = false
	case "up", "k":
		m.themeIndex--
		if m.themeIndex < 0 {
			m.themeIndex = len(names) - 1
		}
	case "down", "j":
		m.themeIndex = (m.themeIndex + 1) % len(names)
	case "enter":
		name := names[m.themeIndex]
		m.theme = config.ThemeByName(name)
		m.themeMenu = false
		m.status = "theme: " + name
		m.reloadRender()
		return m, tea.Sequence(terminalBackgroundCmd(m.theme.Background), tea.ClearScreen)
	}
	return m, nil
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
	m.viewport.Width = m.contentWidth()
	m.content = render.New(m.contentWidth(), m.theme).
		WithSourcePath(m.source.Path).
		WithImages(true).
		Render(doc)
	m.totalLines = lineCount(m.content)
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
	title := m.style().Foreground(lipgloss.Color(m.theme.Heading)).Bold(true).Render("outline")
	lines = append(lines, title)
	for _, h := range m.headings {
		lines = append(lines, strings.Repeat("  ", max(0, h.Level-1))+h.Text)
	}
	return m.panelStyle().Render(strings.Join(lines, "\n"))
}

func (m model) helpView() string {
	help := strings.Join([]string{
		m.style().Foreground(lipgloss.Color(m.theme.Heading)).Bold(true).Render("keys"),
		"q quit    j/k scroll    g/G top/bottom",
		"d/u half-page          arrows scroll",
		"/ search  n/N matches   o outline",
		"t themes  r reload      h help",
		"? toggle help",
	}, "\n")
	return m.panelStyle().Render(help)
}

func (m model) themeMenuView() string {
	names := config.ThemeNames()
	var lines []string
	lines = append(lines, m.style().Foreground(lipgloss.Color(m.theme.Heading)).Bold(true).Render("themes"))
	lines = append(lines, "")
	for i, name := range names {
		marker := " "
		rowStyle := m.style().Foreground(lipgloss.Color(m.theme.Body))
		if i == m.themeIndex {
			marker = "›"
			rowStyle = rowStyle.
				Foreground(lipgloss.Color("#1f2328")).
				Background(lipgloss.Color(m.theme.MatchHighlight)).
				Bold(true)
		}
		current := ""
		if name == m.theme.Name {
			current = " current"
		}
		lines = append(lines, rowStyle.Render(fmt.Sprintf("%s %s%s", marker, name, current)))
	}
	lines = append(lines, "")
	lines = append(lines, m.style().Foreground(lipgloss.Color(m.theme.Muted)).Render("enter apply   esc close"))
	return m.panelStyle().Width(38).Render(strings.Join(lines, "\n"))
}

func (m model) centeredView(content string) string {
	width := m.width
	if width == 0 {
		width = 88
	}
	height := m.viewport.Height
	if height == 0 {
		height = 23
	}
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
}

func (m model) statusLine() string {
	if m.err != nil {
		return m.style().Foreground(lipgloss.Color(m.theme.CalloutWarning)).Render("error: " + m.err.Error())
	}
	watch := "watch:off"
	if m.watch {
		watch = "watch:on"
	}
	totalLines := max(1, m.totalLines)
	currentLine := min(totalLines, m.viewport.YOffset+1)
	search := "matches:0"
	if len(m.matches) > 0 {
		search = fmt.Sprintf("match:%d/%d", m.matchIndex+1, len(m.matches))
	}
	keyDebug := ""
	if m.debugKeys {
		keyDebug = " | key:" + m.lastKey
	}
	return m.style().Foreground(lipgloss.Color(m.theme.Muted)).Render(
		fmt.Sprintf("%s | theme:%s | %s | line:%d/%d | %s | t themes | h keys%s", m.source.Name, m.theme.Name, watch, currentLine, totalLines, search, keyDebug),
	)
}

func keyLabel(msg tea.KeyMsg) string {
	label := msg.String()
	if label != "" {
		return label
	}
	if len(msg.Runes) == 0 {
		return fmt.Sprintf("%d", msg.Type)
	}
	return string(msg.Runes)
}

func terminalBackgroundCmd(background string) tea.Cmd {
	return func() tea.Msg {
		if background == "" {
			fmt.Fprint(os.Stdout, ansi.ResetBackgroundColor)
		} else {
			fmt.Fprint(os.Stdout, ansi.SetBackgroundColor(background))
		}
		return terminalBackgroundMsg{}
	}
}

func opensHelp(msg tea.KeyMsg) bool {
	switch msg.String() {
	case "h", "H", "f1":
		return true
	}
	return false
}

func togglesHelp(msg tea.KeyMsg) bool {
	if msg.String() == "?" || msg.String() == "shift+/" {
		return true
	}
	return len(msg.Runes) == 1 && msg.Runes[0] == '?'
}

func themeIndex(name string) int {
	for i, candidate := range config.ThemeNames() {
		if candidate == name {
			return i
		}
	}
	return 0
}

func (m model) highlightedContent(query string) string {
	lines := strings.Split(m.content, "\n")
	style := lipgloss.NewStyle().
		Background(lipgloss.Color(m.theme.MatchHighlight)).
		Foreground(lipgloss.Color("#1f2328"))
	for i, line := range lines {
		if isImageProtocolLine(line) {
			continue
		}
		if strings.Contains(strings.ToLower(line), query) {
			lines[i] = style.Render(line)
		}
	}
	return strings.Join(lines, "\n")
}

func isImageProtocolLine(line string) bool {
	return strings.Contains(line, "\x1b_G") || strings.Contains(line, "\U0010eeee")
}

func (m model) contentWidth() int {
	width := m.width
	if width == 0 {
		width = 88
	}
	return max(32, width-2*m.horizontalPadding())
}

func (m model) horizontalPadding() int {
	width := m.width
	if width < 72 {
		return 1
	}
	if width < 120 {
		return 2
	}
	return 4
}

func (m model) paddedBody(body string) string {
	padding := m.horizontalPadding()
	if padding == 0 || body == "" {
		return body
	}
	prefix := strings.Repeat(" ", padding)
	lines := strings.Split(body, "\n")
	for i, line := range lines {
		lines[i] = prefix + line
	}
	return strings.Join(lines, "\n")
}

func (m model) panelStyle() lipgloss.Style {
	width := m.width
	if width == 0 {
		width = 88
	}
	return m.style().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(m.theme.Border)).
		Padding(0, 1).
		Width(max(20, min(width-2, 42)))
}

func (m model) style() lipgloss.Style {
	return lipgloss.NewStyle()
}

func (m model) screen(body string, status string) string {
	return body + "\n" + status
}

func lineCount(value string) int {
	value = strings.TrimRight(value, "\n")
	if value == "" {
		return 0
	}
	return strings.Count(value, "\n") + 1
}

func watchFile(path string) tea.Cmd {
	return func() tea.Msg {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			return watchFailed{err: err}
		}
		defer watcher.Close()
		if err := watcher.Add(path); err != nil {
			return watchFailed{err: err}
		}
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return watchFailed{err: fmt.Errorf("file watcher closed")}
				}
				if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) || event.Has(fsnotify.Rename) || event.Has(fsnotify.Remove) {
					return fileChanged{}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return watchFailed{err: fmt.Errorf("file watcher closed")}
				}
				return watchFailed{err: err}
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

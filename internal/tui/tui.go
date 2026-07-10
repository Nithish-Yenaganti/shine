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
	viewCache  *viewportBodyCache
	search     textinput.Model
	searching  bool
	outline    bool
	help       bool
	themeMenu  bool
	themeIndex int
	width      int
	height     int
	content    string
	contentRev uint64
	hasImages  bool
	totalLines int
	headings   []shinemodel.HeadingRef
	matches    []int
	matchIndex int
	status     string
	err        error
	debugKeys  bool
	lastKey    string
}

type viewportBodyCache struct {
	key    viewportBodyCacheKey
	body   string
	valid  bool
	hits   int
	builds int
}

type viewportBodyCacheKey struct {
	contentRev    uint64
	yOffset       int
	viewportWidth int
	viewportLines int
	terminalWidth int
	leftPadding   int
	rightPadding  int
	themeName     string
	searchQuery   string
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

	initialWidth := 88
	initialHeight := 24
	initial := model{width: initialWidth}
	contentWidth := initial.contentWidth()
	content := render.New(contentWidth, opts.Theme).
		WithSourcePath(opts.Source.Path).
		WithImages(true).
		Render(doc)
	m := model{
		source:     opts.Source,
		theme:      opts.Theme,
		watch:      opts.Watch,
		viewport:   newViewport(contentWidth, initialHeight),
		viewCache:  &viewportBodyCache{},
		search:     input,
		help:       opts.ShowKeys,
		themeIndex: themeIndex(opts.Theme.Name),
		debugKeys:  opts.DebugKeys,
		width:      initialWidth,
		height:     initialHeight,
		content:    content,
		contentRev: 1,
		hasImages:  hasImageProtocolLines(content),
		totalLines: lineCount(content),
		status:     "shine",
	}
	m.headings = doc.Headings
	programOptions := []tea.ProgramOption{
		tea.WithInput(os.Stdin),
		tea.WithOutput(os.Stdout),
		tea.WithMouseCellMotion(),
	}
	if opts.UseAltScreen {
		programOptions = append(programOptions, tea.WithAltScreen())
	}
	_, err = tea.NewProgram(m, programOptions...).Run()
	return err
}

func (m model) Init() tea.Cmd {
	m.setViewportContent(m.content)
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
		m.invalidateViewportBodyCache()
		m.reloadRender()
	case fileChanged:
		m.reloadFile()
		if m.watch && m.source.Path != "" {
			return m, watchFile(m.source.Path)
		}
	case watchFailed:
		m.err = msg.err
	case tea.MouseMsg:
		return m.updateMouse(msg)
	case tea.KeyMsg:
		m.lastKey = keyLabel(msg)
		if m.themeMenu {
			return m.updateThemeMenu(msg)
		}
		if m.searching {
			switch msg.String() {
			case "enter":
				m.searching = false
				m.invalidateViewportBodyCache()
				m.applySearch()
				return m, nil
			case "esc":
				m.searching = false
				m.search.SetValue("")
				m.invalidateViewportBodyCache()
				return m, nil
			}
			var cmd tea.Cmd
			m.search, cmd = m.search.Update(msg)
			return m, cmd
		}
		if opensHelp(msg) {
			m.help = true
			m.invalidateViewportBodyCache()
			return m, nil
		}
		if togglesHelp(msg) {
			m.help = !m.help
			m.invalidateViewportBodyCache()
			return m, nil
		}
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Sequence(terminalBackgroundCmd(""), tea.Quit)
		case "/":
			m.searching = true
			m.search.Focus()
			m.invalidateViewportBodyCache()
			return m, nil
		case "n":
			m.nextMatch()
			return m, nil
		case "N":
			m.previousMatch()
			return m, nil
		case "o":
			m.outline = !m.outline
			m.invalidateViewportBodyCache()
			return m, nil
		case "r":
			m.reloadFile()
			return m, nil
		case "t", "T":
			m.themeIndex = themeIndex(m.theme.Name)
			m.themeMenu = true
			m.invalidateViewportBodyCache()
			return m, nil
		case "j", "down":
			m.lineDown(1)
			return m, nil
		case "k", "up":
			m.lineUp(1)
			return m, nil
		case "d", "ctrl+d", "pgdown", "pagedown", " ":
			m.halfViewDown()
			return m, nil
		case "u", "ctrl+u", "pgup", "pageup":
			m.halfViewUp()
			return m, nil
		case "g":
			m.gotoTop()
			return m, nil
		case "G":
			m.gotoBottom()
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m model) updateMouse(msg tea.MouseMsg) (model, tea.Cmd) {
	if m.themeMenu || m.searching || m.help || m.outline {
		return m, nil
	}
	if msg.Action != tea.MouseActionPress {
		return m, nil
	}
	switch msg.Button {
	case tea.MouseButtonWheelUp:
		m.lineUp(m.viewport.MouseWheelDelta)
	case tea.MouseButtonWheelDown:
		m.lineDown(m.viewport.MouseWheelDelta)
	default:
		return m, nil
	}
	return m, nil
}

func (m model) View() string {
	var body string
	if m.themeMenu {
		body = m.centeredView(m.themeMenuView())
	} else {
		body = m.cachedViewportBody()
		if m.outline {
			body = m.paddedBody(m.outlineView()) + "\n\n" + body
		}
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
	case "esc", "q", "t", "T":
		m.themeMenu = false
		m.invalidateViewportBodyCache()
	case "up", "k":
		m.themeIndex--
		if m.themeIndex < 0 {
			m.themeIndex = len(names) - 1
		}
		m.invalidateViewportBodyCache()
	case "down", "j":
		m.themeIndex = (m.themeIndex + 1) % len(names)
		m.invalidateViewportBodyCache()
	case "enter":
		name := names[m.themeIndex]
		m.theme = config.ThemeByName(name)
		m.themeMenu = false
		m.status = "theme: " + name
		m.invalidateViewportBodyCache()
		m.reloadRender()
		return m, tea.Sequence(terminalBackgroundCmd(m.theme.Background), tea.ClearScreen)
	}
	return m, nil
}

func newViewport(width int, height int) viewport.Model {
	vp := viewport.New(width, height)
	vp.MouseWheelDelta = 6
	return vp
}

func (m model) cachedViewportBody() string {
	if m.hasImages || m.viewCache == nil {
		return m.paddedBody(m.viewport.View())
	}
	key := m.viewportBodyCacheKey()
	if m.viewCache.valid && m.viewCache.key == key {
		m.viewCache.hits++
		return m.viewCache.body
	}
	body := m.paddedBody(m.viewport.View())
	m.viewCache.key = key
	m.viewCache.body = body
	m.viewCache.valid = true
	m.viewCache.builds++
	return body
}

func (m model) viewportBodyCacheKey() viewportBodyCacheKey {
	return viewportBodyCacheKey{
		contentRev:    m.contentRev,
		yOffset:       m.viewport.YOffset,
		viewportWidth: m.viewport.Width,
		viewportLines: m.viewport.Height,
		terminalWidth: m.terminalWidth(),
		leftPadding:   m.leftPadding(),
		rightPadding:  m.rightPadding(),
		themeName:     m.theme.Name,
		searchQuery:   strings.ToLower(strings.TrimSpace(m.search.Value())),
	}
}

func (m *model) setViewportContent(content string) {
	m.viewport.SetContent(content)
	m.contentRev++
	m.hasImages = hasImageProtocolLines(content)
	m.invalidateViewportBodyCache()
}

func (m *model) setYOffset(offset int) {
	before := m.viewport.YOffset
	m.viewport.SetYOffset(offset)
	if m.viewport.YOffset != before {
		m.invalidateViewportBodyCache()
	}
}

func (m *model) lineDown(lines int) {
	before := m.viewport.YOffset
	m.viewport.LineDown(lines)
	if m.viewport.YOffset != before {
		m.invalidateViewportBodyCache()
	}
}

func (m *model) lineUp(lines int) {
	before := m.viewport.YOffset
	m.viewport.LineUp(lines)
	if m.viewport.YOffset != before {
		m.invalidateViewportBodyCache()
	}
}

func (m *model) halfViewDown() {
	before := m.viewport.YOffset
	m.viewport.HalfViewDown()
	if m.viewport.YOffset != before {
		m.invalidateViewportBodyCache()
	}
}

func (m *model) halfViewUp() {
	before := m.viewport.YOffset
	m.viewport.HalfViewUp()
	if m.viewport.YOffset != before {
		m.invalidateViewportBodyCache()
	}
}

func (m *model) gotoTop() {
	before := m.viewport.YOffset
	m.viewport.GotoTop()
	if m.viewport.YOffset != before {
		m.invalidateViewportBodyCache()
	}
}

func (m *model) gotoBottom() {
	before := m.viewport.YOffset
	m.viewport.GotoBottom()
	if m.viewport.YOffset != before {
		m.invalidateViewportBodyCache()
	}
}

func (m *model) invalidateViewportBodyCache() {
	if m.viewCache != nil {
		m.viewCache.valid = false
	}
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
	m.hasImages = hasImageProtocolLines(m.content)
	m.invalidateViewportBodyCache()
	m.applySearch()
}

func (m *model) applySearch() {
	query := strings.ToLower(strings.TrimSpace(m.search.Value()))
	m.matches = nil
	if query == "" {
		m.setViewportContent(m.content)
		return
	}
	for i, line := range strings.Split(m.content, "\n") {
		if strings.Contains(strings.ToLower(line), query) {
			m.matches = append(m.matches, i)
		}
	}
	if len(m.matches) > 0 {
		m.matchIndex = 0
		m.setYOffset(m.matches[0])
	}
	m.setViewportContent(m.highlightedContent(query))
}

func (m *model) nextMatch() {
	if len(m.matches) == 0 {
		m.applySearch()
	}
	if len(m.matches) == 0 {
		return
	}
	m.matchIndex = (m.matchIndex + 1) % len(m.matches)
	m.setYOffset(m.matches[m.matchIndex])
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
	m.setYOffset(m.matches[m.matchIndex])
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
		theme := config.ThemeByName(name)
		label := theme.DisplayName
		if label == "" {
			label = theme.Name
		}
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
		lines = append(lines, rowStyle.Render(fmt.Sprintf("%s %s%s", marker, label, current)))
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
		fmt.Sprintf("%s | theme:%s | %s | line:%d/%d | %s | t themes | h keys%s", m.source.Name, themeLabel(m.theme), watch, currentLine, totalLines, search, keyDebug),
	)
}

func themeLabel(theme config.Theme) string {
	if theme.DisplayName != "" {
		return theme.DisplayName
	}
	return theme.Name
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

func hasImageProtocolLines(content string) bool {
	return strings.Contains(content, "\x1b_G") || strings.Contains(content, "\U0010eeee")
}

func (m model) contentWidth() int {
	return max(32, m.terminalWidth()-m.leftPadding()-m.rightPadding())
}

func (m model) terminalWidth() int {
	if m.width == 0 {
		return 88
	}
	return m.width
}

func (m model) rightPadding() int {
	if m.terminalWidth() < 72 {
		return 2
	}
	return m.terminalWidth() / 5
}

func (m model) leftPadding() int {
	if m.terminalWidth() < 72 {
		return 1
	}
	return m.terminalWidth() / 5
}

func (m model) paddedBody(body string) string {
	if body == "" {
		return body
	}
	prefix := strings.Repeat(" ", m.leftPadding())
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

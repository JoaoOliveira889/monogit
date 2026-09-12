package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/JoaoOliveira889/monogit/internal/pkg/ui"
)

// paletteCommand is one entry in the ":" command palette.
type paletteCommand struct {
	Name    string
	Args    string
	Summary string
	// Run either performs the command or returns the key sequence to replay.
	Run func(m *Model, arg string) (tea.Model, tea.Cmd)
}

const maxPaletteResults = 12

// paletteCommands is the command set reachable from ":". It deliberately
// mirrors existing actions rather than adding new ones, so the palette is a
// discoverable route to the shortcuts instead of a second implementation.
var paletteCommands = []paletteCommand{
	{Name: "fetch-all", Summary: "fetch every repository", Run: func(m *Model, _ string) (tea.Model, tea.Cmd) {
		return m.fetchAllFromPalette()
	}},
	{Name: "pull-all", Summary: "pull every clean repository", Run: func(m *Model, _ string) (tea.Model, tea.Cmd) {
		return m.promptConfirm("Pull all repositories?", "Dirty repositories will be skipped.", "pull_all")
	}},
	{Name: "push-all", Summary: "push every repository that is ahead", Run: func(m *Model, _ string) (tea.Model, tea.Cmd) {
		return m.promptConfirm("Push all repositories with pending commits?", "Only repositories ahead of their upstream will be pushed.", "push_all")
	}},
	{Name: "stash-all", Summary: "stash every dirty repository", Run: func(m *Model, _ string) (tea.Model, tea.Cmd) {
		return m.promptConfirm("Stash changes in all dirty repos?", "Clean repositories will be skipped.", "stash_all")
	}},
	{Name: "filter", Args: "all|dirty|behind|ahead|conflicts|tagged", Summary: "filter the repository list", Run: (*Model).paletteFilter},
	{Name: "theme", Args: "<name>", Summary: "switch the colour theme", Run: (*Model).paletteTheme},
	{Name: "relativenumber", Args: "on|off", Summary: "toggle relative line numbers", Run: (*Model).paletteRelativeNumber},
	{Name: "log", Summary: "open the command log", Run: func(m *Model, _ string) (tea.Model, tea.Cmd) {
		m.previousPanel = m.activePanel
		m.activePanel = CommandLogPanel
		m.refreshLogViewport()
		m.logViewport.GotoBottom()
		return m, nil
	}},
	{Name: "config", Summary: "open the configuration panel", Run: func(m *Model, _ string) (tea.Model, tea.Cmd) {
		m.previousPanel = m.activePanel
		m.activePanel = ConfigPanel
		m.configCursor = 0
		m.refreshViewports()
		return m, nil
	}},
	{Name: "help", Summary: "open the shortcut reference", Run: func(m *Model, _ string) (tea.Model, tea.Cmd) {
		m.pushOverlay(OverlayHelp)
		m.activePanel = HelpPanel
		m.helpSearchInput.Reset()
		m.helpViewport.GotoTop()
		return m, m.helpSearchInput.Focus()
	}},
	{Name: "quit", Summary: "leave MonoGit", Run: func(m *Model, _ string) (tea.Model, tea.Cmd) {
		m.quitting = true
		return m, tea.Quit
	}},
}

func (m *Model) fetchAllFromPalette() (tea.Model, tea.Cmd) {
	if len(m.repos) == 0 {
		return m, nil
	}
	m.statusMsg = "Fetching all..."
	for i := range m.repos {
		m.repos[i].Fetching = true
	}
	return m, m.fetchAllCmd()
}

func (m *Model) paletteFilter(arg string) (tea.Model, tea.Cmd) {
	filters := map[string]StatusFilterType{
		"all":       FilterAll,
		"dirty":     FilterDirty,
		"behind":    FilterBehind,
		"ahead":     FilterAhead,
		"conflicts": FilterConflicts,
		"tagged":    FilterTagged,
	}
	filter, ok := filters[strings.ToLower(strings.TrimSpace(arg))]
	if !ok {
		m.statusMsg = "Unknown filter: " + arg
		return m, nil
	}
	m.statusFilter = filter
	m.invalidateFilterCache()
	m.syncCursorToFilter()
	m.refreshViewports()
	m.statusMsg = "Filter: " + arg
	return m, nil
}

func (m *Model) paletteTheme(arg string) (tea.Model, tea.Cmd) {
	arg = strings.TrimSpace(arg)
	for _, t := range ui.Themes {
		if strings.EqualFold(t.Name, arg) {
			m.cfg.Theme = t.Name
			ui.ApplyTheme(t.Name)
			m.refreshViewports()
			m.statusMsg = "Theme: " + t.Name
			return m, saveConfigCmd(m.cfg)
		}
	}

	names := make([]string, 0, len(ui.Themes))
	for _, t := range ui.Themes {
		names = append(names, t.Name)
	}
	m.statusMsg = "Unknown theme. Available: " + strings.Join(names, ", ")
	return m, nil
}

func (m *Model) paletteRelativeNumber(arg string) (tea.Model, tea.Cmd) {
	switch strings.ToLower(strings.TrimSpace(arg)) {
	case "on", "true", "yes":
		m.cfg.RelativeNumber = true
	case "off", "false", "no":
		m.cfg.RelativeNumber = false
	case "":
		m.cfg.RelativeNumber = !m.cfg.RelativeNumber
	default:
		m.statusMsg = "Expected on or off"
		return m, nil
	}
	m.refreshViewports()
	m.statusMsg = "Relative line numbers: " + onOff(m.cfg.RelativeNumber)
	return m, saveConfigCmd(m.cfg)
}

// matchPaletteCommands returns the commands whose name contains every
// whitespace-separated token of the query, ranked with prefix matches first.
func matchPaletteCommands(query string) []paletteCommand {
	query = strings.ToLower(strings.TrimSpace(query))
	name, _, _ := strings.Cut(query, " ")

	var prefixed, contained []paletteCommand
	for _, c := range paletteCommands {
		switch {
		case name == "" || strings.HasPrefix(c.Name, name):
			prefixed = append(prefixed, c)
		case strings.Contains(c.Name, name):
			contained = append(contained, c)
		}
	}

	matches := append(prefixed, contained...)
	if len(matches) > maxPaletteResults {
		matches = matches[:maxPaletteResults]
	}
	return matches
}

// runPaletteCommand parses and executes a palette line.
func (m *Model) runPaletteCommand(line string) (tea.Model, tea.Cmd) {
	line = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), ":"))
	if line == "" {
		return m, nil
	}

	name, arg, _ := strings.Cut(line, " ")
	name = strings.ToLower(name)

	var matches []paletteCommand
	for _, c := range paletteCommands {
		if c.Name == name {
			matches = []paletteCommand{c}
			break
		}
		if strings.HasPrefix(c.Name, name) {
			matches = append(matches, c)
		}
	}

	switch len(matches) {
	case 0:
		m.statusMsg = "Unknown command: " + name
		return m, nil
	case 1:
		return matches[0].Run(m, strings.TrimSpace(arg))
	default:
		names := make([]string, 0, len(matches))
		for _, c := range matches {
			names = append(names, c.Name)
		}
		m.statusMsg = "Ambiguous command. Did you mean: " + strings.Join(names, ", ")
		return m, nil
	}
}

func (m *Model) renderPalette() string {
	width := m.width * 2 / 3
	if width < 40 {
		width = 40
	}
	if width > m.width-4 {
		width = m.width - 4
	}

	var rows []string
	for i, c := range matchPaletteCommands(m.paletteInput.Value()) {
		name := c.Name
		if c.Args != "" {
			name += " " + c.Args
		}
		line := padRight(name, 32) + ui.SubtleStyle.Render(c.Summary)
		if i == m.paletteCursor {
			line = ui.SelectedItemStyle.Render("▶ " + line)
		} else {
			line = "  " + line
		}
		rows = append(rows, line)
	}

	if len(rows) == 0 {
		rows = append(rows, ui.SubtleStyle.Render("  no matching command"))
	}

	body := m.paletteInput.View() + "\n\n" + strings.Join(rows, "\n")

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ui.ColorHighlight).
		Width(width).
		Padding(1, 2).
		Render(body)
}

func (m *Model) openPalette() (tea.Model, tea.Cmd) {
	m.pushOverlay(OverlayPalette)
	m.paletteInput.Reset()
	m.paletteCursor = 0
	return m, m.paletteInput.Focus()
}

func (m *Model) closePalette() {
	m.closeOverlay(OverlayPalette)
	m.paletteInput.Blur()
	m.paletteInput.Reset()
	m.paletteCursor = 0
}

func (m *Model) handlePaletteKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	matches := matchPaletteCommands(m.paletteInput.Value())

	switch msg.String() {
	case "esc", "ctrl+c":
		m.closePalette()
		return m, nil

	case "up", "ctrl+p":
		m.paletteCursor = clamp(m.paletteCursor-1, 0, max(len(matches)-1, 0))
		return m, nil

	case "down", "ctrl+n":
		m.paletteCursor = clamp(m.paletteCursor+1, 0, max(len(matches)-1, 0))
		return m, nil

	case "tab":
		// Complete to the highlighted command, leaving room for an argument.
		if m.paletteCursor < len(matches) {
			completed := matches[m.paletteCursor].Name
			if matches[m.paletteCursor].Args != "" {
				completed += " "
			}
			m.paletteInput.SetValue(completed)
			m.paletteInput.CursorEnd()
		}
		return m, nil

	case "enter":
		line := m.paletteInput.Value()
		// An empty line runs the highlighted command, which is what the list
		// is for; a typed line is parsed so arguments survive.
		if strings.TrimSpace(line) == "" && m.paletteCursor < len(matches) {
			line = matches[m.paletteCursor].Name
		}
		m.closePalette()
		return m.runPaletteCommand(line)
	}

	before := m.paletteInput.Value()
	var cmd tea.Cmd
	m.paletteInput, cmd = m.paletteInput.Update(msg)
	if m.paletteInput.Value() != before {
		m.paletteCursor = 0
	}
	return m, cmd
}

package tui

import (
	"fmt"
	"hash/fnv"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"

	"github.com/JoaoOliveira889/monogit/internal/pkg/ui"
)

func (m *Model) renderHeader() string {
	brand := renderBrandWordmark(true)
	if m.width < 30 {
		brand = "MonoGit"
	}

	stats := fmt.Sprintf("%d repos ", len(m.repos))

	loading := ""
	if m.isBusy() {
		loading = ui.SpinnerStyle.Render(m.spinnerView() + " Loading...")
	}

	spacerLen := m.width - lipgloss.Width(brand) - lipgloss.Width(stats) - lipgloss.Width(loading)
	if spacerLen < 0 {
		spacerLen = 0
	}
	spacer := strings.Repeat(" ", spacerLen)

	headerLine := " " + lipgloss.JoinHorizontal(lipgloss.Bottom,
		brand,
		spacer,
		loading,
		ui.SubtleStyle.Render(stats),
	)
	headerLine += " "

	border := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ui.ColorBorder)).
		Render(strings.Repeat("─", lipgloss.Width(headerLine)))

	if m.statusMsg != "" {
		status := lipgloss.NewStyle().
			Foreground(ui.ColorSubtle).
			Width(m.width).
			Render(" " + m.statusMsg)
		return headerLine + "\n" + status + "\n" + border
	}

	return headerLine + "\n" + border
}

func (m *Model) renderFooter() string {
	sep := ui.SubtleStyle.Render(" • ")
	var parts []string
	switch {
	case m.showConfirmModal:
		parts = []string{
			m.fmtKey("y", "yes"),
			m.fmtKey("n", "no"),
			m.fmtKey("esc", "cancel"),
		}
	case m.showHelp:
		parts = []string{
			m.fmtKey("jk", "scroll"),
			m.fmtKey("esc", "close"),
		}
	case m.tagFilterModal:
		parts = []string{
			m.fmtKey("↑↓", "navigate"),
			m.fmtKey("space", "toggle"),
			m.fmtKey("enter", "apply"),
			m.fmtKey("esc", "cancel"),
		}
	case m.tagAssignModal:
		parts = []string{
			m.fmtKey("↑↓", "navigate"),
			m.fmtKey(altKeys("space", "enter"), "add/new"),
			m.fmtKey("d", "delete"),
			m.fmtKey("esc", "close"),
		}
	case m.searchMode:
		parts = []string{
			m.fmtKey("esc", "cancel"),
			m.fmtKey("enter", "apply"),
			m.fmtKey("↑↓", "navigate"),
		}
	case m.activePanel == CommandLogPanel:
		parts = []string{
			m.fmtKey("jk", "scroll"),
			m.fmtKey(altKeys("v", "y"), "select/copy"),
			m.fmtKey("1", "repos"),
			m.fmtKey("esc", "back"),
		}
	case m.activePanel == CommitWizardPanel:
		switch m.commitStep {
		case StepAddOption:
			parts = []string{
				m.fmtKey("a", "add all"),
				m.fmtKey("v", "select files"),
				m.fmtKey("esc", "cancel"),
			}
		case StepSelectFiles:
			parts = []string{
				m.fmtKey("space", "toggle"),
				m.fmtKey("enter", "done"),
				m.fmtKey("x", "discard"),
				m.fmtKey("esc", "back"),
			}
		case StepMessage:
			parts = []string{
				m.fmtKey("enter", "commit"),
				m.fmtKey("esc", "cancel"),
			}
		}
	case m.showFiles:
		if m.activePanel == DiffPanel {
			parts = []string{
				m.fmtKey("jk", "scroll"),
				m.fmtKey("y", "copy"),
				m.fmtKey("C", "compact"),
				m.fmtKey(altKeys("tab", "2"), "files"),
				m.fmtKey("1", "repos"),
			}
		} else {
			parts = []string{
				m.fmtKey("jk", "nav"),
				m.fmtKey(altKeys("v", "y"), "select/copy"),
				m.fmtKey("space", "select"),
				m.fmtKey("x", "discard"),
				m.fmtKey(altKeys("a", "n"), "all | none"),
				m.fmtKey("enter", "done"),
				m.fmtKey(altKeys("tab", "3"), "diff"),
			}
		}
	case m.showBranches:
		parts = []string{
			m.fmtKey("jk", "nav"),
			m.fmtKey("enter", "checkout"),
			m.fmtKey("M", "merge"),
			m.fmtKey("n", "new"),
			m.fmtKey("d", "delete"),
			m.fmtKey("esc", "back"),
		}
	case m.showStashes:
		parts = []string{
			m.fmtKey("jk", "nav"),
			m.fmtKey(altKeys("p", "enter"), "pop"),
			m.fmtKey("a", "apply"),
			m.fmtKey("d", "drop"),
			m.fmtKey("esc", "back"),
		}
	case m.showConflicts:
		parts = []string{
			m.fmtKey("jk", "nav"),
			m.fmtKey("enter", "resolve"),
			m.fmtKey("esc", "back"),
		}
	case m.activePanel == RepoPanel:
		parts = []string{
			m.fmtKey("hjkl", "nav"),
			m.fmtKey("/", "search"),
			m.fmtKey("ctrl+g", "filter"),
			m.fmtKey("ctrl+t", "tags"),
			m.fmtKey(altKeys("f", "F"), "fetch"),
			m.fmtKey(altKeys("p", "P"), "pull"),
			m.fmtKey(altKeys("u", "U"), "push"),
			m.fmtKey("c", "commit"),
			m.fmtKey("B", "checkout-all"),
			m.fmtKey("Z", "stash-all"),
			m.fmtKey("?", "help"),
		}
	default:
		parts = []string{
			m.fmtKey("jk", "scroll"),
			m.fmtKey("x", "undo"),
			m.fmtKey("g", "graph"),
			m.fmtKey("z/Z", "stash"),
			m.fmtKey("1", "repos"),
		}
	}

	return m.renderResponsiveFooter(parts, sep, m.fmtKey("?", "help"))
}

func (m *Model) renderResponsiveFooter(parts []string, sep, help string) string {
	version := ui.SubtleStyle.Render(fmt.Sprintf("MonoGit %s", Version))

	contentWidth := m.width - 2
	if contentWidth < 10 {
		contentWidth = 10
	}

	rendered := strings.Join(parts, sep)
	maxLeftWidth := contentWidth - lipgloss.Width(version) - 1
	if maxLeftWidth < lipgloss.Width(help) {
		maxLeftWidth = lipgloss.Width(help)
	}
	for len(parts) > 0 && lipgloss.Width(rendered)+lipgloss.Width(sep)+lipgloss.Width(help) > maxLeftWidth {
		parts = parts[:len(parts)-1]
		rendered = strings.Join(parts, sep)
	}

	left := help
	if rendered != "" {
		left = rendered + sep + help
	}

	spacerLen := contentWidth - lipgloss.Width(left) - lipgloss.Width(version)
	if spacerLen < 0 {
		spacerLen = 0
	}
	spacer := strings.Repeat(" ", spacerLen)

	footerText := " " + left + spacer + version
	if footerWidth := lipgloss.Width(footerText); footerWidth < contentWidth+1 {
		footerText += strings.Repeat(" ", contentWidth+1-footerWidth)
	}

	return ui.FooterStyle.Padding(0, 0).Render(footerText)
}

func (m *Model) fmtKey(k, action string) string {
	return ui.FooterKeyStyle.Render(k) + " " + ui.FooterActionStyle.Render(action)
}

func altKeys(keys ...string) string {
	return strings.Join(keys, " | ")
}

func (m *Model) renderConfirmationModal() string {
	content := ui.ValueStyle.Render(m.confirmModalTitle)
	if m.confirmModalDetail != "" {
		content += "\n\n" + ui.SubtleStyle.Render(m.confirmModalDetail)
	}
	options := lipgloss.JoinHorizontal(lipgloss.Center,
		m.fmtKey("y", "Yes"),
		"   ",
		m.fmtKey("n", "No"),
	)

	if m.confirmModalAction == "delete_branch_options" {
		var opts []string
		if m.branchCursor < len(m.branches) {
			b := m.branches[m.branchCursor]
			if b.IsLocal {
				opts = append(opts, m.fmtKey("l", "Local"))
			}
			if b.IsRemote {
				opts = append(opts, m.fmtKey("r", "Remote"))
			}
		}
		opts = append(opts, m.fmtKey("esc", "Cancel"))

		var finalOpts []string
		for i, o := range opts {
			finalOpts = append(finalOpts, o)
			if i < len(opts)-1 {
				finalOpts = append(finalOpts, "   ")
			}
		}
		options = lipgloss.JoinHorizontal(lipgloss.Center, finalOpts...)
	}

	return lipgloss.JoinVertical(lipgloss.Center,
		ui.PanelTitleStyle.Render(" Confirmation "),
		"",
		content,
		"",
		options,
	)
}

func (m *Model) renderInputModal() string {
	title := " Input "
	switch m.inputAction {
	case "create_branch":
		title = " Create Branch "
	case "pattern_stage":
		title = " Stage by Pattern "
	case "create_tag_version":
		title = " Tag Version "
	case "create_tag_message":
		title = " Tag Message "
	case "new_tag":
		title = " New Tag "
	case "commit":
		title = " Commit Message "
	}

	return lipgloss.JoinVertical(lipgloss.Center,
		ui.PanelTitleStyle.Render(title),
		"",
		m.statusMsg,
		"",
		ui.InputStyle.Render(m.commitInput.View()),
		"",
		m.fmtKey("enter", "next/confirm")+"   "+m.fmtKey("esc", "cancel"),
	)
}

func (m *Model) renderSearchSection(width int) string {
	inputWidth := width - 4
	if inputWidth < 10 {
		inputWidth = 10
	}
	searchInput := m.searchInput
	searchInput.Width = inputWidth
	accent := lipgloss.Color(ui.ColorMono)
	if m.searchMode {
		accent = lipgloss.Color(ui.ColorGit)
	}

	label := lipgloss.NewStyle().
		Foreground(accent).
		Bold(true).
		Render(" Search ")
	searchStyle := ui.InputStyle.BorderForeground(accent).Width(inputWidth)
	if m.searchMode {
		searchStyle = searchStyle.Bold(true)
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		label,
		searchStyle.Render(searchInput.View()),
	)
}

func (m *Model) renderEditorModal() string {
	var lines []string
	for i, editor := range m.availableEditors {
		prefix := "  "
		style := lipgloss.NewStyle().Foreground(ui.ColorFg)
		if i == m.editorCursor {
			prefix = ui.IconClean + " "
			style = style.Background(ui.ColorHighlight).Foreground(ui.ColorBg).Bold(true)
		}
		lines = append(lines, prefix+style.Render(editor))
	}

	return lipgloss.JoinVertical(lipgloss.Center,
		ui.PanelTitleStyle.Render(" Select Editor "),
		"",
		strings.Join(lines, "\n"),
		"",
		m.fmtKey("↑↓", "navigate")+"   "+m.fmtKey("enter", "open")+"   "+m.fmtKey("esc", "cancel"),
	)
}

func (m *Model) tagColor(tag string) lipgloss.Color {
	h := fnv.New32a()
	h.Write([]byte(tag))
	idx := int(h.Sum32()) % len(ui.GraphColors)
	return ui.GraphColors[idx]
}

func truncateRunes(value string, max int) string {
	if max <= 0 {
		return ""
	}
	runes := []rune(value)
	if len(runes) <= max {
		return value
	}
	if max <= 3 {
		return string(runes[:max])
	}
	return string(runes[:max-3]) + "..."
}

func (m *Model) renderTagBadge(tag string) string {
	return lipgloss.NewStyle().
		Background(m.tagColor(tag)).
		Foreground(ui.ColorBg).
		Bold(true).
		Padding(0, 1).
		Render(truncateRunes(tag, maxTagLabelWidth))
}

func wrapPlainText(text string, width int) []string {
	if width < 1 {
		width = 1
	}
	if text == "" {
		return []string{""}
	}

	var lines []string
	for _, paragraph := range strings.Split(text, "\n") {
		if paragraph == "" {
			lines = append(lines, "")
			continue
		}

		words := strings.Fields(paragraph)
		if len(words) == 0 {
			lines = append(lines, "")
			continue
		}

		current := words[0]
		for _, word := range words[1:] {
			candidate := current + " " + word
			if lipgloss.Width(candidate) <= width {
				current = candidate
				continue
			}

			lines = append(lines, current)
			if lipgloss.Width(word) <= width {
				current = word
				continue
			}

			runes := []rune(word)
			for len(runes) > width {
				lines = append(lines, string(runes[:width]))
				runes = runes[width:]
			}
			current = string(runes)
		}
		lines = append(lines, current)
	}
	return lines
}

func (m *Model) clampModalSize(widthOffset, minWidth, heightOffset, minHeight int) (int, int) {
	w := m.width - widthOffset
	if w < minWidth {
		w = minWidth
	}
	if w > m.width {
		w = m.width
	}
	h := m.height - heightOffset
	if h < minHeight {
		h = minHeight
	}
	if h > m.height {
		h = m.height
	}
	return w, h
}

func (m *Model) renderModalShell(title, body, footer string) string {
	panelWidth, panelHeight := m.clampModalSize(4, 60, 4, 18)

	panel := ui.ActivePanelStyle.
		BorderStyle(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color(ui.ColorHighlight)).
		Width(panelWidth).
		Height(panelHeight).
		Padding(1, 2)

	content := lipgloss.JoinVertical(lipgloss.Left,
		ui.PanelTitleStyle.Render(" "+title+" "),
		"",
		body,
	)
	if footer != "" {
		content = lipgloss.JoinVertical(lipgloss.Left,
			content,
			"",
			ui.SubtleStyle.Render(footer),
		)
	}

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, panel.Render(content))
}

func (m *Model) renderTagFilterModal(width, height int) string {
	contentWidth := width - 4
	if contentWidth < 40 {
		contentWidth = 40
	}

	if len(m.availableTags) == 0 {
		return lipgloss.JoinVertical(lipgloss.Left,
			ui.SubtleStyle.Render("No tags defined yet."),
			"",
			ui.SubtleStyle.Render("Use ctrl+t to edit tags in the right panel."),
		)
	}

	var lines []string
	if m.tagFilterActive && len(m.tagFilter) > 0 {
		activeStr := "Filtering:"
		for _, t := range m.tagFilter {
			activeStr += " " + m.renderTagBadge(t)
		}
		lines = append(lines, activeStr)
	} else {
		lines = append(lines, ui.SubtleStyle.Render("Showing all tags"))
	}
	lines = append(lines, "")

	for i, tag := range m.availableTags {
		checked := "○"
		isActive := m.tagModalSelections[i]

		if isActive {
			checked = "●"
		}

		row := "  " + checked + "  " + m.renderTagBadge(tag)
		if i == m.tagModalCursor {
			row = lipgloss.NewStyle().
				Background(lipgloss.Color(ui.ColorHighlight)).
				Foreground(lipgloss.Color(ui.ColorBg)).
				Bold(true).
				Width(contentWidth).
				Render("> " + checked + "  " + tag)
		}
		lines = append(lines, row)
	}

	return strings.Join(lines, "\n")
}

func (m *Model) renderRepoTagsSection(width int) string {
	r := m.selectedRepo()
	if r == nil {
		return ""
	}

	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ui.ColorGit)).
		Bold(true).
		Render(" Tags " + fmt.Sprintf("(%d/%d)", len(r.Tags), maxTagsPerRepo))

	if !m.tagAssignModal {
		if len(r.Tags) == 0 {
			return lipgloss.JoinVertical(lipgloss.Left,
				title,
				ui.SubtleStyle.Render("  No tags assigned"),
				ui.SubtleStyle.Render("  Use ctrl+t to edit tags."),
			)
		}

		var badges []string
		for _, tag := range r.Tags {
			badges = append(badges, m.renderTagBadge(tag))
		}
		return lipgloss.JoinVertical(lipgloss.Left,
			title,
			" ",
			"  "+strings.Join(badges, " "),
			ui.SubtleStyle.Render("  Use ctrl+t to edit tags."),
		)
	}

	lines := make([]string, 0, len(r.Tags)+2)
	if len(r.Tags) > 0 {
		var selected []string
		for _, tag := range r.Tags {
			selected = append(selected, m.renderTagBadge(tag))
		}
		lines = append(lines, "  Selected: "+strings.Join(selected, " "))
	} else {
		lines = append(lines, ui.SubtleStyle.Render("  No tags assigned"))
	}
	lines = append(lines, "")

	displayTags := make([]string, len(r.Tags))
	copy(displayTags, r.Tags)
	displayTags = append(displayTags, "+ New tag...")

	for i, tag := range displayTags {
		isNewTag := i >= len(r.Tags)
		isCursor := i == m.tagModalCursor
		tagWidth := width - 10
		if tagWidth < 10 {
			tagWidth = 10
		}

		if isNewTag {
			row := "  + New tag..."
			if isCursor {
				row = lipgloss.NewStyle().Foreground(lipgloss.Color(ui.ColorGit)).Bold(true).Render("> " + row)
			}
			lines = append(lines, row)
			continue
		}

		row := "  " + m.renderTagBadge(truncateRunes(tag, tagWidth))
		if isCursor {
			row = lipgloss.NewStyle().Foreground(lipgloss.Color(ui.ColorGit)).Bold(true).Render("> ") + m.renderTagBadge(truncateRunes(tag, tagWidth))
		}
		lines = append(lines, row)
	}

	footer := m.fmtKey("↑↓", "navigate") + "   " + m.fmtKey("space/enter", "add/new") + "   " + m.fmtKey("d", "delete") + "   " + m.fmtKey("esc", "close")
	if len(r.Tags) >= maxTagsPerRepo {
		footer = ui.ErrorStyle.Render("Tag limit reached: max 4 per repo") + "   " + footer
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		title,
		"",
		strings.Join(lines, "\n"),
		"",
		footer,
	)
}

func (m *Model) renderCommitWizardModal() string {
	var content string
	title := " Commit Wizard "

	switch m.commitStep {
	case StepAddOption:
		content = "How would you like to prepare this commit?\n\n" +
			m.fmtKey("a", "Add all files") + "\n" +
			m.fmtKey("v", "Select files manually")
	case StepSelectFiles:
		content = "Select files in the right panel. Git only changes when you confirm the commit.\n" +
			m.fmtKey("space", "Toggle") + "  " +
			m.fmtKey("enter", "Done") + "  " +
			m.fmtKey("esc", "Cancel")
	case StepMessage:
		content = "Commit message:\n\n" + ui.InputStyle.Render(m.commitInput.View())
	}

	return lipgloss.JoinVertical(lipgloss.Center,
		ui.PanelTitleStyle.Render(title),
		"",
		content,
	)
}

func (m *Model) renderHelpOverlay() string {
	panelWidth, panelHeight := m.clampModalSize(6, 72, 4, 20)

	innerWidth := panelWidth - 6
	if innerWidth < 60 {
		innerWidth = 60
	}
	innerHeight := panelHeight - 6
	if innerHeight < 12 {
		innerHeight = 12
	}

	title := lipgloss.JoinHorizontal(lipgloss.Bottom,
		renderBrandWordmark(true),
		" ",
		ui.BrandTitleStyle.Render("SHORTCUTS"),
	)

	vpHeight := innerHeight - 3
	if vpHeight < 5 {
		vpHeight = 5
	}
	if m.helpViewport.Width != innerWidth || m.helpViewport.Height != vpHeight {
		m.helpViewport = viewport.New(innerWidth, vpHeight)
	} else {
		m.helpViewport.Width = innerWidth
		m.helpViewport.Height = vpHeight
	}

	body := m.renderHelpMenu(innerWidth, 999)
	m.helpViewport.SetContent(body)

	content := lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.NewStyle().Align(lipgloss.Center).Width(innerWidth).Render(title),
		"",
		m.helpViewport.View(),
		"",
		lipgloss.NewStyle().Align(lipgloss.Center).Width(innerWidth).Render(ui.SubtleStyle.Render("Press ESC or ctrl+p to close")),
	)

	panelStyle := ui.ActivePanelStyle.
		BorderStyle(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color(ui.ColorHighlight)).
		Width(panelWidth).
		Height(panelHeight).
		Padding(1, 2)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, panelStyle.Render(content))
}

func (m *Model) renderHelpMenu(width, height int) string {
	type helpEntry struct {
		key    string
		action string
	}
	type helpSection struct {
		title   string
		entries []helpEntry
	}

	sections := []helpSection{
		{
			title: "NAVIGATION",
			entries: []helpEntry{
				{key: "jk | arrows", action: "Move selection"},
				{key: "hl | arrows", action: "Switch panels"},
				{key: "< | >", action: "Resize panels"},
				{key: "1 | 2 | 3", action: "Jump to panel"},
				{key: "tab", action: "Cycle focus"},
				{key: "v | y", action: "Start selection | copy"},
				{key: "/ | ctrl+f", action: "Search repos"},
				{key: "esc", action: "Back | cancel modal"},
				{key: "q | ctrl+c", action: "Quit"},
			},
		},
		{
			title: "REPOSITORY",
			entries: []helpEntry{
				{key: "f | F", action: "Fetch (one | all, direct)"},
				{key: "p | P", action: "Pull (one | all)"},
				{key: "u | U", action: "Push (one | all)"},
				{key: "c", action: "Commit wizard"},
				{key: "b", action: "List branches"},
				{key: "m", action: "Resolve merge conflicts"},
				{key: "t", action: "Deploy tag"},
				{key: "s", action: "Stash changes"},
				{key: "S", action: "Stash list panel"},
				{key: "Z", action: "Stash all (dirty filtered)"},
				{key: "B", action: "Checkout branch (all filtered)"},
				{key: "z", action: "Undo last commit"},
				{key: "e", action: "Open repo in editor"},
				{key: "w", action: "Open repo in browser"},
			},
		},
		{
			title: "FILES & DIFF",
			entries: []helpEntry{
				{key: "space", action: "Toggle file selection"},
				{key: "a | n", action: "Select all | deselect all"},
				{key: "x", action: "Discard file changes"},
				{key: "g", action: "Toggle graph log view"},
				{key: "C", action: "Toggle compact diff (functions)"},
				{key: "o", action: "Command log"},
			},
		},
		{
			title: "COMMIT WIZARD",
			entries: []helpEntry{
				{key: "a", action: "Add all files"},
				{key: "v", action: "Select files manually"},
				{key: "space", action: "Toggle file selection"},
				{key: "enter", action: "Advance | confirm"},
				{key: "esc", action: "Cancel"},
			},
		},
		{
			title: "STASH MODE",
			entries: []helpEntry{
				{key: "p | enter", action: "Pop stash"},
				{key: "a", action: "Apply stash"},
				{key: "d", action: "Drop stash"},
				{key: "esc", action: "Close stash panel"},
			},
		},
		{
			title: "BRANCH MODE",
			entries: []helpEntry{
				{key: "enter", action: "Checkout branch"},
				{key: "M", action: "Merge branch into HEAD"},
				{key: "n", action: "Create new branch"},
				{key: "d", action: "Delete branch (local|remote)"},
				{key: "esc", action: "Close branch panel"},
			},
		},
		{
			title: "CONFLICTS",
			entries: []helpEntry{
				{key: "enter", action: "Open mergetool for file"},
				{key: "jk", action: "Navigate conflict list"},
				{key: "esc", action: "Close conflicts panel"},
			},
		},
		{
			title: "TAGS",
			entries: []helpEntry{
				{key: "ctrl+g", action: "Filter repos by tags"},
				{key: "ctrl+t", action: "Edit tags in right panel"},
				{key: "d", action: "Remove tag (in editor)"},
			},
		},
	}

	renderSection := func(section helpSection) string {
		keyWidth := 0
		for _, entry := range section.entries {
			if w := lipgloss.Width(entry.key); w > keyWidth {
				keyWidth = w
			}
		}
		lines := []string{ui.PanelTitleStyle.Render(" " + section.title + " ")}
		for _, entry := range section.entries {
			actionWidth := width - keyWidth - 6
			if actionWidth < 18 {
				actionWidth = 18
			}
			wrappedAction := wrapPlainText(entry.action, actionWidth)
			for i, wrapped := range wrappedAction {
				if i == 0 {
					lines = append(lines, ui.LabelStyle.Render("  "+entry.key+":")+"  "+ui.ValueStyle.Render(wrapped))
				} else {
					lines = append(lines, strings.Repeat(" ", keyWidth+4)+"  "+ui.ValueStyle.Render(wrapped))
				}
			}
		}
		return strings.Join(lines, "\n")
	}

	columnCount := 1
	contentWidth := width
	if contentWidth < 56 {
		contentWidth = 56
	}
	switch {
	case contentWidth >= 140:
		columnCount = 3
	case contentWidth >= 100:
		columnCount = 2
	}
	if height >= 28 && columnCount < 3 && contentWidth >= 110 {
		columnCount++
	}
	if columnCount > len(sections) {
		columnCount = len(sections)
	}

	var columnGroups [][]helpSection
	switch columnCount {
	case 3:
		columnGroups = [][]helpSection{
			sections[0:2],
			sections[2:5],
			sections[5:8],
		}
	case 2:
		columnGroups = [][]helpSection{
			sections[0:3],
			sections[3:8],
		}
	default:
		columnGroups = [][]helpSection{sections}
	}

	columnWidth := contentWidth
	if columnCount > 1 {
		columnWidth = (contentWidth - 4*(columnCount-1)) / columnCount
	}
	if columnWidth < 24 {
		columnWidth = 24
	}

	var columns []string
	for _, group := range columnGroups {
		var columnSections []string
		for _, section := range group {
			columnSections = append(columnSections, renderSection(section))
		}
		columns = append(columns, lipgloss.NewStyle().Width(columnWidth).Render(lipgloss.JoinVertical(lipgloss.Left, columnSections...)))
	}

	content := columns[0]
	for i := 1; i < len(columns); i++ {
		content = lipgloss.JoinHorizontal(lipgloss.Top, content, "    ", columns[i])
	}

	return content
}

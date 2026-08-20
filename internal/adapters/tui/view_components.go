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
	var brand string
	switch {
	case m.width < 30:
		brand = ui.BrandMonoStyle.Render("MG")
	case m.width < 60:
		brand = renderBrandWordmark(true)
	default:
		brand = renderBrandWordmark(true)
	}

	healthSummary := m.renderWorkspaceHealth()

	loading := ""
	if m.isBusy() {
		if m.width >= 40 {
			loading = ui.SpinnerStyle.Render(m.spinnerView() + " " + m.busyLabel())
		} else {
			loading = ui.SpinnerStyle.Render(m.spinnerView())
		}
	}

	sep := ui.SubtleStyle.Render(" │ ")
	headerLine := " " + brand + sep + healthSummary
	if loading != "" {
		headerLine += "  " + loading
	}

	border := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ui.ColorBorder)).
		Render(strings.Repeat("─", m.width))

	if m.statusMsg != "" {
		styledStatus := m.renderHeaderStatusBar()
		return headerLine + "\n" + styledStatus + "\n" + border
	}

	return headerLine + "\n" + border
}

func (m *Model) renderWorkspaceHealth() string {
	total := len(m.repos)
	if total == 0 {
		return ui.SubtleStyle.Render("● No repos")
	}

	var cleanCount, dirtyCount, aheadCount, behindCount, conflictCount int
	for _, r := range m.repos {
		if r.HasConflicts {
			conflictCount++
		}
		if r.IsDirty {
			dirtyCount++
		}
		if r.Ahead > 0 {
			aheadCount++
		}
		if r.Behind > 0 {
			behindCount++
		}
		if !r.IsDirty && r.Ahead == 0 && r.Behind == 0 && !r.HasConflicts {
			cleanCount++
		}
	}

	var parts []string
	filtered := len(m.filteredRepos())
	dot := ui.SubtleStyle.Render("●")
	sep := ui.SubtleStyle.Render(" · ")

	if m.searchFilterQuery() != "" || len(m.tagFilter) > 0 || m.statusFilter != FilterAll {
		parts = append(parts, fmt.Sprintf("%s %d/%d repos", dot, filtered, total))
	} else {
		parts = append(parts, fmt.Sprintf("%s %d repos", dot, total))
	}

	if cleanCount > 0 {
		parts = append(parts, ui.CleanStyle.Render(fmt.Sprintf("%s %d clean", ui.IconClean, cleanCount)))
	}
	if behindCount > 0 {
		parts = append(parts, ui.BehindStyle.Render(fmt.Sprintf("%s %d behind", ui.IconBehind, behindCount)))
	}
	if aheadCount > 0 {
		parts = append(parts, ui.AheadStyle.Render(fmt.Sprintf("%s %d ahead", ui.IconAhead, aheadCount)))
	}
	if dirtyCount > 0 {
		parts = append(parts, ui.DirtyStyle.Render(fmt.Sprintf("%s %d dirty", ui.IconDirty, dirtyCount)))
	}
	if conflictCount > 0 {
		if conflictCount == 1 {
			parts = append(parts, ui.ErrorStyle.Render("! 1 conflict"))
		} else {
			parts = append(parts, ui.ErrorStyle.Render(fmt.Sprintf("! %d conflicts", conflictCount)))
		}
	}

	return strings.Join(parts, sep)
}

func (m *Model) renderHeaderStatusBar() string {
	if m.statusMsg == "" {
		return ""
	}
	switch {
	case strings.HasPrefix(m.statusMsg, "✓"):
		return ui.StatusSuccessStyle.Width(m.width).Render(" " + m.statusMsg)
	case strings.HasPrefix(m.statusMsg, "✗"):
		return ui.StatusErrorStyle.Width(m.width).Render(" " + m.statusMsg)
	case strings.HasPrefix(m.statusMsg, "⚠"):
		return ui.StatusWarningStyle.Width(m.width).Render(" " + m.statusMsg)
	default:
		return ui.StatusInfoStyle.Width(m.width).Render(" " + m.statusMsg)
	}
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
	case m.filterModal:
		parts = []string{
			m.fmtKey("↑↓", "navigate"),
			m.fmtKey("enter", "select"),
			m.fmtKey("esc", "cancel"),
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
	case m.activePanel == ConfigPanel:
		parts = []string{
			m.fmtKey("↑↓/jk", "navigate"),
			m.fmtKey("enter", "select/edit"),
			m.fmtKey("esc/,", "close"),
		}
	case m.showRebase:
		parts = []string{
			m.fmtKey("jk", "nav"),
			m.fmtKey("p/s/f/r/d", "action"),
			m.fmtKey("J/K", "reorder"),
			m.fmtKey("enter", "rebase"),
			m.fmtKey("esc", "cancel"),
		}
	case m.activePanel == RepoPanel:
		parts = []string{
			m.fmtKey("hjkl", "nav"),
			m.fmtKey("enter", "open"),
			m.fmtKey("f", "fetch"),
			m.fmtKey("u", "push"),
			m.fmtKey("b", "branches"),
		}
	default:
		parts = []string{
			m.fmtKey("hjkl", "nav"),
			m.fmtKey("enter", "details"),
			m.fmtKey("d", "diff"),
			m.fmtKey("y", "copy hash"),
			m.fmtKey("g", "graph"),
			m.fmtKey("esc", "back"),
		}
	}

	return m.renderResponsiveFooter(parts, sep)
}

func (m *Model) renderResponsiveFooter(parts []string, sep string) string {
	version := ui.SubtleStyle.Render(fmt.Sprintf("MonoGit %s", Version))
	help := m.fmtKey("?", "help")
	fixedRight := help + "  " + version

	contentWidth := m.width - 2
	if contentWidth < 10 {
		contentWidth = 10
	}

	rendered := strings.Join(parts, sep)
	maxLeftWidth := contentWidth - lipgloss.Width(fixedRight) - 1

	for len(parts) > 0 && lipgloss.Width(rendered) > maxLeftWidth {
		parts = parts[:len(parts)-1]
		rendered = strings.Join(parts, sep)
	}

	left := rendered
	spacerLen := contentWidth - lipgloss.Width(left) - lipgloss.Width(fixedRight)
	if spacerLen < 0 {
		spacerLen = 0
	}
	spacer := strings.Repeat(" ", spacerLen)

	footerText := " " + left + spacer + fixedRight
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

// joinFooterKeys joins formatted key hints with the standard " • " separator.
func (m *Model) joinFooterKeys(parts ...string) string {
	sep := ui.SubtleStyle.Render(" • ")
	return strings.Join(parts, sep)
}

// busyLabel returns a short contextual label for the spinner in the header.
func (m *Model) busyLabel() string {
	if m.scanning {
		return "Scanning…"
	}
	if m.diffFetching || m.compactFetching {
		return "Loading diff…"
	}
	for _, r := range m.repos {
		if r.Fetching {
			return "Fetching…"
		}
		if r.Pulling {
			return "Pulling…"
		}
		if r.Pushing {
			return "Pushing…"
		}
		if r.Committing {
			return "Committing…"
		}
		if r.Stashing {
			return "Stashing…"
		}
		if r.Tagging {
			return "Tagging…"
		}
		if r.CheckingOut {
			return "Checking out…"
		}
		if r.Merging {
			return "Merging…"
		}
	}
	return "Loading…"
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
	if lipgloss.Width(value) <= max {
		return value
	}
	ellipsis := "…"
	ellipsisW := lipgloss.Width(ellipsis)
	limit := max - ellipsisW
	if limit <= 0 {
		return ellipsis
	}
	var b strings.Builder
	accumW := 0
	for _, r := range value {
		rw := lipgloss.Width(string(r))
		if accumW+rw > limit {
			break
		}
		b.WriteRune(r)
		accumW += rw
	}
	return b.String() + ellipsis
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

	panel := ui.ActivePanelStyle.
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ui.ColorHighlight)).
		Width(m.modalWidthForContent(content)).
		Padding(1, 2)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, panel.Render(content))
}

func (m *Model) renderFilterModal(width, height int) string {
	type filterCategory struct {
		label  string
		filter StatusFilterType
		count  int
	}

	var dirtyCount, behindCount, aheadCount, conflictCount, taggedCount int
	for _, r := range m.repos {
		if r.IsDirty {
			dirtyCount++
		}
		if r.Behind > 0 {
			behindCount++
		}
		if r.Ahead > 0 {
			aheadCount++
		}
		if r.HasConflicts {
			conflictCount++
		}
		if len(r.Tags) > 0 {
			taggedCount++
		}
	}

	categories := []filterCategory{
		{label: "All", filter: FilterAll, count: len(m.repos)},
		{label: "Dirty", filter: FilterDirty, count: dirtyCount},
		{label: "Behind", filter: FilterBehind, count: behindCount},
		{label: "Ahead", filter: FilterAhead, count: aheadCount},
		{label: "Conflicts", filter: FilterConflicts, count: conflictCount},
		{label: "Tagged", filter: FilterTagged, count: taggedCount},
	}

	var lines []string
	lines = append(lines, ui.SubtleStyle.Render("Status filter categories:"))
	lines = append(lines, "")

	for i, cat := range categories {
		radio := "○"
		if m.statusFilter == cat.filter {
			radio = "●"
		}
		row := fmt.Sprintf("  %s  %-14s %3d", radio, cat.label, cat.count)
		if i == m.filterModalCursor {
			row = lipgloss.NewStyle().
				Background(lipgloss.Color(ui.ColorHighlight)).
				Foreground(lipgloss.Color(ui.ColorBg)).
				Bold(true).
				Render("> " + fmt.Sprintf("%s  %-14s %3d", radio, cat.label, cat.count))
		}
		lines = append(lines, row)
	}

	return strings.Join(lines, "\n")
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
	panelWidth := m.width - 4
	if panelWidth > 140 {
		panelWidth = 140
	}
	if panelWidth < 30 {
		panelWidth = 30
	}

	innerWidth := panelWidth - 4
	if innerWidth < 20 {
		innerWidth = 20
	}

	title := lipgloss.JoinHorizontal(lipgloss.Bottom,
		renderBrandWordmark(true),
		" ",
		ui.BrandTitleStyle.Render("SHORTCUTS"),
	)

	maxViewportHeight := m.height - 10
	if maxViewportHeight < 5 {
		maxViewportHeight = 5
	}
	vpHeight := maxViewportHeight
	if m.helpViewport.Width != innerWidth-1 || m.helpViewport.Height != vpHeight {
		m.helpViewport = viewport.New(innerWidth-1, vpHeight)
	} else {
		m.helpViewport.Width = innerWidth - 1
		m.helpViewport.Height = vpHeight
	}

	body := m.renderHelpMenu(innerWidth-1, vpHeight)
	if contentHeight := lipgloss.Height(body); contentHeight < vpHeight {
		vpHeight = contentHeight
		if vpHeight < 5 {
			vpHeight = 5
		}
		m.helpViewport.Height = vpHeight
	}
	m.helpViewport.SetContent(body)

	content := lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.NewStyle().Align(lipgloss.Center).Width(innerWidth).Render(title),
		"",
		renderViewportWithScrollbar(m.helpViewport, true),
		"",
		lipgloss.NewStyle().Align(lipgloss.Center).Width(innerWidth).Render(ui.SubtleStyle.Render("Press ESC or ctrl+p to close")),
	)

	panelStyle := ui.ActivePanelStyle.
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ui.ColorHighlight)).
		Width(panelWidth).
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
				{key: "R", action: "Interactive rebase"},
				{key: "ctrl+y", action: "Cherry-pick commit"},
				{key: "ctrl+r", action: "Revert commit"},
				{key: ",", action: "Open configuration"},
			},
		},
		{
			title: "REBASE MODE",
			entries: []helpEntry{
				{key: "p / s / f / r / d", action: "pick / squash / fixup / reword / drop"},
				{key: "J / K (+ / -)", action: "Reorder commit sequence"},
				{key: "enter", action: "Execute rebase"},
				{key: "esc", action: "Cancel rebase"},
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
				{key: "E", action: "Export command log (in log panel)"},
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

	renderSection := func(section helpSection, secWidth int) string {
		keyWidth := 0
		for _, entry := range section.entries {
			if w := lipgloss.Width(entry.key); w > keyWidth {
				keyWidth = w
			}
		}
		if keyWidth < 14 {
			keyWidth = 14
		}

		header := lipgloss.NewStyle().
			Foreground(ui.ColorCyan).
			Bold(true).
			Render(section.title)

		divLen := secWidth
		if divLen < 20 {
			divLen = 20
		}
		divider := ui.SubtleStyle.Render(strings.Repeat("─", divLen))

		lines := []string{header, divider}

		actionWidth := secWidth - keyWidth - 2
		if actionWidth < 12 {
			actionWidth = 12
		}

		for _, entry := range section.entries {
			wrappedAction := wrapPlainText(entry.action, actionWidth)
			keyPadded := fmt.Sprintf("%-*s", keyWidth, entry.key)
			for i, wrapped := range wrappedAction {
				if i == 0 {
					lines = append(lines, ui.FooterKeyStyle.Render(keyPadded)+"  "+ui.ValueStyle.Render(wrapped))
				} else {
					lines = append(lines, strings.Repeat(" ", keyWidth)+"  "+ui.ValueStyle.Render(wrapped))
				}
			}
		}
		return strings.Join(lines, "\n")
	}

	columnCount := 1
	contentWidth := width
	if contentWidth >= 70 {
		columnCount = 2
	}
	if contentWidth >= 96 {
		columnCount = 3
	}

	columnWidth := contentWidth
	if columnCount > 1 {
		columnWidth = (contentWidth - 4*(columnCount-1)) / columnCount
	}
	if columnWidth < 32 && contentWidth >= 32 {
		columnWidth = 32
	}

	columnSections := make([][]string, columnCount)
	columnHeights := make([]int, columnCount)
	for _, section := range sections {
		target := 0
		for i := 1; i < columnCount; i++ {
			if columnHeights[i] < columnHeights[target] {
				target = i
			}
		}
		rendered := renderSection(section, columnWidth)
		columnSections[target] = append(columnSections[target], rendered)
		columnHeights[target] += lipgloss.Height(rendered) + 2
	}

	columns := make([]string, 0, columnCount)
	for _, sections := range columnSections {
		columns = append(columns, lipgloss.NewStyle().Width(columnWidth).Render(strings.Join(sections, "\n\n")))
	}

	content := columns[0]
	for i := 1; i < len(columns); i++ {
		content = lipgloss.JoinHorizontal(lipgloss.Top, content, "    ", columns[i])
	}

	return content
}

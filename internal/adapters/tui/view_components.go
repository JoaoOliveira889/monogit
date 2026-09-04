package tui

import (
	"fmt"
	"hash/fnv"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"

	"github.com/JoaoOliveira889/monogit/internal/pkg/ui"
)

func (m *Model) currentModeBadge() string {
	switch {
	case m.searchMode:
		return lipgloss.NewStyle().Background(ui.ColorOrange).Foreground(ui.ColorBg).Bold(true).Render(" SEARCH ")
	case m.showConflicts:
		return lipgloss.NewStyle().Background(ui.ColorError).Foreground(ui.ColorBg).Bold(true).Render(" CONFLICTS ")
	case m.showRebase:
		return lipgloss.NewStyle().Background(ui.ColorWarning).Foreground(ui.ColorBg).Bold(true).Render(" REBASE ")
	case m.showBranches:
		return lipgloss.NewStyle().Background(ui.ColorCyan).Foreground(ui.ColorBg).Bold(true).Render(" BRANCHES ")
	case m.showFiles:
		if m.activePanel == DiffPanel {
			return lipgloss.NewStyle().Background(ui.ColorAccent).Foreground(ui.ColorBg).Bold(true).Render(" DIFF ")
		}
		return lipgloss.NewStyle().Background(ui.ColorSuccess).Foreground(ui.ColorBg).Bold(true).Render(" FILES ")
	case m.showStashes:
		return lipgloss.NewStyle().Background(ui.ColorIndigo).Foreground(ui.ColorBg).Bold(true).Render(" STASH ")
	case m.activePanel == ConfigPanel:
		return lipgloss.NewStyle().Background(ui.ColorAmber).Foreground(ui.ColorBg).Bold(true).Render(" CONFIG ")
	case m.activePanel == CommandLogPanel:
		return lipgloss.NewStyle().Background(ui.ColorCyan).Foreground(ui.ColorBg).Bold(true).Render(" LOGS ")
	default:
		return ""
	}
}

func (m *Model) renderHeader() string {
	var brand string
	mode := m.currentModeBadge()
	switch {
	case m.width < 35:
		brand = ui.BrandMonoStyle.Render("MG")
	case m.width < 70 || mode == "":
		brand = renderBrandWordmark(true)
	default:
		brand = mode + " " + renderBrandWordmark(true)
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
	leftHeader := " " + brand + sep + healthSummary
	if loading != "" {
		leftHeader += "  " + loading
	}

	rightHeader := ""
	if m.statusMsg != "" {
		rightHeader = m.renderHeaderStatusBar() + " "
	}

	headerLine := renderHeaderBetween(leftHeader, rightHeader, m.width)

	border := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ui.ColorBorder)).
		Render(strings.Repeat("─", m.width))

	return headerLine + "\n" + border
}

func renderHeaderBetween(left, right string, totalWidth int) string {
	leftW := lipgloss.Width(left)
	rightW := lipgloss.Width(right)
	if right == "" {
		return left
	}
	spaces := totalWidth - leftW - rightW
	if spaces < 2 {
		maxRightW := totalWidth - leftW - 2
		if maxRightW > 6 {
			right = truncateRunes(right, maxRightW)
			rightW = lipgloss.Width(right)
			spaces = totalWidth - leftW - rightW
		} else {
			return left
		}
	}
	return left + strings.Repeat(" ", spaces) + right
}

func (m *Model) renderWorkspaceHealth() string {
	total := len(m.repos)
	if total == 0 {
		return ui.SubtleStyle.Render("● No repos")
	}

	if !m.healthCache.valid || m.healthCache.total != total {
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
		m.healthCache = workspaceHealthStats{
			total:     total,
			clean:     cleanCount,
			dirty:     dirtyCount,
			ahead:     aheadCount,
			behind:    behindCount,
			conflicts: conflictCount,
			valid:     true,
		}
	}

	cleanCount := m.healthCache.clean
	dirtyCount := m.healthCache.dirty
	aheadCount := m.healthCache.ahead
	behindCount := m.healthCache.behind
	conflictCount := m.healthCache.conflicts

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
		return ui.CleanStyle.Render(m.statusMsg)
	case strings.HasPrefix(m.statusMsg, "✗") || strings.HasPrefix(m.statusMsg, "Error"):
		return ui.ErrorStyle.Render(m.statusMsg)
	case strings.HasPrefix(m.statusMsg, "⚠") || strings.HasPrefix(m.statusMsg, "Warn"):
		return ui.WarningStyle.Render(m.statusMsg)
	default:
		return ui.ValueStyle.Render(m.statusMsg)
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
				m.fmtKey("ctrl+d/u", "page"),
				m.fmtKey("y", "copy"),
				m.fmtKey("C", "compact"),
				m.fmtKey(altKeys("tab", "2"), "files"),
				m.fmtKey("1", "repos"),
			}
		} else {
			parts = []string{
				m.fmtKey("jk", "nav"),
				m.fmtKey("ctrl+d/u", "page"),
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
			m.fmtKey("ctrl+d/u", "page"),
			m.fmtKey("enter", "checkout"),
			m.fmtKey("M", "merge"),
			m.fmtKey("n", "new"),
			m.fmtKey("d", "delete"),
			m.fmtKey(altKeys("h", "esc"), "back"),
		}
	case m.showStashes:
		parts = []string{
			m.fmtKey("jk", "nav"),
			m.fmtKey(altKeys("p", "enter"), "pop"),
			m.fmtKey("a", "apply"),
			m.fmtKey("d", "drop"),
			m.fmtKey(altKeys("h", "esc"), "back"),
		}
	case m.showConflicts:
		parts = []string{
			m.fmtKey("jk", "nav"),
			m.fmtKey("enter", "resolve"),
			m.fmtKey(altKeys("h", "esc"), "back"),
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
			m.fmtKey("jk", "nav"),
			m.fmtKey("ctrl+d/u", "page"),
			m.fmtKey("enter/l", "details"),
			m.fmtKey("f", "fetch"),
			m.fmtKey("u", "push"),
			m.fmtKey("b", "branches"),
		}
	default:
		parts = []string{
			m.fmtKey("jk", "nav"),
			m.fmtKey("ctrl+d/u", "page"),
			m.fmtKey("enter", "details"),
			m.fmtKey("d", "diff"),
			m.fmtKey("y", "copy hash"),
			m.fmtKey("g", "graph"),
			m.fmtKey(altKeys("h", "esc"), "back"),
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
	if panelWidth < 32 {
		panelWidth = 32
	}

	innerWidth := panelWidth - 4
	if innerWidth < 24 {
		innerWidth = 24
	}

	title := lipgloss.JoinHorizontal(lipgloss.Bottom,
		renderBrandWordmark(true),
		" ",
		ui.BrandTitleStyle.Render("SHORTCUTS"),
	)

	maxViewportHeight := m.height - 8
	if maxViewportHeight < 6 {
		maxViewportHeight = 6
	}
	vpHeight := maxViewportHeight
	vpWidth := innerWidth - 2
	if vpWidth < 20 {
		vpWidth = 20
	}

	if m.helpViewport.Width != vpWidth || m.helpViewport.Height != vpHeight {
		m.helpViewport = viewport.New(vpWidth, vpHeight)
	} else {
		m.helpViewport.Width = vpWidth
		m.helpViewport.Height = vpHeight
	}

	body := m.renderHelpMenu(vpWidth, vpHeight)
	contentHeight := lipgloss.Height(body)
	if contentHeight < vpHeight {
		vpHeight = contentHeight
		if vpHeight < 5 {
			vpHeight = 5
		}
		m.helpViewport.Height = vpHeight
	}
	m.helpViewport.SetContent(body)

	titleBar := lipgloss.NewStyle().Align(lipgloss.Center).Width(innerWidth).Render(title)
	footerHint := lipgloss.NewStyle().Align(lipgloss.Center).Width(innerWidth).Render(
		ui.SubtleStyle.Render("esc / ctrl+p close  •  ↑/↓ / jk scroll  •  g / G top / bottom"),
	)

	var helpBody string
	if contentHeight <= vpHeight {
		helpBody = m.helpViewport.View()
	} else {
		helpBody = renderViewportWithScrollbar(m.helpViewport, true)
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		titleBar,
		"",
		helpBody,
		"",
		footerHint,
	)

	panelStyle := ui.ActivePanelStyle.
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ui.ColorCyan)).
		Width(panelWidth).
		Padding(1, 2)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, panelStyle.Render(content))
}

func padRight(s string, width int) string {
	w := lipgloss.Width(s)
	if w >= width {
		return s
	}
	return s + strings.Repeat(" ", width-w)
}

func (m *Model) renderHelpMenu(width, height int) string {
	type helpEntry struct {
		key    string
		action string
	}
	type helpCard struct {
		title   string
		entries []helpEntry
	}

	threeCards := []helpCard{
		{
			title: "NAVIGATION",
			entries: []helpEntry{
				{key: "jk | ↑↓", action: "Move cursor"},
				{key: "ctrl+d/u", action: "Half-page scroll"},
				{key: "G | home", action: "Jump top / bottom"},
				{key: "hl | ←→", action: "Switch panel focus"},
				{key: "< | >", action: "Resize left panel"},
				{key: "1 | 2 | 3", action: "Jump to panel 1/2/3"},
				{key: "tab", action: "Cycle panel focus"},
				{key: "v | y", action: "Visual select / copy"},
				{key: "/ | ctrl+f", action: "Search repositories"},
				{key: "esc", action: "Back / cancel modal"},
				{key: "q | ctrl+c", action: "Quit MonoGit"},
			},
		},
		{
			title: "REPOSITORY & COMMIT WIZARD",
			entries: []helpEntry{
				{key: "enter | l", action: "Open details & log"},
				{key: "f | F", action: "Fetch current / all"},
				{key: "p | P", action: "Pull current / all"},
				{key: "u | U", action: "Push current / all"},
				{key: "c", action: "Start commit wizard"},
				{key: "z", action: "Undo last commit"},
				{key: "a", action: "Stage all files"},
				{key: "v", action: "Select files"},
				{key: "space", action: "Toggle selection"},
				{key: "enter", action: "Confirm / advance"},
				{key: "B", action: "Bulk checkout branch"},
			},
		},
		{
			title: "BRANCH, DIFF & STASH MODE",
			entries: []helpEntry{
				{key: "b", action: "Open branch manager"},
				{key: "enter", action: "Checkout branch"},
				{key: "M", action: "Merge into HEAD"},
				{key: "n | d", action: "Create / del branch"},
				{key: "m", action: "Resolve conflicts"},
				{key: "R", action: "Interactive rebase"},
				{key: "d", action: "View file diff"},
				{key: "C", action: "Toggle compact diff"},
				{key: "p | enter", action: "Pop stash"},
				{key: "a | d", action: "Apply / drop stash"},
				{key: "o | E", action: "Command log / export"},
			},
		},
	}

	twoCards := []helpCard{
		{
			title: "NAVIGATION & REPOSITORY",
			entries: []helpEntry{
				{key: "jk | ↑↓", action: "Move cursor"},
				{key: "ctrl+d/u", action: "Half-page scroll"},
				{key: "G | home", action: "Jump top / bottom"},
				{key: "hl | ←→", action: "Switch panel focus"},
				{key: "< | >", action: "Resize left panel"},
				{key: "1 | 2 | 3", action: "Jump to panel 1/2/3"},
				{key: "tab", action: "Cycle panel focus"},
				{key: "v | y", action: "Visual select / copy"},
				{key: "/ | ctrl+f", action: "Search repositories"},
				{key: "esc", action: "Back / cancel modal"},
				{key: "q | ctrl+c", action: "Quit MonoGit"},
				{key: "enter | l", action: "Open details & log"},
				{key: "f | F", action: "Fetch current / all"},
				{key: "p | P", action: "Pull current / all"},
				{key: "u | U", action: "Push current / all"},
				{key: "c", action: "Start commit wizard"},
				{key: "z", action: "Undo last commit"},
				{key: "B", action: "Bulk checkout branch"},
			},
		},
		{
			title: "BRANCHES, DIFF & STASH MODE",
			entries: []helpEntry{
				{key: "b", action: "Open branch manager"},
				{key: "enter", action: "Checkout branch"},
				{key: "M", action: "Merge into HEAD"},
				{key: "n | d", action: "Create / del branch"},
				{key: "m", action: "Resolve conflicts"},
				{key: "R", action: "Interactive rebase"},
				{key: "a", action: "Stage all files"},
				{key: "v", action: "Select files"},
				{key: "space", action: "Toggle selection"},
				{key: "d", action: "View file diff"},
				{key: "C", action: "Toggle compact diff"},
				{key: "p | enter", action: "Pop stash"},
				{key: "a | d", action: "Apply / drop stash"},
				{key: "o | E", action: "Command log / export"},
			},
		},
	}

	cardCount := 1
	cards := []helpCard{twoCards[0]}
	if width >= 65 {
		cardCount = 2
		cards = twoCards
	}
	if width >= 98 {
		cardCount = 3
		cards = threeCards
	}

	gap := 2
	cardWidth := width
	if cardCount > 1 {
		cardWidth = (width - gap*(cardCount-1)) / cardCount
	}
	if cardWidth < 28 && width >= 28 {
		cardWidth = 28
	}

	renderCard := func(card helpCard, cWidth int) string {
		contentW := cWidth - 4
		if contentW < 14 {
			contentW = 14
		}

		keyWidth := 0
		for _, e := range card.entries {
			if w := lipgloss.Width(e.key); w > keyWidth {
				keyWidth = w
			}
		}
		if keyWidth < 6 {
			keyWidth = 6
		}
		if keyWidth > 11 {
			keyWidth = 11
		}

		actW := contentW - keyWidth - 3
		if actW < 8 {
			actW = 8
		}

		titleStyle := lipgloss.NewStyle().Foreground(ui.ColorCyan).Bold(true)
		headerKeyStyle := ui.LabelStyle
		headerActStyle := ui.LabelStyle
		keyColStyle := ui.FooterKeyStyle
		valColStyle := ui.ValueStyle
		sepStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ui.ColorBorder))

		title := titleStyle.Render(card.title)
		titlePadding := contentW - lipgloss.Width(title)
		if titlePadding < 0 {
			titlePadding = 0
		}
		titleLine := title + strings.Repeat(" ", titlePadding)
		titleDiv := sepStyle.Render(strings.Repeat("─", contentW))

		hKey := padRight("KEY", keyWidth)
		hAct := padRight("ACTION", actW)
		headerLine := headerKeyStyle.Render(hKey) + " " + sepStyle.Render("│") + " " + headerActStyle.Render(hAct)
		divLine := sepStyle.Render(strings.Repeat("─", keyWidth+1) + "┼" + strings.Repeat("─", actW+1))

		lines := []string{titleLine, titleDiv, headerLine, divLine}

		for _, e := range card.entries {
			kPadded := padRight(e.key, keyWidth)
			if lipgloss.Width(e.action) <= actW {
				actPadded := padRight(e.action, actW)
				row := keyColStyle.Render(kPadded) + " " + sepStyle.Render("│") + " " + valColStyle.Render(actPadded)
				lines = append(lines, row)
			} else {
				wAct := wrapPlainText(e.action, actW)
				for i, wrapped := range wAct {
					actPadded := padRight(wrapped, actW)
					if i == 0 {
						row := keyColStyle.Render(kPadded) + " " + sepStyle.Render("│") + " " + valColStyle.Render(actPadded)
						lines = append(lines, row)
					} else {
						emptyK := strings.Repeat(" ", keyWidth)
						row := emptyK + " " + sepStyle.Render("│") + " " + valColStyle.Render(actPadded)
						lines = append(lines, row)
					}
				}
			}
		}

		cardStyle := lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(ui.ColorBorder)).
			Width(cWidth - 2).
			Padding(0, 1)

		return cardStyle.Render(strings.Join(lines, "\n"))
	}

	renderedCards := make([]string, 0, len(cards))
	for _, card := range cards {
		renderedCards = append(renderedCards, renderCard(card, cardWidth))
	}

	content := renderedCards[0]
	for i := 1; i < len(renderedCards); i++ {
		content = lipgloss.JoinHorizontal(lipgloss.Top, content, "  ", renderedCards[i])
	}

	return content
}


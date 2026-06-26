package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"

	"github.com/JoaoOliveira889/monogit/internal/domain"
	"github.com/JoaoOliveira889/monogit/internal/pkg/ui"
)

func (m *Model) renderBody() string {
	leftWidth := m.leftPanelWidth()
	rightWidth := m.rightPanelWidth()

	headerHeight := 2
	if m.statusMsg != "" {
		headerHeight = 3
	}
	footerHeight := 1
	bodyHeight := m.height - headerHeight - footerHeight
	if bodyHeight < 5 {
		bodyHeight = 5
	}

	left := m.renderRepoList(leftWidth, bodyHeight)
	right := m.renderDetailPanel(rightWidth, bodyHeight)

	panels := lipgloss.JoinHorizontal(lipgloss.Top, left, right)

	return panels
}

func (m *Model) renderTitledPanel(width, height int, title string, content string, active bool, accent lipgloss.Color) string {
	style := ui.LeftPanelStyle
	if active {
		style = ui.ActivePanelStyle.BorderStyle(lipgloss.DoubleBorder())
	}

	style = style.BorderTop(false)
	borderColor := accent
	if active {
		borderColor = lipgloss.Color(ui.ColorHighlight)
	}
	style = style.BorderForeground(borderColor)

	border := style.GetBorderStyle()
	maxTitleWidth := width - 6
	if maxTitleWidth < 5 {
		maxTitleWidth = 5
	}
	truncatedTitle := title
	titleRunes := []rune(title)
	if len(titleRunes) > maxTitleWidth {
		truncatedTitle = string(titleRunes[:maxTitleWidth-3]) + "..."
	}

	titleText := fmt.Sprintf("─[%s]─", truncatedTitle)
	titleWidth := lipgloss.Width(titleText)

	repeatCount := width - titleWidth - 2
	if repeatCount < 0 {
		repeatCount = 0
	}
	topLine := border.TopLeft + titleText + strings.Repeat(border.Top, repeatCount) + border.TopRight

	topLineStyle := lipgloss.NewStyle().Foreground(borderColor)
	if !active {
		topLineStyle = topLineStyle.Foreground(accent)
	}
	if active {
		topLineStyle = topLineStyle.Bold(true)
	}
	styledTopLine := topLineStyle.Render(topLine)

	innerWidth := width - 2
	if innerWidth < 0 {
		innerWidth = 0
	}
	innerHeight := height - 2
	if innerHeight < 0 {
		innerHeight = 0
	}

	panel := style.
		Width(innerWidth).
		Height(innerHeight).
		Render(content)

	return lipgloss.JoinVertical(lipgloss.Left, styledTopLine, panel)
}

func (m *Model) renderRepoList(width, height int) string {
	content := renderViewportWithScrollbar(m.repoViewport, m.activePanel == RepoPanel)
	title := m.getPanelNumber(RepoPanel)
	if m.tagFilterActive && len(m.tagFilter) > 0 {
		title += " [" + strings.Join(m.tagFilter, ", ") + "]"
	}
	if query := m.searchFilterQuery(); query != "" {
		title += " [" + query + "]"
	}
	if m.searchMode {
		searchSection := m.renderSearchSection(width)
		content = lipgloss.JoinVertical(lipgloss.Left, searchSection, content)
	}

	accent := lipgloss.Color(ui.ColorMono)
	return m.renderTitledPanel(width, height, title, content, m.activePanel == RepoPanel, accent)
}

func (m *Model) renderRepoLine(index int, r domain.Repository, maxWidth int) string {
	selected := index == m.cursor
	selectedRange := m.lineSelected(RepoPanel, index)
	bg := ui.ColorBg
	if selected || selectedRange {
		bg = ui.ColorHighlight
	}

	var indicators []string
	indicatorStyle := lipgloss.NewStyle().Background(bg).Bold(true)

	for _, badge := range m.repoHealthBadges(r, bg) {
		indicators = append(indicators, badge)
	}

	indicators = append(indicators, indicatorStyle.Foreground(ui.ColorSuccess).Render(fmt.Sprintf("%s%d", ui.IconAhead, r.Ahead)))
	indicators = append(indicators, indicatorStyle.Foreground(ui.ColorWarning).Render(fmt.Sprintf("%s%d", ui.IconBehind, r.Behind)))

	if r.IsDirty {
		indicators = append(indicators, indicatorStyle.Foreground(ui.ColorError).Render(ui.IconDirty))
	} else if r.Branch != "" {
		indicators = append(indicators, indicatorStyle.Foreground(ui.ColorSuccess).Render(ui.IconClean))
	}

	space := lipgloss.NewStyle().Background(bg).Render(" ")
	indicatorStr := strings.Join(indicators, space)

	var prefix string
	if selected {
		prefix = lipgloss.NewStyle().Foreground(ui.ColorHighlight).Background(ui.ColorBg).Render("▌") +
			lipgloss.NewStyle().Background(bg).Render(" ")
	} else {
		prefix = lipgloss.NewStyle().Background(bg).Render("  ")
	}
	midSpace := lipgloss.NewStyle().Background(bg).Render(" ")

	availableNameWidth := maxWidth - lipgloss.Width(prefix) - lipgloss.Width(midSpace) - lipgloss.Width(indicatorStr)
	if availableNameWidth < 5 {
		availableNameWidth = 5
	}

	var nameStr string
	if r.Branch != "" {
		branchTextLen := len(r.Branch) + 3 // for " (" and ")"
		if branchTextLen >= availableNameWidth {
			combined := r.Name + " (" + r.Branch + ")"
			if len(combined) > availableNameWidth {
				combined = "…" + combined[len(combined)-availableNameWidth+1:]
			}
			nameStyle := lipgloss.NewStyle().Foreground(ui.ColorFg)
			if selected || selectedRange {
				nameStyle = nameStyle.Background(bg).Foreground(ui.ColorBg).Bold(true)
			}
			nameStr = nameStyle.Width(availableNameWidth).Render(combined)
		} else {
			allowedRepoWidth := availableNameWidth - branchTextLen
			repoName := r.Name
			if len(repoName) > allowedRepoWidth {
				repoName = repoName[:allowedRepoWidth-1] + "…"
			}
			var repoStr, branchStr string
			if selected || selectedRange {
				repoStr = lipgloss.NewStyle().Background(bg).Foreground(ui.ColorBg).Bold(true).Render(repoName)
				branchStr = lipgloss.NewStyle().Background(bg).Foreground(ui.ColorBg).Render(" (") +
					lipgloss.NewStyle().Background(bg).Foreground(ui.ColorSelected).Bold(true).Render(r.Branch) +
					lipgloss.NewStyle().Background(bg).Foreground(ui.ColorBg).Render(")")
			} else {
				repoStr = lipgloss.NewStyle().Foreground(ui.ColorFg).Render(repoName)
				branchStr = lipgloss.NewStyle().Foreground(ui.ColorSubtle).Render(" (") +
					ui.BranchStyle.Render(r.Branch) +
					lipgloss.NewStyle().Foreground(ui.ColorSubtle).Render(")")
			}
			printedWidth := lipgloss.Width(repoStr + branchStr)
			nameStr = repoStr + branchStr
			if printedWidth < availableNameWidth {
				nameStr += lipgloss.NewStyle().Background(bg).Render(strings.Repeat(" ", availableNameWidth-printedWidth))
			}
		}
	} else {
		repoName := r.Name
		if len(repoName) > availableNameWidth {
			repoName = repoName[:availableNameWidth-1] + "…"
		}
		nameStyle := lipgloss.NewStyle().Foreground(ui.ColorFg)
		if selected || selectedRange {
			nameStyle = nameStyle.Background(bg).Foreground(ui.ColorBg).Bold(true)
		}
		nameStr = nameStyle.Width(availableNameWidth).Render(repoName)
	}

	line := prefix + nameStr + midSpace + indicatorStr

	padLen := maxWidth - lipgloss.Width(line)
	if padLen > 0 {
		line += lipgloss.NewStyle().Background(bg).Render(strings.Repeat(" ", padLen))
	}

	return line
}

func (m *Model) repoHealthBadges(r domain.Repository, bg lipgloss.Color) []string {
	indicatorStyle := lipgloss.NewStyle().Background(bg).Bold(true)

	fgWarning := ui.ColorWarning
	fgAmber := ui.ColorAmber
	fgError := ui.ColorError
	fgOrange := ui.ColorOrange
	fgCyan := ui.ColorCyan

	if bg == ui.ColorHighlight {
		fg := ui.ColorBg
		fgWarning = fg
		fgAmber = fg
		fgError = fg
		fgOrange = fg
		fgCyan = fg
	}

	var badges []string
	if r.IsDetached {
		badges = append(badges, indicatorStyle.Foreground(fgWarning).Render("DET"))
	}
	if !r.IsDetached && !r.HasUpstream && r.Branch != "" {
		badges = append(badges, indicatorStyle.Foreground(fgAmber).Render("UP"))
	}
	if r.HasConflicts {
		badges = append(badges, indicatorStyle.Foreground(fgError).Render("CF"))
	}
	if r.IsStale {
		badges = append(badges, indicatorStyle.Foreground(fgOrange).Render("ST"))
	}
	if r.HasUnpushedTag {
		badges = append(badges, indicatorStyle.Foreground(fgCyan).Render("TG"))
	}
	return badges
}

func (m *Model) renderDetailPanel(width, height int) string {
	r := m.selectedRepo()

	if r == nil {
		return ui.RightPanelStyle.Render("No repository selected")
	}

	var panelNum, panelLabel string
	if m.activePanel == CommandLogPanel {
		panelNum = m.getPanelNumber(CommandLogPanel)
		panelLabel = "Command Log"
	} else if m.activePanel == ConfigPanel {
		panelNum = m.getPanelNumber(ConfigPanel)
		panelLabel = "Configuration"
	} else if m.showConflicts {
		panelNum = m.getPanelNumber(ConflictPanel)
		panelLabel = "Conflicts"
	} else {
		panelNum = m.getPanelNumber(LogPanel)

		var label string
		if m.showFiles {
			label = "File Selection"
		} else if m.showBranches {
			label = "Branches"
		} else if m.showStashes {
			label = "Stashes"
		} else {
			label = r.Name
		}
		panelLabel = label
	}

	var content string
	if m.activePanel == CommandLogPanel {
		content = renderViewportWithScrollbar(m.logViewport, m.activePanel == CommandLogPanel)
	} else if m.activePanel == ConfigPanel {
		content = m.renderConfigPanel(width)
	} else if m.showConflicts {
		content = m.renderConflictList(width)
	} else if m.showFiles {
		listContent := renderViewportWithScrollbar(m.fileViewport, m.activePanel == LogPanel || m.activePanel == DiffPanel)

		diffTitleStyle := ui.DiffTabStyle(m.activePanel == DiffPanel)
		if m.activePanel == DiffPanel {
			diffTitleStyle = diffTitleStyle.Bold(true)
		}

		diffFileName := ""
		if m.fileCursor < len(m.files) {
			diffFileName = " — " + m.files[m.fileCursor].Name
			if len(diffFileName) > 40 {
				diffFileName = " — …" + diffFileName[len(diffFileName)-36:]
			}
		}
		diffHeader := diffTitleStyle.Width(width - 2).Render("[" + m.getPanelNumber(DiffPanel) + "] Diff" + diffFileName)

		var diffContent string
		if m.compactDiff {
			diffContent = m.renderCompactDiffContent()
		} else if m.diffFetching {
			diffContent = ui.SpinnerStyle.Render("   " + m.spinnerView() + " Loading diff...")
		} else if m.currentDiff == "" {
			diffContent = ui.SubtleStyle.Render("   No diff available")
		} else {
			diffContent = renderViewportWithScrollbar(m.diffViewport, m.activePanel == DiffPanel)
		}

		content = lipgloss.JoinVertical(lipgloss.Left,
			listContent,
			diffHeader,
			diffContent,
		)
	} else if m.showBranches {
		content = m.renderBranchesList(width)
	} else if m.showStashes {
		content = m.renderStashList(width)
	} else {
		content = renderViewportWithScrollbar(m.viewport, m.activePanel == LogPanel)
	}

	if m.tagAssignModal {
		content = m.renderRepoTagsSection(width)
	} else if m.activePanel != CommandLogPanel && !m.showFiles && !m.showBranches && !m.showStashes && !m.showConflicts {
		tagsSection := m.renderRepoTagsSection(width)
		if tagsSection != "" {
			content = lipgloss.JoinVertical(lipgloss.Left, tagsSection, content)
		}
	}

	content = clipRenderedContent(content, height-2)

	active := m.activePanel == LogPanel || m.activePanel == DiffPanel || m.activePanel == CommandLogPanel || m.activePanel == ConflictPanel || m.tagAssignModal
	accent := lipgloss.Color(ui.ColorGit)
	return m.renderTitledPanel(width, height, panelNum+"-"+panelLabel, content, active, accent)
}

func clipRenderedContent(content string, maxLines int) string {
	if maxLines <= 0 {
		return ""
	}
	lines := strings.Split(content, "\n")
	if len(lines) <= maxLines {
		return content
	}
	return strings.Join(lines[:maxLines], "\n")
}

func (m *Model) renderBeautifiedLog(log string, width int) string {
	if log == "" || log == "(no commits yet)" {
		return "  " + ui.SubtleStyle.Render(log)
	}

	var lines []string
	for _, rawLine := range strings.Split(log, "\n") {
		if rawLine == "" {
			continue
		}

		parts := strings.Split(rawLine, "||")
		graphAndHash := parts[0]

		var hash, graphPart string
		lastSpace := strings.LastIndex(graphAndHash, " ")
		if lastSpace != -1 {
			graphPart = graphAndHash[:lastSpace+1]
			hash = graphAndHash[lastSpace+1:]
		} else {
			hash = graphAndHash
		}

		beautifiedGraph := ""
		for i, char := range graphPart {
			style := ui.GraphCharStyles[i%len(ui.GraphCharStyles)]
			switch char {
			case '*':
				beautifiedGraph += style.Copy().Bold(true).Render("●")
			case '|':
				beautifiedGraph += style.Render("│")
			case '/':
				beautifiedGraph += style.Render("╯")
			case '\\':
				beautifiedGraph += style.Render("╰")
			case '_':
				beautifiedGraph += style.Render("─")
			case ' ':
				beautifiedGraph += " "
			default:
				beautifiedGraph += style.Render(string(char))
			}
		}

		if len(parts) >= 5 {
			prefix := fmt.Sprintf("  %s %s ", beautifiedGraph, ui.SubtleStyle.Render(hash))
			prefixWidth := lipgloss.Width(prefix)

			suffix := fmt.Sprintf("%s  %s  %s", parts[2], parts[3], parts[4])

			availableWidth := width - prefixWidth
			if availableWidth < 10 {
				availableWidth = 10
			}

			wrapped := wrapPlainText(suffix, availableWidth)
			for i, w := range wrapped {
				if i == 0 {
					lines = append(lines, prefix+ui.ValueStyle.Render(w))
				} else {
					indent := strings.Repeat(" ", prefixWidth)
					lines = append(lines, indent+ui.ValueStyle.Render(w))
				}
			}
		} else {
			lines = append(lines, "  "+ui.ValueStyle.Render(rawLine))
		}
	}

	return strings.Join(lines, "\n")
}

func (m *Model) renderBeautifiedDiff(diff string) string {
	if diff == "" {
		return ""
	}

	lines := strings.Split(diff, "\n")
	var beautified []string

	for _, line := range lines {
		switch {
		case strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++"):
			beautified = append(beautified, ui.DiffAddStyle.Render(line))
		case strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---"):
			beautified = append(beautified, ui.DiffDelStyle.Render(line))
		case strings.HasPrefix(line, "@@"):
			beautified = append(beautified, ui.DiffHunkStyle.Render(line))
		default:
			beautified = append(beautified, line)
		}
	}

	return strings.Join(beautified, "\n")
}

func (m *Model) repoDetailCacheFor(repoPath string) *repoDetailCacheEntry {
	if entry, ok := m.detailCache[repoPath]; ok {
		return &entry
	}
	return nil
}

func (m *Model) renderViewportContent() string {
	r := m.selectedRepo()
	if r == nil {
		return ui.SubtleStyle.Render("No repository selected")
	}

	width := m.rightPanelWidth() - 2
	if width < 10 {
		width = 10
	}
	sections := make([]string, 0, 16)

	cache := m.repoDetailCacheFor(r.Path)

	statusStr := ui.CleanStyle.Render("Clean " + ui.IconClean)
	if r.IsDirty {
		statusStr = ui.DirtyStyle.Render("Modified " + ui.IconDirty)
	}

	sections = append(sections,
		ui.LabelStyle.Render("   Branch:  ")+ui.BranchStyle.Render(r.Branch),
		ui.LabelStyle.Render("   Status:  ")+statusStr,
	)

	healthLabels := m.repoHealthLabels(r)
	if len(healthLabels) == 0 {
		sections = append(sections, ui.LabelStyle.Render("   Health:  ")+ui.SuccessStyle.Render("healthy"))
	} else {
		sections = append(sections, ui.LabelStyle.Render("   Health:  ")+ui.WarningStyle.Render(strings.Join(healthLabels, " | ")))
	}

	if r.Tagging {
		sections = append(sections, ui.ValueStyle.Render("   "+m.spinnerView()+" Tagging & Deploying..."))
	}

	modifiedCount := m.cachedModifiedCount
	untrackedCount := m.cachedUntrackedCount
	if cache != nil {
		modifiedCount = cache.modifiedCount
		untrackedCount = cache.untrackedCount
	}

	sections = append(sections,
		fmt.Sprintf("   %s %d modified, %d untracked", ui.IconSpace, modifiedCount, untrackedCount),
	)

	aheadStr := ui.SubtleStyle.Render("0")
	if r.Ahead > 0 {
		aheadStr = ui.AheadStyle.Render(fmt.Sprintf("%d pushes pending %s", r.Ahead, ui.IconAhead))
	}
	behindStr := ui.SubtleStyle.Render("0")
	if r.Behind > 0 {
		behindStr = ui.BehindStyle.Render(fmt.Sprintf("%d pulls pending %s", r.Behind, ui.IconBehind))
	}

	sections = append(sections,
		ui.LabelStyle.Render("   Ahead:   ")+aheadStr,
		ui.LabelStyle.Render("   Behind:  ")+behindStr,
	)

	lastCommit := m.cachedLastCommit
	if cache != nil {
		lastCommit = cache.lastCommit
	}

	if lastCommit != "" && lastCommit != "(no commits yet)" {
		sections = append(sections, "", ui.LabelStyle.Render("   Last Commit:"))
		parts := strings.Split(lastCommit, " ")
		if len(parts) > 1 {
			hash := ui.SubtleStyle.Render(parts[0])
			msg := strings.Join(parts[1:], " ")
			for i, line := range wrapPlainText(msg, width-7) {
				if i == 0 {
					sections = append(sections, fmt.Sprintf("     %s %s", hash, ui.ValueStyle.Render(line)))
				} else {
					sections = append(sections, "     "+ui.ValueStyle.Render(line))
				}
			}
		}
	}

	if r.Error != "" {
		sections = append(sections, "", ui.ErrorStyle.Render("  ✗ Error: "+r.Error))
	}

	title := "Commit Graph"
	if !m.viewGraph {
		title = "Recent Commits"
	}
	sections = append(sections, "", ui.LabelStyle.Render("   "+title+":"))
	sections = append(sections, ui.SubtleStyle.Render("   "+strings.Repeat("─", 40)))

	if m.detailLoading && m.cachedDetailFor == r.Path {
		sections = append(sections, ui.SubtleStyle.Render("   Loading repository details..."))
	}

	log := m.cachedLog
	showLog := m.cachedLogFor == r.Path && m.cachedLog != ""
	if !showLog && cache != nil && cache.log != "" {
		log = cache.log
		showLog = true
	}

	if showLog {
		sections = append(sections, m.renderBeautifiedLog(log, width))
	} else {
		sections = append(sections, ui.SubtleStyle.Render("   Loading commit history..."))
	}

	if r.LastOutput != "" {
		sections = append(sections, "", ui.LabelStyle.Render("   Last Output:"))
		sections = append(sections, ui.SubtleStyle.Render("   "+strings.Repeat("─", 40)))
		for _, line := range strings.Split(r.LastOutput, "\n") {
			for _, wrapped := range wrapPlainText(line, width-4) {
				sections = append(sections, "   "+ui.ValueStyle.Render(wrapped))
			}
		}
	}

	return strings.Join(sections, "\n")
}

func (m *Model) repoHealthLabels(r *domain.Repository) []string {
	if r == nil {
		return nil
	}
	labels := make([]string, 0, 5)
	if r.IsDetached {
		labels = append(labels, "detached HEAD")
	}
	if !r.IsDetached && !r.HasUpstream && r.Branch != "" {
		labels = append(labels, "no upstream")
	}
	if r.HasConflicts {
		labels = append(labels, "merge conflicts")
	}
	if r.IsStale {
		labels = append(labels, "stale branch")
	}
	if r.HasUnpushedTag {
		labels = append(labels, "unpushed tag")
	}
	return labels
}

func (m *Model) renderRepoViewportContent() string {
	width := m.leftPanelWidth()
	repos := m.filteredRepos()
	if len(repos) == 0 {
		if len(m.repos) == 0 {
			return ui.SubtleStyle.Render("  No repositories found")
		}
		return ui.SubtleStyle.Render("  No repositories match the filter")
	}

	realIndex := make(map[string]int, len(m.repos))
	for i, r := range m.repos {
		realIndex[r.Path] = i
	}

	lines := make([]string, 0, len(repos))
	for _, r := range repos {
		idx, ok := realIndex[r.Path]
		if !ok {
			idx = 0
		}
		lines = append(lines, m.renderRepoLine(idx, r, width-2))
	}

	return strings.Join(lines, "\n")
}

func (m *Model) renderFileViewportContent() string {
	if !m.showFiles {
		return ""
	}

	width := m.rightPanelWidth()
	maxNameWidth := width - 15
	if maxNameWidth < 20 {
		maxNameWidth = 20
	}

	if len(m.files) == 0 {
		return ui.SubtleStyle.Render("  No modified files")
	}

	lines := make([]string, 0, len(m.files))
	for i, f := range m.files {
		lines = append(lines, m.renderFileListItem(i, f, width, maxNameWidth))
	}

	return strings.Join(lines, "\n")
}

func (m *Model) renderFileListItem(index int, f domain.FileStatus, width, maxNameWidth int) string {
	selected := index == m.fileCursor && m.activePanel != DiffPanel
	selectedRange := m.lineSelected(LogPanel, index) && m.showFiles
	bg := ui.ColorBg
	if selected || selectedRange {
		bg = ui.ColorHighlight
	}

	bgStyle := lipgloss.NewStyle().Background(bg)

	selectedInList := m.fileSelections[index]
	checkboxStr := "○"
	cbStyle := lipgloss.NewStyle().Background(bg).Foreground(ui.ColorSubtle)
	if selectedInList {
		checkboxStr = "●"
		cbStyle = cbStyle.Foreground(ui.ColorSuccess).Bold(true)
	}
	checkbox := cbStyle.Render(checkboxStr)

	statusIcon := " "
	statusStyle := lipgloss.NewStyle().Background(bg).Foreground(ui.ColorSubtle)
	if f.Untracked {
		statusIcon = "?"
		statusStyle = statusStyle.Foreground(ui.ColorError).Bold(true)
	} else if f.Modified {
		statusIcon = "M"
		statusStyle = statusStyle.Foreground(ui.ColorError).Bold(true)
	} else if f.Staged {
		statusIcon = "A"
		statusStyle = statusStyle.Foreground(ui.ColorSuccess).Bold(true)
	}
	statusInd := statusStyle.Render(statusIcon)

	name := f.Name
	if len(name) > maxNameWidth {
		name = "…" + name[len(name)-maxNameWidth+1:]
	}

	nameStyle := lipgloss.NewStyle().Foreground(ui.ColorFg)
	if selected || selectedRange {
		nameStyle = nameStyle.Background(bg).Foreground(ui.ColorBg).Bold(true)
	}
	nameStr := nameStyle.Render(name)

	lineContent := bgStyle.Render("    ") + checkbox + bgStyle.Render(" ") + statusInd + bgStyle.Render("  ") + nameStr

	padLen := (width - 2) - lipgloss.Width(lineContent)
	if padLen > 0 {
		lineContent += bgStyle.Render(strings.Repeat(" ", padLen))
	}

	return lineContent
}

func (m *Model) renderBranchesList(width int) string {
	if len(m.branches) == 0 {
		return ui.SubtleStyle.Render("  No branches found")
	}

	lines := make([]string, 0, len(m.branches))
	for i, b := range m.branches {
		selected := i == m.branchCursor
		selectedRange := m.lineSelected(LogPanel, i) && m.showBranches
		bg := ui.ColorBg
		if selected || selectedRange {
			bg = ui.ColorHighlight
		}

		bgStyle := lipgloss.NewStyle().Background(bg)

		prefix := "   "
		if b.IsCurrent {
			prefix = " ✓ "
		}
		if selected || selectedRange {
			prefix = bgStyle.Render(prefix)
		}

		nameStyle := lipgloss.NewStyle().Foreground(ui.ColorFg)
		if selected || selectedRange {
			nameStyle = ui.SelectedItemStyle
		}
		nameStr := nameStyle.Render(b.Name)

		indicators := []string{}
		if b.IsLocal {
			indicators = append(indicators, "local")
		}
		if b.IsRemote {
			indicators = append(indicators, "remote")
		}

		indicatorText := " (" + strings.Join(indicators, ", ") + ")"
		indicator := ui.SubtleStyle.Render(indicatorText)
		if selected || selectedRange {
			indicator = lipgloss.NewStyle().Background(bg).Foreground(ui.ColorBg).Render(indicator)
		}

		line := prefix + nameStr + indicator

		padLen := (width - 2) - lipgloss.Width(line)
		if padLen > 0 {
			padSpaces := strings.Repeat(" ", padLen)
			if selected || selectedRange {
				padSpaces = bgStyle.Render(padSpaces)
			}
			line += padSpaces
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func (m *Model) renderStashList(width int) string {
	if len(m.stashes) == 0 {
		return ui.SubtleStyle.Render("  No stashes found")
	}

	lines := make([]string, 0, len(m.stashes))
	for i, s := range m.stashes {
		selected := i == m.stashCursor
		selectedRange := m.lineSelected(LogPanel, i) && m.showStashes
		bg := ui.ColorBg
		if selected || selectedRange {
			bg = ui.ColorHighlight
		}

		bgStyle := lipgloss.NewStyle().Background(bg)

		prefix := "   "
		if selected || selectedRange {
			if m.stashFilesFocus {
				prefix = " • "
			} else {
				prefix = " > "
			}
			prefix = bgStyle.Render(prefix)
		}

		indexStyle := lipgloss.NewStyle().Foreground(ui.ColorHighlight)
		if selected || selectedRange {
			indexStyle = ui.SelectedItemStyle
		}
		indexStr := indexStyle.Render(fmt.Sprintf("stash@{%d}", s.Index))

		msgStyle := lipgloss.NewStyle().Foreground(ui.ColorFg)
		if selected || selectedRange {
			msgStyle = ui.SelectedItemStyle
		}
		msgStr := msgStyle.Render(" " + s.Message)

		line := prefix + indexStr + msgStr

		padLen := (width - 2) - lipgloss.Width(line)
		if padLen > 0 {
			padSpaces := strings.Repeat(" ", padLen)
			if selected || selectedRange {
				padSpaces = bgStyle.Render(padSpaces)
			}
			line += padSpaces
		}
		lines = append(lines, line)
	}

	if m.stashFiles != nil {
		lines = append(lines, "", ui.SubtleStyle.Render("  ── Files ──"))
		if len(m.stashFiles) == 0 {
			lines = append(lines, ui.SubtleStyle.Render("   (no files)"))
		} else {
			for i, f := range m.stashFiles {
				if m.stashFilesFocus && i == m.stashFileCursor {
					lines = append(lines, ui.SelectedItemStyle.Render("   ▸ "+f))
				} else {
					lines = append(lines, "     "+f)
				}
			}
		}
	}

	return strings.Join(lines, "\n")
}

func (m *Model) renderConflictList(width int) string {
	if len(m.conflictFiles) == 0 {
		return ui.SubtleStyle.Render("  No merge conflicts found")
	}

	lines := make([]string, 0, len(m.conflictFiles))
	for i, c := range m.conflictFiles {
		selected := i == m.conflictCursor
		selectedRange := m.lineSelected(ConflictPanel, i)
		bg := ui.ColorBg
		if selected || selectedRange {
			bg = ui.ColorHighlight
		}

		bgStyle := lipgloss.NewStyle().Background(bg)

		prefix := "   "
		if selected || selectedRange {
			prefix = " > "
			prefix = bgStyle.Render(prefix)
		}

		statusStyle := lipgloss.NewStyle().Foreground(ui.ColorError).Bold(true)
		statusStr := statusStyle.Render(" ⬌ " + c.Status + " ")

		nameStyle := lipgloss.NewStyle().Foreground(ui.ColorFg)
		if selected || selectedRange {
			nameStyle = ui.SelectedItemStyle
		}
		nameStr := nameStyle.Render(c.Name)

		line := prefix + statusStr + nameStr

		padLen := (width - 2) - lipgloss.Width(line)
		if padLen > 0 {
			padSpaces := strings.Repeat(" ", padLen)
			if selected || selectedRange {
				padSpaces = bgStyle.Render(padSpaces)
			}
			line += padSpaces
		}
		lines = append(lines, line)
	}

	lines = append(lines, "", ui.SubtleStyle.Render("  Press enter to resolve the selected file with the mergetool"))

	return strings.Join(lines, "\n")
}

func (m *Model) renderCompactDiffContent() string {
	if m.compactFetching {
		return ui.SpinnerStyle.Render("   " + m.spinnerView() + " Loading compact diff...")
	}

	if len(m.compactChanges) == 0 {
		return ui.SubtleStyle.Render("   No changes detected")
	}

	lines := make([]string, 0, len(m.compactChanges))
	for _, ch := range m.compactChanges {
		funcStyle := ui.ValueStyle.Foreground(ui.ColorGit)
		rangeStyle := ui.SubtleStyle
		location := ""
		if ch.LineRange != "" {
			location = " @" + ch.LineRange
		}
		line := fmt.Sprintf("  %s %s%s",
			ui.DiffHunkStyle.Render("Δ"),
			funcStyle.Render(ch.FunctionName),
			rangeStyle.Render(location),
		)
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func (m *Model) refreshFileViewport() {
	m.fileViewport.SetContent(m.renderFileViewportContent())

	if m.fileCursor < m.fileViewport.YOffset {
		m.fileViewport.YOffset = m.fileCursor
	} else if m.fileCursor >= m.fileViewport.YOffset+m.fileViewport.Height {
		m.fileViewport.YOffset = m.fileCursor - m.fileViewport.Height + 1
	}
}

func (m *Model) refreshLogViewport() {
	m.logViewport.SetContent(m.renderCommandLog(m.logViewport.Width))
	if m.commandLogCursor < m.logViewport.YOffset {
		m.logViewport.YOffset = m.commandLogCursor
	} else if m.commandLogCursor >= m.logViewport.YOffset+m.logViewport.Height {
		m.logViewport.YOffset = m.commandLogCursor - m.logViewport.Height + 1
	}
}

func (m *Model) renderCommandLog(width int) string {
	if len(m.commandLogs) == 0 {
		return ui.SubtleStyle.Render("  No commands executed yet.")
	}

	contentWidth := width - 6
	if contentWidth < 0 {
		contentWidth = 0
	}

	var sb strings.Builder
	for i, entry := range m.commandLogs {
		timeStr := entry.Time.Format("15:04:05")
		repoStr := ui.BranchStyle.Render(entry.RepoName)
		cmdStr := ui.ValueStyle.Render(entry.Command)
		selected := i == m.commandLogCursor
		selectedRange := m.lineSelected(CommandLogPanel, i)

		status := ui.CleanStyle.Render("SUCCESS")
		if entry.Error != nil {
			status = ui.ErrorStyle.Render("FAILED")
		}

		headLine := fmt.Sprintf("  [%s] %s > %s : %s", timeStr, repoStr, cmdStr, status)
		if selected || selectedRange {
			headLine = ui.SelectedItemStyle.Render(headLine)
		}
		fmt.Fprintln(&sb, headLine)

		if entry.Error != nil {
			errText := "Error: " + entry.Error.Error()
			wrappedErr := lipgloss.NewStyle().
				Foreground(ui.ColorError).
				Width(contentWidth).
				Render(errText)

			for _, line := range strings.Split(wrappedErr, "\n") {
				if selected || selectedRange {
					line = ui.SelectedItemStyle.Render(line)
				}
				fmt.Fprintf(&sb, "    %s\n", line)
			}
		}

		if entry.Output != "" {
			wrappedOutput := lipgloss.NewStyle().
				Foreground(ui.ColorSubtle).
				Width(contentWidth).
				Render(strings.TrimSpace(entry.Output))

			for _, line := range strings.Split(wrappedOutput, "\n") {
				if selected || selectedRange {
					line = ui.SelectedItemStyle.Render(line)
				}
				fmt.Fprintf(&sb, "    %s\n", line)
			}
		}

		if i < len(m.commandLogs)-1 {
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

func (m *Model) getPanelNumber(p Panel) string {
	visible := m.GetVisiblePanels()
	for i, vp := range visible {
		if vp == p {
			return fmt.Sprintf("%d", i+1)
		}
	}
	return "?"
}

func (m *Model) syncScrollPositions() {
	if m.repoViewport.Height > 0 {
		filtered := m.filteredRepos()
		filteredIdx := 0
		for i, r := range filtered {
			if r.Path == m.repos[m.cursor].Path {
				filteredIdx = i
				break
			}
		}
		if filteredIdx < m.repoViewport.YOffset {
			m.repoViewport.YOffset = filteredIdx
		} else if filteredIdx >= m.repoViewport.YOffset+m.repoViewport.Height {
			m.repoViewport.YOffset = filteredIdx - m.repoViewport.Height + 1
		}
	}

	if m.fileViewport.Height > 0 {
		if m.fileCursor < m.fileViewport.YOffset {
			m.fileViewport.YOffset = m.fileCursor
		} else if m.fileCursor >= m.fileViewport.YOffset+m.fileViewport.Height {
			m.fileViewport.YOffset = m.fileCursor - m.fileViewport.Height + 1
		}
	}
}

func renderViewportWithScrollbar(vp viewport.Model, active bool) string {
	view := vp.View()
	totalLines := vp.TotalLineCount()
	visibleLines := vp.Height
	yOffset := vp.YOffset

	// If there's no need for a scrollbar, pad the right side of each line with a space
	// to make the layout consistent with scrollbar-enabled views.
	if totalLines <= visibleLines {
		lines := strings.Split(view, "\n")
		for i, line := range lines {
			lines[i] = line + " "
		}
		return strings.Join(lines, "\n")
	}

	thumbHeight := visibleLines * visibleLines / totalLines
	if thumbHeight < 1 {
		thumbHeight = 1
	}

	scrollableRange := totalLines - visibleLines
	thumbRange := visibleLines - thumbHeight

	thumbStart := 0
	if scrollableRange > 0 {
		thumbStart = yOffset * thumbRange / scrollableRange
	}

	var sb strings.Builder
	for i := 0; i < visibleLines; i++ {
		if i >= thumbStart && i < thumbStart+thumbHeight {
			if active {
				sb.WriteString(ui.PointerStyle.Render("█"))
			} else {
				sb.WriteString(ui.SubtleStyle.Render("█"))
			}
		} else {
			sb.WriteString(ui.SubtleStyle.Render("░"))
		}
		if i < visibleLines-1 {
			sb.WriteString("\n")
		}
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, view, sb.String())
}

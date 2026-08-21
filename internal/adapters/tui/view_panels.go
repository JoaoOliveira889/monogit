package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"

	"github.com/JoaoOliveira889/monogit/internal/domain"
	"github.com/JoaoOliveira889/monogit/internal/pkg/ui"
)

func (m *Model) renderBody() string {
	leftWidth := m.leftPanelWidth()
	rightWidth := m.rightPanelWidth()

	headerHeight := 3
	footerHeight := 1
	bodyHeight := m.height - headerHeight - footerHeight
	if bodyHeight < 5 {
		bodyHeight = 5
	}
	if m.isCompactLayout() {
		if m.activePanel == RepoPanel {
			return m.renderRepoList(m.width, bodyHeight)
		}
		return m.renderDetailPanel(m.width, bodyHeight)
	}

	left := m.renderRepoList(leftWidth, bodyHeight)
	right := m.renderDetailPanel(rightWidth, bodyHeight)

	panels := lipgloss.JoinHorizontal(lipgloss.Top, left, right)

	return panels
}

func (m *Model) renderTitledPanel(width, height int, title string, content string, active bool, accent lipgloss.Color) string {
	borderColor := lipgloss.Color(ui.ColorBorder)

	border := lipgloss.RoundedBorder()

	maxTitleWidth := width - 6
	if maxTitleWidth < 5 {
		maxTitleWidth = 5
	}
	truncatedTitle := title
	titleRunes := []rune(title)
	if len(titleRunes) > maxTitleWidth {
		truncatedTitle = string(titleRunes[:maxTitleWidth-3]) + "..."
	}

	borderStyle := lipgloss.NewStyle().Foreground(borderColor)

	var titleStyled string
	if active {
		titleStyled = lipgloss.NewStyle().Foreground(accent).Bold(true).Render(truncatedTitle)
	} else {
		titleStyled = ui.SubtleStyle.Render(truncatedTitle)
	}

	titleLen := lipgloss.Width(truncatedTitle)
	// TopLine parts: TopLeft(1) + "─"(1) + title(titleLen) + "─"*repeatCount + TopRight(1) = width
	repeatCount := width - titleLen - 3
	if repeatCount < 0 {
		repeatCount = 0
	}

	topLine := borderStyle.Render(border.TopLeft+"─") +
		titleStyled +
		borderStyle.Render(strings.Repeat(border.Top, repeatCount)+border.TopRight)

	innerWidth := width - 2
	if innerWidth < 0 {
		innerWidth = 0
	}
	innerHeight := height - 2
	if innerHeight < 0 {
		innerHeight = 0
	}

	panelStyle := lipgloss.NewStyle().
		Border(border, false, true, true, true).
		BorderForeground(borderColor).
		Width(innerWidth).
		Height(innerHeight)

	panel := panelStyle.Render(content)

	return lipgloss.JoinVertical(lipgloss.Left, topLine, panel)
}

func (m *Model) renderRepoList(width, height int) string {
	content := renderViewportWithScrollbar(m.repoViewport, m.activePanel == RepoPanel)
	title := "[" + m.getPanelNumber(RepoPanel) + "] Repositories"
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

	accent := lipgloss.Color(ui.ColorCyan)
	return m.renderTitledPanel(width, height, title, content, m.activePanel == RepoPanel, accent)
}

func (m *Model) renderRepoLine(index int, r domain.Repository, maxWidth int) string {
	selected := index == m.cursor
	selectedRange := m.lineSelected(RepoPanel, index)
	isSelected := selected || selectedRange

	fullMode := maxWidth >= 46

	var statusIndicators []string
	if r.HasConflicts {
		if fullMode {
			statusIndicators = append(statusIndicators, ui.ErrorStyle.Render("!1 conflict"))
		} else {
			statusIndicators = append(statusIndicators, ui.ErrorStyle.Render("!1"))
		}
	}
	if r.Ahead > 0 {
		if fullMode {
			statusIndicators = append(statusIndicators, ui.AheadStyle.Render(fmt.Sprintf("%s%d ahead", ui.IconAhead, r.Ahead)))
		} else {
			statusIndicators = append(statusIndicators, ui.AheadStyle.Render(fmt.Sprintf("%s%d", ui.IconAhead, r.Ahead)))
		}
	}
	if r.Behind > 0 {
		if fullMode {
			statusIndicators = append(statusIndicators, ui.BehindStyle.Render(fmt.Sprintf("%s%d behind", ui.IconBehind, r.Behind)))
		} else {
			statusIndicators = append(statusIndicators, ui.BehindStyle.Render(fmt.Sprintf("%s%d", ui.IconBehind, r.Behind)))
		}
	}
	if r.IsDirty {
		dirtyCount := r.DirtyCount()
		if cache := m.repoDetailCacheFor(r.Path); cache != nil && cache.modifiedCount+cache.untrackedCount > 0 {
			dirtyCount = cache.modifiedCount + cache.untrackedCount
		}
		if dirtyCount > 0 {
			if fullMode {
				statusIndicators = append(statusIndicators, ui.DirtyStyle.Render(fmt.Sprintf("%s%d dirty", ui.IconDirty, dirtyCount)))
			} else {
				statusIndicators = append(statusIndicators, ui.DirtyStyle.Render(fmt.Sprintf("%s%d", ui.IconDirty, dirtyCount)))
			}
		} else {
			if fullMode {
				statusIndicators = append(statusIndicators, ui.DirtyStyle.Render(ui.IconDirty+" dirty"))
			} else {
				statusIndicators = append(statusIndicators, ui.DirtyStyle.Render(ui.IconDirty))
			}
		}
	}
	// Note: clean state is silent (no checkmark icon)

	healthBadges := m.repoHealthBadges(r, isSelected)
	var healthStr string
	if len(healthBadges) > 0 {
		healthStr = strings.Join(healthBadges, " ")
	}

	var statusStr string
	if len(statusIndicators) > 0 {
		sep := " · "
		if !fullMode {
			sep = " "
		}
		statusStr = strings.Join(statusIndicators, sep)
	}

	var metaStr string
	if healthStr != "" && statusStr != "" {
		metaStr = healthStr + "  " + statusStr
	} else if healthStr != "" {
		metaStr = healthStr
	} else {
		metaStr = statusStr
	}

	var prefix string
	if selected {
		prefix = lipgloss.NewStyle().Foreground(ui.ColorCyan).Bold(true).Render("▶ ")
	} else if selectedRange {
		prefix = lipgloss.NewStyle().Foreground(ui.ColorCyan).Bold(true).Render("┃ ")
	} else {
		prefix = "  "
	}

	prefixWidth := lipgloss.Width(prefix)
	metaWidth := lipgloss.Width(metaStr)

	availForRepoAndBranch := maxWidth - prefixWidth - metaWidth - 1
	if availForRepoAndBranch < 5 {
		availForRepoAndBranch = 5
	}

	repoName := r.Name
	branchName := r.Branch

	var repoStr, branchStr string
	repoWidth := lipgloss.Width(repoName)

	if repoWidth >= availForRepoAndBranch {
		if repoWidth > availForRepoAndBranch {
			repoName = truncateRunes(repoName, availForRepoAndBranch)
		}
		if isSelected {
			repoStr = lipgloss.NewStyle().Foreground(ui.ColorFg).Bold(true).Render(repoName)
		} else {
			repoStr = lipgloss.NewStyle().Foreground(ui.ColorFg).Render(repoName)
		}
	} else {
		if isSelected {
			repoStr = lipgloss.NewStyle().Foreground(ui.ColorFg).Bold(true).Render(repoName)
		} else {
			repoStr = lipgloss.NewStyle().Foreground(ui.ColorFg).Render(repoName)
		}

		if branchName != "" {
			availForBranch := availForRepoAndBranch - repoWidth - 1
			if availForBranch >= 3 {
				if lipgloss.Width(branchName) > availForBranch {
					branchName = truncateRunes(branchName, availForBranch)
				}
				if branchName != "" {
					branchStr = ui.BranchStyle.Render(branchName)
				}
			}
		}
	}

	leftContent := prefix + repoStr
	if branchStr != "" {
		leftContent += " " + branchStr
	}

	leftWidth := lipgloss.Width(leftContent)
	gapLen := maxWidth - leftWidth - metaWidth
	if gapLen < 1 {
		gapLen = 1
	}

	gap := strings.Repeat(" ", gapLen)

	line := leftContent
	if metaStr != "" {
		line += gap + metaStr
	}
	if isSelected {
		return renderActiveRow(line, maxWidth)
	}
	return line
}

func renderActiveRow(line string, width int) string {
	_ = width
	return ui.SelectedItemStyle.Render(line)
}

func (m *Model) repoHealthBadges(r domain.Repository, isSelected bool) []string {
	indicatorStyle := lipgloss.NewStyle().Bold(true)

	fgWarning := ui.ColorWarning
	fgAmber := ui.ColorAmber
	fgOrange := ui.ColorOrange
	fgCyan := ui.ColorCyan

	var badges []string
	if r.IsDetached {
		badges = append(badges, indicatorStyle.Foreground(fgWarning).Render("DET"))
	}
	if !r.IsDetached && !r.HasUpstream && r.Branch != "" {
		badges = append(badges, indicatorStyle.Foreground(fgAmber).Render("UP"))
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
	} else if m.activePanel == RebasePanel {
		panelNum = m.getPanelNumber(RebasePanel)
		panelLabel = "Interactive Rebase · " + r.Name
	} else if m.showConflicts {
		panelNum = m.getPanelNumber(ConflictPanel)
		panelLabel = "Conflicts · " + r.Name
	} else {
		panelNum = m.getPanelNumber(LogPanel)

		var label string
		if m.showFiles {
			label = "Files · " + r.Name
		} else if m.showBranches {
			label = "Branches · " + r.Name
		} else if m.showStashes {
			label = "Stashes · " + r.Name
		} else {
			label = "Repository · " + r.Name
		}
		panelLabel = label
	}

	var content string
	if m.activePanel == CommandLogPanel {
		content = renderViewportWithScrollbar(m.logViewport, m.activePanel == CommandLogPanel)
	} else if m.activePanel == ConfigPanel {
		content = m.renderConfigPanel(width)
	} else if m.activePanel == RebasePanel {
		content = m.renderRebasePanel(width)
	} else if m.showConflicts {
		content = m.renderConflictList(width)
	} else if m.showFiles {
		content = m.renderFilesWorkspace(width)
	} else if m.showBranches {
		content = lipgloss.JoinVertical(lipgloss.Left,
			m.renderBranchesList(width),
			"",
			m.renderBranchPreview(width),
		)
	} else if m.showStashes {
		content = m.renderStashList(width)
	} else {
		content = renderViewportWithScrollbar(m.viewport, m.activePanel == LogPanel)
	}

	if m.tagAssignModal {
		content = m.renderRepoTagsSection(width)
	}

	content = clipRenderedContent(content, height-2)

	active := m.activePanel == LogPanel || m.activePanel == DiffPanel || m.activePanel == CommandLogPanel || m.activePanel == ConflictPanel || m.tagAssignModal
	accent := lipgloss.Color(ui.ColorCyan)
	return m.renderTitledPanel(width, height, "["+panelNum+"] "+panelLabel, content, active, accent)
}

func (m *Model) renderFilesWorkspace(width int) string {
	listContent := renderViewportWithScrollbar(m.fileViewport, m.activePanel == LogPanel)
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
	var diffContent string
	if m.compactDiff {
		diffContent = m.renderCompactDiffContent()
	} else if m.diffFetching {
		diffContent = ui.SpinnerStyle.Render("  " + m.spinnerView() + " Loading diff…")
	} else if m.currentDiff == "" {
		diffContent = ui.SubtleStyle.Render("  No diff available")
	} else {
		diffContent = renderViewportWithScrollbar(m.diffViewport, m.activePanel == DiffPanel)
	}

	if !m.usesSideBySideDiff() {
		diffHeader := diffTitleStyle.Width(width - 2).Render("[" + m.getPanelNumber(DiffPanel) + "] Diff" + diffFileName)
		return lipgloss.JoinVertical(lipgloss.Left, listContent, diffHeader, diffContent)
	}

	filePaneWidth := m.fileViewport.Width + 1
	diffPaneWidth := m.diffViewport.Width + 1
	filesHeader := ui.DiffTabStyle(m.activePanel == LogPanel).Width(filePaneWidth).Render("[" + m.getPanelNumber(LogPanel) + "] Files (" + fmt.Sprint(len(m.files)) + ")")
	diffHeader := diffTitleStyle.Width(diffPaneWidth).Render("[" + m.getPanelNumber(DiffPanel) + "] Diff" + diffFileName)
	filesPane := lipgloss.NewStyle().Width(filePaneWidth).Render(lipgloss.JoinVertical(lipgloss.Left, filesHeader, listContent))
	diffPane := lipgloss.NewStyle().Width(diffPaneWidth).Render(lipgloss.JoinVertical(lipgloss.Left, diffHeader, diffContent))
	return lipgloss.JoinHorizontal(lipgloss.Top, filesPane, " ", diffPane)
}

func clipRenderedContent(content string, maxLines int) string {
	if maxLines <= 0 {
		return ""
	}
	// Count newlines without allocating a slice; return a prefix of the string.
	count := 0
	for i := 0; i < len(content); i++ {
		if content[i] == '\n' {
			count++
			if count >= maxLines {
				return content[:i]
			}
		}
	}
	return content
}

func compactAge(ageStr string) string {
	s := strings.TrimSpace(ageStr)
	replacements := []struct {
		old string
		new string
	}{
		{" years ago", "y ago"}, {" year ago", "y ago"},
		{" months ago", "mo ago"}, {" month ago", "mo ago"},
		{" weeks ago", "w ago"}, {" week ago", "w ago"},
		{" days ago", "d ago"}, {" day ago", "d ago"},
		{" hours ago", "h ago"}, {" hour ago", "h ago"},
		{" minutes ago", "m ago"}, {" minute ago", "m ago"},
		{" seconds ago", "s ago"}, {" second ago", "s ago"},
	}
	for _, r := range replacements {
		if strings.HasSuffix(s, r.old) {
			return strings.TrimSuffix(s, r.old) + r.new
		}
	}
	return s
}

func compactRelativeDuration(d time.Duration) string {
	if d < time.Minute {
		return "just now"
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	if d < 7*24*time.Hour {
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	}
	if d < 30*24*time.Hour {
		return fmt.Sprintf("%dw", int(d.Hours()/(24*7)))
	}
	if d < 365*24*time.Hour {
		return fmt.Sprintf("%dmo", int(d.Hours()/(24*30)))
	}
	return fmt.Sprintf("%dy", int(d.Hours()/(24*365)))
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
		if len(parts) < 5 {
			lines = append(lines, "  "+ui.ValueStyle.Render(rawLine))
			continue
		}

		graphAndHash := parts[0]
		var hash, graphPart string
		lastSpace := strings.LastIndex(graphAndHash, " ")
		if lastSpace != -1 {
			graphPart = graphAndHash[:lastSpace+1]
			hash = graphAndHash[lastSpace+1:]
		} else {
			hash = graphAndHash
		}

		beautifiedGraph1 := ""
		beautifiedGraph2 := ""
		for i, char := range graphPart {
			style := ui.GraphCharStyles[i%len(ui.GraphCharStyles)]
			switch char {
			case '*':
				beautifiedGraph1 += style.Copy().Bold(true).Render("●")
				beautifiedGraph2 += style.Render("│")
			case '|':
				beautifiedGraph1 += style.Render("│")
				beautifiedGraph2 += style.Render("│")
			case '/':
				beautifiedGraph1 += style.Render("╯")
				beautifiedGraph2 += " "
			case '\\':
				beautifiedGraph1 += style.Render("╰")
				beautifiedGraph2 += " "
			case '_':
				beautifiedGraph1 += style.Render("─")
				beautifiedGraph2 += " "
			case ' ':
				beautifiedGraph1 += " "
				beautifiedGraph2 += " "
			default:
				beautifiedGraph1 += style.Render(string(char))
				beautifiedGraph2 += " "
			}
		}

		subject := parts[2]
		prNum := ""
		if idx := strings.LastIndex(subject, " (#"); idx != -1 && strings.HasSuffix(subject, ")") {
			prNum = subject[idx+2 : len(subject)-1]
			subject = strings.TrimSpace(subject[:idx])
		} else if idx := strings.LastIndex(subject, " #"); idx != -1 {
			prNum = subject[idx+1:]
			subject = strings.TrimSpace(subject[:idx])
		}

		prefix1 := fmt.Sprintf("  %s %s  ", beautifiedGraph1, ui.SubtleStyle.Render(hash))
		prefix1Width := lipgloss.Width(prefix1)

		availSubjectWidth := width - prefix1Width - 1
		if availSubjectWidth < 10 {
			availSubjectWidth = 10
		}

		subjStr := truncateRunes(subject, availSubjectWidth)
		line1 := prefix1 + ui.ValueStyle.Render(subjStr)
		lines = append(lines, line1)

		var metaParts []string
		if prNum != "" {
			metaParts = append(metaParts, prNum)
		}
		if parts[3] != "" {
			metaParts = append(metaParts, compactAge(parts[3]))
		}
		if parts[4] != "" {
			metaParts = append(metaParts, parts[4])
		}
		if parts[1] != "" {
			metaParts = append(metaParts, ui.BranchStyle.Render(parts[1]))
		}

		metaStr := strings.Join(metaParts, " · ")
		indent2 := strings.Repeat(" ", lipgloss.Width(hash)+3)
		prefix2 := fmt.Sprintf("  %s%s", beautifiedGraph2, indent2)
		line2 := prefix2 + ui.SubtleStyle.Render(metaStr)
		lines = append(lines, line2)
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
	sections := make([]string, 0, 20)
	cache := m.repoDetailCacheFor(r.Path)

	dividerWidth := width - 2
	if dividerWidth < 10 {
		dividerWidth = 10
	}
	divider := ui.SubtleStyle.Render("─" + strings.Repeat("─", dividerWidth))

	branchVal := ui.BranchStyle.Render(r.Branch)
	if r.Branch == "" {
		branchVal = ui.SubtleStyle.Render("HEAD")
	}

	var workingTreeVal string
	if r.HasConflicts {
		workingTreeVal = ui.ErrorStyle.Render("Conflict")
	} else if r.IsDirty {
		dirtyCount := r.DirtyCount()
		if cache != nil && cache.modifiedCount+cache.untrackedCount > 0 {
			dirtyCount = cache.modifiedCount + cache.untrackedCount
		}
		if dirtyCount > 0 {
			workingTreeVal = ui.DirtyStyle.Render(fmt.Sprintf("%d modified", dirtyCount))
		} else {
			workingTreeVal = ui.DirtyStyle.Render("Dirty")
		}
	} else {
		workingTreeVal = ui.CleanStyle.Render("Clean")
	}
	var remoteVal string
	if r.Ahead == 0 && r.Behind == 0 {
		remoteVal = ui.CleanStyle.Render("Synced")
	} else if r.Ahead > 0 && r.Behind == 0 {
		remoteVal = ui.AheadStyle.Render(fmt.Sprintf("%d ahead", r.Ahead))
	} else if r.Ahead == 0 && r.Behind > 0 {
		remoteVal = ui.BehindStyle.Render(fmt.Sprintf("%d behind", r.Behind))
	} else {
		remoteVal = fmt.Sprintf("%s · %s",
			ui.AheadStyle.Render(fmt.Sprintf("%d ahead", r.Ahead)),
			ui.BehindStyle.Render(fmt.Sprintf("%d behind", r.Behind)),
		)
	}
	modifiedCount := r.ModifiedCount
	untrackedCount := r.UntrackedCount
	if cache != nil {
		modifiedCount = cache.modifiedCount
		untrackedCount = cache.untrackedCount
	}
	lastFetchStr := "Never"
	if !r.LastFetch.IsZero() {
		lastFetchStr = compactRelativeDuration(time.Since(r.LastFetch)) + " ago"
	}
	metrics := []string{
		renderOverviewMetric("Branch", branchVal),
		renderOverviewMetric("Tree", workingTreeVal),
		renderOverviewMetric("Remote", remoteVal),
		renderOverviewMetric("Changes", ui.ValueStyle.Render(fmt.Sprintf("%d modified · %d untracked", modifiedCount, untrackedCount))),
	}
	if width >= 64 {
		sections = append(sections,
			"  "+metrics[0]+"    "+metrics[1],
			"  "+metrics[2]+"    "+metrics[3],
		)
	} else {
		for _, metric := range metrics {
			sections = append(sections, "  "+metric)
		}
	}
	sections = append(sections, "  "+renderOverviewMetric("Fetched", ui.ValueStyle.Render(lastFetchStr)))

	healthLabels := m.repoHealthLabels(r)
	if len(healthLabels) > 0 {
		sections = append(sections, "  "+renderOverviewMetric("Health", ui.WarningStyle.Render(strings.Join(healthLabels, " | "))))
	}

	if r.Error != "" {
		sections = append(sections, "", ui.ErrorStyle.Render("  ✗ Error: "+r.Error))
	}

	if len(r.Tags) > 0 {
		sections = append(sections, "")
		var tagBadges []string
		for _, tag := range r.Tags {
			tagBadges = append(tagBadges, m.renderTagBadge(tag))
		}
		sections = append(sections,
			"  "+renderOverviewMetric("Tags", strings.Join(tagBadges, " ")),
		)
	}

	title := "Recent Activity"
	if !m.viewGraph {
		title = "Recent Activity"
	}
	sections = append(sections, "", ui.PanelTitleStyle.Render(title))
	sections = append(sections, divider)

	if m.detailLoading && m.cachedDetailFor == r.Path {
		sections = append(sections, ui.SpinnerStyle.Render("  "+m.spinnerView()+" Loading repository details…"))
	}

	log := m.cachedLog
	showLog := m.cachedLogFor == r.Path && m.cachedLog != ""
	if !showLog && cache != nil && cache.log != "" {
		log = cache.log
		showLog = true
	}

	if showLog {
		sections = append(sections, m.renderBeautifiedLog(log, width))
	} else if m.detailLoading {
		sections = append(sections, ui.SpinnerStyle.Render("  "+m.spinnerView()+" Loading commits…"))
	} else {
		sections = append(sections, ui.SubtleStyle.Render("  No commits yet"))
	}

	if r.LastOutput != "" {
		sections = append(sections, "", ui.PanelTitleStyle.Render("Last Output"))
		sections = append(sections, divider)
		for _, line := range strings.Split(r.LastOutput, "\n") {
			for _, wrapped := range wrapPlainText(line, width-4) {
				sections = append(sections, "  "+ui.ValueStyle.Render(wrapped))
			}
		}
	}

	return strings.Join(sections, "\n")
}

func renderOverviewMetric(label, value string) string {
	return ui.LabelStyle.Render(label+":") + " " + value
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
		lines = append(lines, m.renderRepoLine(idx, r, width-4))
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
	isSelected := selected || selectedRange

	var prefix string
	if selected {
		prefix = lipgloss.NewStyle().Foreground(ui.ColorCyan).Bold(true).Render("▶ ")
	} else if selectedRange {
		prefix = lipgloss.NewStyle().Foreground(ui.ColorCyan).Bold(true).Render("┃ ")
	} else {
		prefix = "  "
	}

	selectedInList := m.fileSelections[index]
	checkboxStr := "○"
	cbStyle := ui.SubtleStyle
	if selectedInList {
		checkboxStr = "●"
		cbStyle = lipgloss.NewStyle().Foreground(ui.ColorSuccess).Bold(true)
	}
	checkbox := cbStyle.Render(checkboxStr)

	statusIcon := " "
	statusStyle := ui.SubtleStyle
	if f.Untracked {
		statusIcon = "?"
		statusStyle = lipgloss.NewStyle().Foreground(ui.ColorError).Bold(true)
	} else if f.Modified {
		statusIcon = "M"
		statusStyle = lipgloss.NewStyle().Foreground(ui.ColorError).Bold(true)
	} else if f.Staged {
		statusIcon = "A"
		statusStyle = lipgloss.NewStyle().Foreground(ui.ColorSuccess).Bold(true)
	}
	statusInd := statusStyle.Render(statusIcon)

	name := f.Name
	if len(name) > maxNameWidth {
		name = "…" + name[len(name)-maxNameWidth+1:]
	}

	nameStyle := lipgloss.NewStyle().Foreground(ui.ColorFg)
	if isSelected {
		nameStyle = nameStyle.Bold(true)
	}
	nameStr := nameStyle.Render(name)

	line := prefix + checkbox + " " + statusInd + "  " + nameStr
	if isSelected {
		return renderActiveRow(line, width-4)
	}
	return line
}

func (m *Model) renderBranchesList(width int) string {
	if len(m.branches) == 0 {
		return ui.SubtleStyle.Render("  No branches found")
	}

	lines := make([]string, 0, len(m.branches))
	for i, b := range m.branches {
		selected := i == m.branchCursor
		selectedRange := m.lineSelected(LogPanel, i) && m.showBranches
		isSelected := selected || selectedRange

		var prefix string
		if selected {
			prefix = lipgloss.NewStyle().Foreground(ui.ColorCyan).Bold(true).Render("▶ ")
		} else if selectedRange {
			prefix = lipgloss.NewStyle().Foreground(ui.ColorCyan).Bold(true).Render("┃ ")
		} else if b.IsCurrent {
			prefix = ui.CleanStyle.Render("✓ ")
		} else {
			prefix = "  "
		}

		nameStyle := lipgloss.NewStyle().Foreground(ui.ColorFg)
		if isSelected {
			nameStyle = nameStyle.Bold(true)
		} else if b.IsWorktree {
			nameStyle = lipgloss.NewStyle().Foreground(ui.ColorCyan)
		}
		nameStr := nameStyle.Render(b.Name)

		if isSelected && b.IsCurrent {
			nameStr = ui.CleanStyle.Render("✓ ") + nameStr
		}

		indicators := []string{}
		if b.IsLocal {
			indicators = append(indicators, "local")
		}
		if b.IsRemote {
			indicators = append(indicators, "remote")
		}
		if b.IsWorktree {
			indicators = append(indicators, "worktree")
		}

		scopeText := strings.Join(indicators, " · ")
		scopeStr := ui.SubtleStyle.Render(scopeText)

		wtBadge := ""
		if b.IsWorktree {
			wtBadge = " " + lipgloss.NewStyle().Background(ui.ColorCyan).Foreground(ui.ColorBg).Bold(true).Render("WT")
		}

		leftPart := prefix + nameStr + wtBadge
		leftLen := lipgloss.Width(leftPart)
		scopeLen := lipgloss.Width(scopeStr)

		padLen := (width - 4) - leftLen - scopeLen
		if padLen < 1 {
			padLen = 1
		}
		padSpaces := strings.Repeat(" ", padLen)

		line := leftPart + padSpaces + scopeStr
		if isSelected {
			line = renderActiveRow(line, width-4)
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func (m *Model) renderBranchPreview(width int) string {
	if m.branchCursor < 0 || m.branchCursor >= len(m.branches) {
		return ""
	}

	b := m.branches[m.branchCursor]
	scope := make([]string, 0, 3)
	if b.IsLocal {
		scope = append(scope, "local")
	}
	if b.IsRemote {
		scope = append(scope, "remote")
	}
	if b.IsWorktree {
		scope = append(scope, "worktree")
	}
	if len(scope) == 0 {
		scope = append(scope, "unknown")
	}

	dividerWidth := width - 6
	if dividerWidth < 12 {
		dividerWidth = 12
	}
	action := "enter checkout"
	if b.IsWorktree {
		action = "enter open worktree"
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		ui.PanelTitleStyle.Render("Selected branch"),
		ui.SubtleStyle.Render(strings.Repeat("─", dividerWidth)),
		"  "+renderOverviewMetric("Name", ui.BranchStyle.Render(truncateRunes(b.Name, width-14))),
		"  "+renderOverviewMetric("Scope", ui.ValueStyle.Render(strings.Join(scope, " · "))),
		ui.SubtleStyle.Render("  "+action+"  ·  M merge  ·  d delete"),
	)
}

func (m *Model) renderStashList(width int) string {
	if len(m.stashes) == 0 {
		return ui.SubtleStyle.Render("  No stashes found")
	}

	lines := make([]string, 0, len(m.stashes))
	for i, s := range m.stashes {
		selected := i == m.stashCursor
		selectedRange := m.lineSelected(LogPanel, i) && m.showStashes
		isSelected := selected || selectedRange

		var prefix string
		if isSelected {
			prefix = lipgloss.NewStyle().Foreground(ui.ColorCyan).Bold(true).Render("▶ ")
		} else {
			prefix = "  "
		}

		indexStyle := lipgloss.NewStyle().Foreground(ui.ColorHighlight)
		if isSelected {
			indexStyle = indexStyle.Bold(true)
		}
		indexStr := indexStyle.Render(fmt.Sprintf("stash@{%d}", s.Index))

		msgStyle := lipgloss.NewStyle().Foreground(ui.ColorFg)
		if isSelected {
			msgStyle = msgStyle.Bold(true)
		}
		msgStr := msgStyle.Render(" " + s.Message)

		line := prefix + indexStr + msgStr
		if isSelected {
			line = renderActiveRow(line, width-4)
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
					lines = append(lines, lipgloss.NewStyle().Foreground(ui.ColorCyan).Bold(true).Render("   ▶ ")+f)
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
		isSelected := selected || selectedRange

		var prefix string
		if isSelected {
			prefix = lipgloss.NewStyle().Foreground(ui.ColorCyan).Bold(true).Render("▶ ")
		} else {
			prefix = "  "
		}

		statusStyle := lipgloss.NewStyle().Foreground(ui.ColorError).Bold(true)
		statusStr := statusStyle.Render("⬌ " + c.Status + " ")

		nameStyle := lipgloss.NewStyle().Foreground(ui.ColorFg)
		if isSelected {
			nameStyle = nameStyle.Bold(true)
		}
		nameStr := nameStyle.Render(c.Name)

		line := prefix + statusStr + nameStr
		if isSelected {
			line = renderActiveRow(line, width-4)
		}
		lines = append(lines, line)
	}

	lines = append(lines, "", ui.SubtleStyle.Render("  Press enter to resolve the selected file with the mergetool"))

	return strings.Join(lines, "\n")
}

func (m *Model) renderCompactDiffContent() string {
	if m.compactFetching {
		return ui.SpinnerStyle.Render("  " + m.spinnerView() + " Loading compact diff…")
	}

	if len(m.compactChanges) == 0 {
		return ui.SubtleStyle.Render("  No changes detected")
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

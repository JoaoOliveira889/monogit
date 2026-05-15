package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"monogit/internal/domain"
	"monogit/internal/pkg/ui"
)

func (m *Model) renderBody() string {
	leftWidth := m.leftPanelWidth()
	rightWidth := m.rightPanelWidth()

	headerHeight := 1
	if m.statusMsg != "" {
		headerHeight = 2
	}
	footerHeight := 1
	bodyHeight := m.height - headerHeight - footerHeight
	if bodyHeight < 5 {
		bodyHeight = 5
	}

	left := m.renderRepoList(leftWidth, bodyHeight)
	right := m.renderDetailPanel(rightWidth, bodyHeight)

	return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
}

func (m *Model) renderTitledPanel(width, height int, title string, content string, active bool) string {
	style := ui.LeftPanelStyle
	if active {
		style = ui.ActivePanelStyle
	}

	style = style.BorderTop(false)

	border := style.GetBorderStyle()
	titleText := fmt.Sprintf("─[%s]─", title)
	titleWidth := lipgloss.Width(titleText)

	repeatCount := width - titleWidth - 2
	if repeatCount < 0 {
		repeatCount = 0
	}
	topLine := border.TopLeft + titleText + strings.Repeat(border.Top, repeatCount) + border.TopRight

	var styledTopLine string
	if active {
		styledTopLine = lipgloss.NewStyle().Foreground(ui.ColorHighlight).Render(topLine)
	} else {
		styledTopLine = lipgloss.NewStyle().Foreground(ui.ColorBorder).Render(topLine)
	}

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
	content := m.repoViewport.View()
	return m.renderTitledPanel(width, height, m.getPanelNumber(RepoPanel), content, m.activePanel == RepoPanel)
}

func (m *Model) renderRepoLine(index int, r domain.Repository, maxWidth int) string {
	selected := index == m.cursor
	bg := ui.ColorBg
	if selected {
		bg = ui.ColorHighlight
	}

	name := r.Name
	if len(name) > 15 {
		name = name[:14] + "…"
	}

	nameStyle := lipgloss.NewStyle().Foreground(ui.ColorFg)
	if selected {
		nameStyle = nameStyle.Background(bg).Foreground(ui.ColorBg).Bold(true)
	}
	nameStr := nameStyle.Width(15).Render(name)

	var indicators []string
	indicatorStyle := lipgloss.NewStyle().Background(bg).Bold(true)

	indicators = append(indicators, indicatorStyle.Foreground(ui.ColorSuccess).Render(fmt.Sprintf("%s%d", ui.IconAhead, r.Ahead)))
	indicators = append(indicators, indicatorStyle.Foreground(ui.ColorWarning).Render(fmt.Sprintf("%s%d", ui.IconBehind, r.Behind)))

	if r.IsDirty {
		indicators = append(indicators, indicatorStyle.Foreground(ui.ColorError).Render(ui.IconDirty))
	} else if r.Branch != "" {
		indicators = append(indicators, indicatorStyle.Foreground(ui.ColorSuccess).Render(ui.IconClean))
	}

	space := lipgloss.NewStyle().Background(bg).Render(" ")
	indicatorStr := strings.Join(indicators, space)

	prefix := lipgloss.NewStyle().Background(bg).Render("  ")
	midSpace := lipgloss.NewStyle().Background(bg).Render(" ")

	line := prefix + nameStr + midSpace + indicatorStr

	padLen := maxWidth - lipgloss.Width(line)
	if padLen > 0 {
		line += lipgloss.NewStyle().Background(bg).Render(strings.Repeat(" ", padLen))
	}

	return line
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
	} else {
		panelNum = m.getPanelNumber(LogPanel)
		if m.showFiles {
			panelLabel = "File Selection"
		} else if m.showBranches {
			panelLabel = "Branches"
		} else {
			panelLabel = r.Name
		}
	}

	var content string
	if m.activePanel == CommandLogPanel {
		content = m.logViewport.View()
	} else if m.showFiles {
		listContent := m.fileViewport.View()

		var diffTitleStyle lipgloss.Style
		if m.activePanel == DiffPanel {
			diffTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#7aa2f7")).
				Bold(true).
				Border(lipgloss.NormalBorder(), true, false, false, false).
				BorderForeground(lipgloss.Color("#7aa2f7")).
				PaddingLeft(1)
		} else {
			diffTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#565f89")).
				Border(lipgloss.NormalBorder(), true, false, false, false).
				BorderForeground(lipgloss.Color("#414868")).
				PaddingLeft(1)
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
		if m.diffFetching {
			diffContent = ui.SpinnerStyle.Render("   " + m.spinnerView() + " Loading diff...")
		} else if m.currentDiff == "" {
			diffContent = ui.SubtleStyle.Render("   No diff available")
		} else {
			diffContent = m.diffViewport.View()
		}

		content = lipgloss.JoinVertical(lipgloss.Left,
			listContent,
			diffHeader,
			diffContent,
		)
	} else if m.showBranches {
		content = m.renderBranchesList(width)
	} else {
		content = m.viewport.View()
	}

	return m.renderTitledPanel(width, height, panelNum+"-"+panelLabel, content, m.activePanel == LogPanel || m.activePanel == DiffPanel || m.activePanel == CommandLogPanel)
}

func (m *Model) renderBeautifiedLog(log string) string {
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
			style := lipgloss.NewStyle().Foreground(ui.GraphColors[i%len(ui.GraphColors)])
			switch char {
			case '*':
				beautifiedGraph += style.Bold(true).Render("●")
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
			line := fmt.Sprintf("  %s %s %s %s",
				beautifiedGraph,
				ui.SubtleStyle.Render(hash),
				ui.ValueStyle.Render(parts[2]),
				ui.SubtleStyle.Render(parts[3]),
			)
			lines = append(lines, line)
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

func (m *Model) renderViewportContent() string {
	r := m.selectedRepo()
	if r == nil {
		return ui.SubtleStyle.Render("No repository selected")
	}

	sections := make([]string, 0, 16)

	statusStr := ui.CleanStyle.Render("Clean " + ui.IconClean)
	if r.IsDirty {
		statusStr = ui.DirtyStyle.Render("Modified " + ui.IconDirty)
	}

	sections = append(sections,
		ui.LabelStyle.Render("   Branch:  ")+ui.BranchStyle.Render(r.Branch),
		ui.LabelStyle.Render("   Status:  ")+statusStr,
	)

	if r.Tagging {
		sections = append(sections, ui.ValueStyle.Render("   "+m.spinnerView()+" Tagging & Deploying..."))
	}

	sections = append(sections,
		fmt.Sprintf("   %s %d modified, %d untracked", ui.IconSpace, m.cachedModifiedCount, m.cachedUntrackedCount),
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

	if m.cachedLastCommit != "" && m.cachedLastCommit != "(no commits yet)" {
		sections = append(sections, "", ui.LabelStyle.Render("   Last Commit:"))
		parts := strings.Split(m.cachedLastCommit, " ")
		if len(parts) > 1 {
			hash := ui.SubtleStyle.Render(parts[0])
			msg := ui.ValueStyle.Render(strings.Join(parts[1:], " "))
			sections = append(sections, fmt.Sprintf("     %s %s", hash, msg))
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

	var log string
	if m.viewGraph {
		log, _ = m.gitUC.GetGraphLog(r.Path, 30)
	} else {
		log, _ = m.gitUC.GetSimpleLog(r.Path, 30)
	}

	renderedLog := m.renderBeautifiedLog(log)
	sections = append(sections, renderedLog)

	if r.LastOutput != "" {
		sections = append(sections, "", ui.LabelStyle.Render("   Last Output:"))
		sections = append(sections, ui.SubtleStyle.Render("   "+strings.Repeat("─", 40)))
		for _, line := range strings.Split(r.LastOutput, "\n") {
			sections = append(sections, "   "+ui.ValueStyle.Render(line))
		}
	}

	return strings.Join(sections, "\n")
}

func (m *Model) renderRepoViewportContent() string {
	width := m.leftPanelWidth()
	if len(m.repos) == 0 {
		return ui.SubtleStyle.Render("  No repositories found")
	}

	lines := make([]string, 0, len(m.repos))
	for i, r := range m.repos {
		lines = append(lines, m.renderRepoLine(i, r, width-2))
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
	bg := ui.ColorBg
	if selected {
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
	if selected {
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
		bg := ui.ColorBg
		if selected {
			bg = ui.ColorHighlight
		}

		bgStyle := lipgloss.NewStyle().Background(bg)

		prefix := "   "
		if b.IsCurrent {
			prefix = " ✓ "
		}
		if selected {
			prefix = bgStyle.Render(prefix)
		}

		nameStyle := lipgloss.NewStyle().Foreground(ui.ColorFg)
		if selected {
			nameStyle = lipgloss.NewStyle().Background(ui.ColorHighlight).Foreground(ui.ColorBg).Bold(true)
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
		if selected {
			indicator = lipgloss.NewStyle().Background(bg).Foreground(ui.ColorBg).Render(indicator)
		}

		line := prefix + nameStr + indicator

		padLen := (width - 2) - lipgloss.Width(line)
		if padLen > 0 {
			padSpaces := strings.Repeat(" ", padLen)
			if selected {
				padSpaces = bgStyle.Render(padSpaces)
			}
			line += padSpaces
		}
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
		
		status := ui.CleanStyle.Render("SUCCESS")
		if entry.Error != nil {
			status = ui.ErrorStyle.Render("FAILED")
		}

		fmt.Fprintf(&sb, "  [%s] %s > %s : %s\n", timeStr, repoStr, cmdStr, status)
		
		if entry.Error != nil {
			errText := "Error: " + entry.Error.Error()
			wrappedErr := lipgloss.NewStyle().
				Foreground(ui.ColorError).
				Width(contentWidth).
				Render(errText)
			
			for _, line := range strings.Split(wrappedErr, "\n") {
				fmt.Fprintf(&sb, "    %s\n", line)
			}
		}
		
		if entry.Output != "" {
			wrappedOutput := lipgloss.NewStyle().
				Foreground(ui.ColorSubtle).
				Width(contentWidth).
				Render(strings.TrimSpace(entry.Output))
			
			for _, line := range strings.Split(wrappedOutput, "\n") {
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
		if m.cursor < m.repoViewport.YOffset {
			m.repoViewport.YOffset = m.cursor
		} else if m.cursor >= m.repoViewport.YOffset+m.repoViewport.Height {
			m.repoViewport.YOffset = m.cursor - m.repoViewport.Height + 1
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

package tui

import (
	tea "charm.land/bubbletea/v2"
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/JoaoOliveira889/monogit/internal/domain"
	"github.com/JoaoOliveira889/monogit/internal/pkg/ui"
)

const (
	// diffFileListWidth is the fixed width of the file column. The diff gets
	// everything else, since it is the part worth reading.
	diffFileListWidth    = 34
	diffMinFileListWidth = 18

	// diffGutterWidth is the width of one line-number column.
	diffGutterWidth = 4
)

// renderDiffViewer draws the full-screen diff: files on the left, the selected
// file's diff on the right with old and new line numbers.
func (m *Model) renderDiffViewer(width, height int) string {
	listWidth := diffFileListWidth
	if width < 100 {
		listWidth = diffMinFileListWidth
	}
	if listWidth > width/2 {
		listWidth = width / 2
	}
	// The separator column owns one cell plus a space on each side.
	const separatorWidth = 3

	diffWidth := width - listWidth - separatorWidth
	if diffWidth < 20 {
		diffWidth = 20
	}

	files := m.renderDiffFileList(listWidth, height)
	diff := m.renderDiffPane(diffWidth, height)

	// Exactly height rows, with no trailing newline: one extra row here would
	// push the footer off the frame.
	rule := ui.SubtleStyle.Render("│")
	rows := make([]string, height)
	for i := range rows {
		rows[i] = " " + rule + " "
	}
	separator := strings.Join(rows, "\n")

	return lipgloss.JoinHorizontal(lipgloss.Top, files, separator, diff)
}

// diffFileBadge returns the single-letter status marker and its colour.
func diffFileBadge(f domain.FileStatus) (string, lipgloss.Style) {
	switch {
	case f.Untracked:
		return "A", ui.DiffAddStyle
	case f.Staged && !f.Modified:
		return "S", ui.SuccessStyle
	case f.Modified:
		return "M", ui.WarningStyle
	default:
		return "M", ui.WarningStyle
	}
}

func (m *Model) renderDiffFileList(width, height int) string {
	title := ui.PanelTitleStyle.Render(fmt.Sprintf("Files (%d)", len(m.files)))
	rows := []string{title, ui.SubtleStyle.Render(strings.Repeat("─", max(width, 1)))}

	if len(m.files) == 0 {
		rows = append(rows, ui.SubtleStyle.Render(" working tree clean"))
	}

	// Keep the cursor inside the visible window.
	visible := height - len(rows)
	first := 0
	if visible > 0 && m.fileCursor >= visible {
		first = m.fileCursor - visible + 1
	}

	for i := first; i < len(m.files) && len(rows) < height; i++ {
		f := m.files[i]
		badge, badgeStyle := diffFileBadge(f)

		marker := "  "
		if i == m.fileCursor {
			marker = ui.CursorMarkerStyle.Render("▌ ")
		}
		selected := ""
		if m.fileSelections[i] {
			selected = ui.SuccessStyle.Render("✓")
		} else {
			selected = " "
		}

		nameWidth := width - lipgloss.Width(marker) - 4
		name := truncatePathLeft(f.Name, nameWidth)

		nameStyle := ui.ValueStyle
		if i == m.fileCursor {
			nameStyle = nameStyle.Bold(true)
		}

		rows = append(rows, marker+badgeStyle.Render(badge)+selected+" "+nameStyle.Render(name))
	}

	return lipgloss.NewStyle().Width(width).Render(
		clipToBox(strings.Join(rows, "\n"), width, height))
}

// truncatePathLeft shortens a path from the left, which keeps the file name —
// the part that identifies it — rather than the directory prefix.
func truncatePathLeft(path string, width int) string {
	if width <= 1 {
		return ""
	}
	if ansi.StringWidth(path) <= width {
		return path
	}
	runes := []rune(path)
	keep := width - 1
	if keep >= len(runes) {
		return path
	}
	return "…" + string(runes[len(runes)-keep:])
}

func (m *Model) renderDiffPane(width, height int) string {
	header := m.renderDiffHeader(width)
	headerHeight := lipgloss.Height(header)

	bodyHeight := height - headerHeight
	if bodyHeight < 1 {
		bodyHeight = 1
	}

	m.diffViewport.SetWidth(width)
	m.diffViewport.SetHeight(bodyHeight)

	body := renderViewportWithScrollbar(m.diffViewport, true)
	return lipgloss.JoinVertical(lipgloss.Left, header, clipToBox(body, width, bodyHeight))
}

func (m *Model) renderDiffHeader(width int) string {
	name := "No file selected"
	if m.fileCursor < len(m.files) {
		name = m.files[m.fileCursor].Name
	}

	stats := ""
	if m.parsedDiff.Added > 0 || m.parsedDiff.Removed > 0 {
		stats = ui.DiffAddStyle.Render(fmt.Sprintf("+%d", m.parsedDiff.Added)) + " " +
			ui.DiffDelStyle.Render(fmt.Sprintf("−%d", m.parsedDiff.Removed))
	}
	if m.diffFetching {
		stats = ui.SpinnerStyle.Render(m.spinnerView() + " loading")
	}

	statsWidth := lipgloss.Width(stats)
	nameWidth := width - statsWidth - 2
	title := ui.ValueStyle.Bold(true).Render(truncatePathLeft(name, max(nameWidth, 1)))

	gap := width - lipgloss.Width(title) - statsWidth
	if gap < 1 {
		gap = 1
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		title+strings.Repeat(" ", gap)+stats,
		ui.SubtleStyle.Render(strings.Repeat("─", max(width, 1))),
	)
}

// renderDiffBody lays out the parsed diff with a two-column line-number gutter,
// the way `delta` and GitHub present a unified diff.
func (m *Model) renderDiffBody(width int) string {
	if m.diffFetching {
		return ui.SpinnerStyle.Render("  " + m.spinnerView() + " Loading diff…")
	}
	if len(m.parsedDiff.Lines) == 0 {
		return ui.SubtleStyle.Render("  No changes to show")
	}

	gutter := diffGutterWidth*2 + 3
	textWidth := width - gutter - 1
	if textWidth < 10 {
		textWidth = 10
	}

	lines := make([]string, 0, len(m.parsedDiff.Lines))
	for _, dl := range m.parsedDiff.Lines {
		if isDiffNoise(dl) {
			continue
		}
		lines = append(lines, m.renderDiffRow(dl, textWidth))
	}
	if len(lines) == 0 {
		return ui.SubtleStyle.Render("  No changes to show")
	}
	return strings.Join(lines, "\n")
}

// isDiffNoise reports rows that repeat what the header already says. The blob
// hashes and the a/b paths carry nothing a reader needs while the file name is
// on screen; the mode and rename lines do, so they stay.
func isDiffNoise(dl diffLine) bool {
	switch dl.Kind {
	case diffFileHeader:
		return true
	case diffMeta:
		return strings.HasPrefix(dl.Text, "diff --git") ||
			strings.HasPrefix(dl.Text, "index ") ||
			strings.HasPrefix(dl.Text, "similarity index")
	}
	return false
}

func (m *Model) renderDiffRow(dl diffLine, textWidth int) string {
	switch dl.Kind {
	case diffHunkHeader:
		return ui.DiffHunkStyle.Render(truncateRunes(dl.Text, textWidth+diffGutterWidth*2+3))
	case diffFileHeader, diffMeta:
		return ui.SubtleStyle.Render(truncateRunes(dl.Text, textWidth+diffGutterWidth*2+3))
	}

	oldNum := blankNumber()
	newNum := blankNumber()
	if dl.OldLine > 0 {
		oldNum = fmt.Sprintf("%*d", diffGutterWidth, dl.OldLine)
	}
	if dl.NewLine > 0 {
		newNum = fmt.Sprintf("%*d", diffGutterWidth, dl.NewLine)
	}

	var sign string
	var style lipgloss.Style
	switch dl.Kind {
	case diffAdded:
		sign, style = "+", ui.DiffAddStyle
	case diffRemoved:
		sign, style = "-", ui.DiffDelStyle
	default:
		sign, style = " ", ui.ValueStyle
	}

	text := truncateRunes(dl.Text, textWidth)
	return ui.GutterStyle.Render(oldNum+" "+newNum) +
		ui.SubtleStyle.Render("│") +
		style.Render(sign+text)
}

func blankNumber() string {
	return strings.Repeat(" ", diffGutterWidth)
}

// refreshDiffViewport re-renders the diff into its viewport. Called whenever the
// diff, the selected file or the available width changes.
func (m *Model) refreshDiffViewport() {
	width := m.diffViewport.Width()
	if width <= 0 {
		width = 80
	}
	m.diffViewport.SetContent(m.renderDiffBody(width))
}

// jumpHunk moves the diff viewport to the next or previous hunk header.
func (m *Model) jumpHunk(direction int) {
	if len(m.parsedDiff.HunkStarts) == 0 {
		m.statusMsg = "No hunks in this diff"
		return
	}

	current := m.diffViewport.YOffset()
	target := -1
	if direction > 0 {
		target = m.parsedDiff.hunkAfter(current)
	} else {
		target = m.parsedDiff.hunkBefore(current)
	}

	if target < 0 {
		if direction > 0 {
			m.statusMsg = "Last hunk"
		} else {
			m.statusMsg = "First hunk"
		}
		return
	}
	m.diffViewport.SetYOffset(target)
}

// toggleDiffViewer opens the full-screen diff, or closes it when already open.
func (m *Model) toggleDiffViewer() (tea.Model, tea.Cmd) {
	if m.showDiff() {
		m.setDetailView(DetailLog)
		m.activePanel = RepoPanel
		m.refreshViewports()
		return m, nil
	}

	r := m.selectedRepo()
	if r == nil {
		return m, nil
	}

	m.cancelSpecialModes()
	m.setDetailView(DetailDiff)
	m.activePanel = DiffPanel
	m.fileCursor = 0
	m.statusMsg = ""
	return m, m.fetchFilesCmd(r.Path)
}

// handleDiffKeys drives the full-screen diff: files with j/k, hunks with J/K,
// and the diff body with the usual scrolling motions.
func (m *Model) handleDiffKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	seq, absorbed := m.consumePending(msg)
	if absorbed {
		return m, nil
	}
	count := m.countOr(1)
	m.clearPending()

	switch {
	case seq == "esc" || matchesSeq(seq, keys.Diff...) || seq == "q":
		m.setDetailView(DetailLog)
		m.activePanel = RepoPanel
		m.refreshViewports()
		return m, nil

	case seq == "ctrl+c":
		m.quitting = true
		return m, tea.Quit

	case matchesSeq(seq, keys.Help...) || matchesSeq(seq, keys.HelpAlt...):
		m.pushOverlay(OverlayHelp)
		m.activePanel = HelpPanel
		m.helpSearchInput.Reset()
		m.helpViewport.GotoTop()
		return m, m.helpSearchInput.Focus()

	case seq == "j" || seq == "down":
		return m.moveDiffFile(count)

	case seq == "k" || seq == "up":
		return m.moveDiffFile(-count)

	case seq == "J":
		for range count {
			m.jumpHunk(1)
		}
		return m, nil

	case seq == "K":
		for range count {
			m.jumpHunk(-1)
		}
		return m, nil

	case matchesSeq(seq, keys.HalfPageDown...):
		m.diffViewport.ScrollDown(m.diffViewport.Height() / 2 * count)
		return m, nil

	case matchesSeq(seq, keys.HalfPageUp...):
		m.diffViewport.ScrollUp(m.diffViewport.Height() / 2 * count)
		return m, nil

	case seq == "gg":
		m.diffViewport.GotoTop()
		return m, nil

	case matchesSeq(seq, keys.Bottom...):
		m.diffViewport.GotoBottom()
		return m, nil

	case matchesSeq(seq, keys.Space...):
		return m.toggleDiffFileStaged()

	case matchesSeq(seq, keys.Discard...):
		if m.fileCursor < len(m.files) {
			file := m.files[m.fileCursor]
			return m.promptConfirm(
				"Discard changes in '"+file.Name+"'?",
				"This will restore the file from Git.",
				"discard",
			)
		}
		return m, nil

	case matchesSeq(seq, keys.Copy...):
		return m.copyCurrentSelection()
	}

	return m, nil
}

func (m *Model) moveDiffFile(delta int) (tea.Model, tea.Cmd) {
	if len(m.files) == 0 {
		return m, nil
	}
	next := clamp(m.fileCursor+delta, 0, len(m.files)-1)
	if next == m.fileCursor {
		return m, nil
	}
	m.fileCursor = next

	r := m.selectedRepo()
	if r == nil {
		return m, nil
	}
	m.diffFetching = true
	m.currentDiff = ""
	m.parsedDiff = parsedDiff{}
	m.diffViewport.GotoTop()
	return m, m.fetchDiffCmd(r.Path, m.files[m.fileCursor])
}

func (m *Model) toggleDiffFileStaged() (tea.Model, tea.Cmd) {
	if m.fileCursor >= len(m.files) {
		return m, nil
	}
	r := m.selectedRepo()
	if r == nil {
		return m, nil
	}
	file := m.files[m.fileCursor]
	return m, m.toggleFileCmd(r.Path, file)
}

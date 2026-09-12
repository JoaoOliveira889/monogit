package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/JoaoOliveira889/monogit/internal/pkg/ui"
)

func (m *Model) View() string {
	if m.quitting {
		return ""
	}

	if m.showSplash {
		return m.renderSplash()
	}

	if m.width < minTerminalWidth || m.height < minTerminalHeight {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			ui.ErrorStyle.Render(fmt.Sprintf("Terminal too small.\nPlease resize to at least %d×%d.", minTerminalWidth, minTerminalHeight)),
		)
	}

	// The innermost overlay owns the whole frame, so return before laying out
	// the body underneath it.
	if overlay, ok := m.topOverlay(); ok && overlay.replacesFrame() {
		return m.renderOverlay(overlay)
	}

	if m.activePanel == CommitWizardPanel {
		return m.renderCenteredModal(m.renderCommitWizardModal())
	}

	m.syncViewports()

	view := lipgloss.JoinVertical(lipgloss.Left,
		m.renderHeader(),
		m.renderBody(),
		m.renderFooter(),
	)

	framed := lipgloss.NewStyle().
		MaxWidth(m.width).
		MaxHeight(m.height).
		Render(view)

	return m.overlayWhichKey(framed)
}

// refreshViewports marks the panel viewports as stale. Rendering them is
// deferred to the next View, so a burst of messages — one status update per
// repository after a scan — costs one render instead of one per message.
func (m *Model) refreshViewports() {
	m.viewportsDirty = true
}

// syncViewports re-renders the stale viewports. View calls it once, right
// before laying out the body.
func (m *Model) syncViewports() {
	if !m.viewportsDirty {
		return
	}
	m.viewportsDirty = false

	// Scroll offsets first: the list renderers only style the rows inside the
	// visible window, so they need the final offsets.
	m.syncScrollPositions()

	m.viewport.SetContent(m.renderViewportContent())
	m.repoViewport.SetContent(m.renderRepoViewportContent())
	m.fileViewport.SetContent(m.renderFileViewportContent())
	m.refreshLogViewport()
}

func (m *Model) renderCenteredModal(content string) string {
	modalStyle := ui.ActivePanelStyle.
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ui.ColorHighlight)).
		Width(m.modalWidthForContent(content)).
		Padding(1, 2)
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		modalStyle.Render(content),
	)
}

func (m Model) modalWidthForContent(content string) int {
	longest := 0
	for _, line := range strings.Split(content, "\n") {
		if width := lipgloss.Width(line); width > longest {
			longest = width
		}
	}

	width := longest + 6
	if width < 42 {
		width = 42
	}
	if width > 72 {
		width = 72
	}
	if maxWidth := m.width - 4; maxWidth > 0 && width > maxWidth {
		width = maxWidth
	}
	return width
}

// renderOverlay draws one overlay as a full frame.
func (m *Model) renderOverlay(overlay Overlay) string {
	switch overlay {
	case OverlayConfirm:
		return m.renderCenteredModal(m.renderConfirmationModal())
	case OverlayInput:
		return m.renderCenteredModal(m.renderInputModal())
	case OverlayHelp:
		return m.renderHelpOverlay()
	case OverlayPalette:
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, m.renderPalette())
	case OverlayEditorPicker:
		return m.renderCenteredModal(m.renderEditorModal())
	case OverlayStatusFilter:
		return m.renderModalShell(
			"Filter Repositories",
			m.renderFilterModal(m.width-8, m.height-8),
			m.joinFooterKeys(
				m.fmtKey("↑↓", "navigate"),
				m.fmtKey("enter", "select"),
				m.fmtKey("esc", "cancel"),
			),
		)
	case OverlayTagFilter:
		return m.renderModalShell(
			"Filter by Tags",
			m.renderTagFilterModal(m.width-8, m.height-8),
			m.joinFooterKeys(
				m.fmtKey("↑↓", "navigate"),
				m.fmtKey("space", "toggle"),
				m.fmtKey("enter", "apply"),
				m.fmtKey("esc", "cancel"),
			),
		)
	}
	return ""
}

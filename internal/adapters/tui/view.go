package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	"github.com/JoaoOliveira889/monogit/internal/pkg/ui"
)

const minTerminalWidth = 60
const minTerminalHeight = 10

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

	header := m.renderHeader()
	footer := m.renderFooter()
	body := m.renderBody()

	view := lipgloss.JoinVertical(lipgloss.Left,
		header,
		body,
		footer,
	)

	if m.showConfirmModal {
		return m.renderCenteredModal(m.renderConfirmationModal())
	}

	if m.inputMode {
		return m.renderCenteredModal(m.renderInputModal())
	}

	if m.showHelp {
		return m.renderHelpOverlay()
	}

	if m.showEditorModal {
		return m.renderCenteredModal(m.renderEditorModal())
	}

	if m.tagFilterModal {
		return m.renderModalShell(
			"Filter by Tags",
			m.renderTagFilterModal(m.width-8, m.height-8),
			"↑↓ navigate   space toggle   enter apply   esc cancel",
		)
	}

	if m.activePanel == CommitWizardPanel {
		return m.renderCenteredModal(m.renderCommitWizardModal())
	}

	return lipgloss.NewStyle().
		MaxWidth(m.width).
		MaxHeight(m.height).
		Render(view)
}

func (m *Model) refreshViewports() {
	m.viewport.SetContent(m.renderViewportContent())
	m.repoViewport.SetContent(m.renderRepoViewportContent())
	m.fileViewport.SetContent(m.renderFileViewportContent())
	m.syncScrollPositions()
}

func (m *Model) renderCenteredModal(content string) string {
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		ui.ActivePanelStyle.Padding(1, 2).Render(content),
	)
}

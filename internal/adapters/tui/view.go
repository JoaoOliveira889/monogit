package tui

import (
	"github.com/charmbracelet/lipgloss"

	"monogit/internal/pkg/ui"
)

func (m *Model) View() string {
	if m.quitting {
		return ""
	}

	if m.width < 60 || m.height < 10 {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			ui.ErrorStyle.Render("Terminal too small.\nPlease resize to at least 60×10."),
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
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			ui.ActivePanelStyle.Padding(1, 2).Render(m.renderConfirmationModal()),
		)
	}

	if m.inputMode {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			ui.ActivePanelStyle.Padding(1, 2).Render(m.renderInputModal()),
		)
	}

	if m.showHelp {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			ui.ActivePanelStyle.Padding(1, 2).Render(m.renderHelpMenu()),
		)
	}

	if m.activePanel == CommitWizardPanel {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			ui.ActivePanelStyle.Padding(1, 2).Render(m.renderCommitWizardModal()),
		)
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

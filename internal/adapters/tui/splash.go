package tui

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/JoaoOliveira889/monogit/internal/pkg/ui"
)

func (m *Model) renderSplash() string {
	status := ui.SpinnerStyle.Render(m.spinnerView() + " Opening MonoGit...")
	subtitle := ui.SubtleStyle.Render("Temporary log sessions and explicit actions, by design.")

	body := lipgloss.JoinVertical(lipgloss.Center,
		renderBrandWordmark(false),
		"",
		ui.ValueStyle.Render("Multi-repo Git dashboard for your terminal."),
		"",
		status,
		ui.SubtleStyle.Render(" "+spinnerFrames[m.splashFrame%len(spinnerFrames)]+" starting up"),
		"",
		subtitle,
	)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, body)
}

func renderBrandWordmark(compact bool) string {
	mono := ui.BrandMonoStyle
	git := ui.BrandGitStyle
	subtle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ui.ColorSubtle))

	if compact {
		return lipgloss.JoinHorizontal(lipgloss.Bottom,
			mono.Render("Mono"),
			git.Render("Git"),
		)
	}

	return lipgloss.JoinVertical(lipgloss.Center,
		lipgloss.JoinHorizontal(lipgloss.Top,
			mono.Render("Mono"),
			git.Render("Git"),
		),
		subtle.Render("multi-repo git dashboard"),
	)
}

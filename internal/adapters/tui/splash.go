package tui

import (
	"github.com/charmbracelet/lipgloss"

	"monogit/internal/pkg/ui"
)

func (m *Model) renderSplash() string {
	status := ui.SpinnerStyle.Render(m.spinnerView() + " Opening MonoGit...")
	subtitle := ui.SubtleStyle.Render("Temporary log sessions and explicit actions, by design.")

	body := lipgloss.JoinVertical(lipgloss.Center,
		renderSplashWordmark(),
		"",
		ui.ValueStyle.Render(" Multi-repo Git dashboard for your terminal."),
		"",
		status,
		ui.SubtleStyle.Render(" "+spinnerFrames[m.splashFrame%len(spinnerFrames)]+" starting up"),
		"",
		subtitle,
	)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, body)
}

func renderSplashWordmark() string {
	accent := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ui.ColorHighlight)).
		Bold(true)
	cyan := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ui.ColorCyan)).
		Bold(true)
	subtle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ui.ColorSubtle))

	return lipgloss.JoinVertical(lipgloss.Center,
		cyan.Render("   ◉────────◉   ◉────────◉"),
		subtle.Render(" ╭──┤                    ├──╮"),
		accent.Render(" │       M O N O G I T      │"),
		subtle.Render(" ╰──┤                    ├──╯"),
		cyan.Render("   ◉────────◉   ◉────────◉"),
	)
}

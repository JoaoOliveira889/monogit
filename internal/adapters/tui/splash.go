package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/JoaoOliveira889/monogit/internal/pkg/ui"
)

// splashDots renders an animated "scanning…" indicator using the current splash
// frame.  Three dots pulse in sequence: " ·  ", "  · ", "   ·", "·   ".
func (m *Model) splashProgressDots() string {
	frames := []string{"⠋ scanning", "⠙ scanning", "⠹ scanning", "⠸ scanning", "⠼ scanning", "⠴ scanning", "⠦ scanning", "⠧ scanning", "⠇ scanning", "⠏ scanning"}
	return frames[m.splashFrame%len(frames)]
}

func (m *Model) renderSplash() string {
	// Progress bar — filled based on splashFrame animation
	barWidth := 20
	filled := (m.splashFrame * 2) % (barWidth + 1)
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
	progressBar := ui.SpinnerStyle.Render(bar)

	scanStatus := ui.SpinnerStyle.Render(m.splashProgressDots() + "…")
	version := ui.SubtleStyle.Render("v" + Version)
	subtitle := ui.SubtleStyle.Render("Multi-repo Git dashboard for your terminal")

	body := lipgloss.JoinVertical(lipgloss.Center,
		renderBrandWordmark(false),
		"",
		subtitle,
		"",
		progressBar,
		"",
		scanStatus,
		"",
		version,
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

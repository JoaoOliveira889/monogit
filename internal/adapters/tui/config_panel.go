package tui

import (
	"fmt"
	"strings"

	"github.com/JoaoOliveira889/monogit/internal/pkg/ui"
	"github.com/charmbracelet/lipgloss"
)

const (
	configMergeToolIdx    = 1
	configScanExcludesIdx = 2
	configThemeIdx        = 3
	numConfigOptions      = 4
)

// configOption defines a single row in the config panel.
type configOption struct {
	icon  string
	name  string
	value string
	desc  string
}

func (m *Model) configOptions() []configOption {
	return []configOption{
		{
			icon:  "⇄",
			name:  "Left Panel Width Ratio",
			value: fmt.Sprintf("%d%%", int(m.leftPanelRatio*100)),
			desc:  "Use '<' and '>' keys to adjust the panel width ratio",
		},
		{
			icon:  "⚙",
			name:  "Merge Tool Command",
			value: m.cfg.MergeTool,
			desc:  "Git merge tool command — press Enter to edit",
		},
		{
			icon:  "⊘",
			name:  "Scan Exclude Folders",
			value: strings.Join(m.cfg.ScanExcludes, ", "),
			desc:  "Comma-separated folder names to skip during scan — press Enter to edit",
		},
		{
			icon:  "◑",
			name:  "Color Theme",
			value: m.cfg.Theme,
			desc:  "Visual color theme — press Enter to cycle through themes",
		},
	}
}

func (m *Model) renderConfigPanel(width int) string {
	var sb strings.Builder
	contentWidth := width - 6
	if contentWidth < 12 {
		contentWidth = 12
	}

	// Header
	sep := ui.ConfigSepStyle.Render(strings.Repeat("─", contentWidth))
	sb.WriteString("\n  " + ui.LabelStyle.Render("⚡ Configuration") + "\n")
	sb.WriteString("  " + sep + "\n\n")

	options := m.configOptions()

	for i, opt := range options {
		selected := i == m.configCursor

		// Cursor arrow
		pointer := "   "
		if selected {
			pointer = lipgloss.NewStyle().Foreground(ui.ColorCyan).Bold(true).Render(" ▶ ")
		}

		// Icon + name
		icon := opt.icon + " "
		var nameStr string
		if selected {
			nameStr = lipgloss.NewStyle().Foreground(ui.ColorFg).Bold(true).Render(icon + truncateRunes(opt.name, contentWidth-6))
		} else {
			nameStr = ui.ConfigItemStyle.Render(icon + truncateRunes(opt.name, contentWidth-6))
		}

		// Value
		value := opt.value
		var valueStr string
		if value == "" {
			valueStr = ui.SubtleStyle.Render("(not set)")
		} else if selected {
			valueStr = lipgloss.NewStyle().Foreground(ui.ColorCyan).Bold(true).Render(truncateRunes(value, contentWidth-5))
		} else {
			valueStr = ui.ConfigValueStyle.Foreground(ui.ColorHighlight).Render(truncateRunes(value, contentWidth-5))
		}

		sb.WriteString(pointer + nameStr + "\n")
		sb.WriteString("     " + valueStr + "\n")

		if selected {
			desc := truncateRunes(opt.desc, contentWidth-5)
			sb.WriteString("     " + ui.SubtleStyle.Italic(true).Render(desc) + "\n\n")
		} else {
			sb.WriteString("\n")
		}
	}

	sb.WriteString("  " + sep + "\n")
	note := truncateRunes("↑↓ navigate  •  Enter edit/select  •  Esc close", contentWidth)
	sb.WriteString("  " + ui.SubtleStyle.Render(note))

	return sb.String()
}

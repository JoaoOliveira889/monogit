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

func (m *Model) renderConfigPanel(width int) string {
	var sb strings.Builder
	contentWidth := width - 6
	if contentWidth < 12 {
		contentWidth = 12
	}

	sb.WriteString("\n  " + ui.LabelStyle.Render("Interactive Settings") + "\n\n")

	options := []struct {
		name  string
		value string
		desc  string
	}{
		{
			name:  "Left Panel Width Ratio",
			value: fmt.Sprintf("%d%%", int(m.leftPanelRatio*100)),
			desc:  "Use '<' and '>' keys to adjust width ratio",
		},
		{
			name:  "Merge Tool Command",
			value: m.cfg.MergeTool,
			desc:  "Default Git merge tool command (Press Enter to edit)",
		},
		{
			name:  "Scan Exclude Folders",
			value: strings.Join(m.cfg.ScanExcludes, ", "),
			desc:  "Comma-separated folder names to ignore (Press Enter to edit)",
		},
		{
			name:  "Color Theme",
			value: m.cfg.Theme,
			desc:  "Theme scheme for visual interface (Press Enter to change)",
		},
	}

	for i, opt := range options {
		selected := i == m.configCursor
		bg := ui.ColorBg
		if selected {
			bg = ui.ColorHighlight
		}

		bgStyle := lipgloss.NewStyle().Background(bg)

		prefix := "   "
		if selected {
			prefix = " > "
			prefix = bgStyle.Render(prefix)
		}

		nameStyle := lipgloss.NewStyle().Foreground(ui.ColorFg)
		if selected {
			nameStyle = ui.SelectedItemStyle
		}
		nameStr := nameStyle.Render(truncateRunes(opt.name, contentWidth-3))

		valueStyle := lipgloss.NewStyle().Foreground(ui.ColorHighlight)
		if selected {
			valueStyle = ui.SelectedItemStyle
		}
		value := truncateRunes(opt.value, contentWidth-5)
		valueStr := valueStyle.Render(value)
		if opt.value == "" {
			valueStr = ui.SubtleStyle.Render("(not set)")
		}

		sb.WriteString(prefix + nameStr + "\n")
		sb.WriteString("     " + valueStr + "\n")

		if selected {
			desc := truncateRunes(opt.desc, contentWidth-5)
			sb.WriteString("     " + ui.SubtleStyle.Render(desc) + "\n\n")
		} else {
			sb.WriteString("\n")
		}
	}

	note := truncateRunes("Press Esc to close and return to repository dashboard.", contentWidth)
	sb.WriteString("\n  " + ui.SubtleStyle.Render(note))

	return sb.String()
}

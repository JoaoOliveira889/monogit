package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/JoaoOliveira889/monogit/internal/pkg/ui"
)

const (
	configLeftPanelRatioIdx = 0
	configMergeToolIdx      = 1
	configScanExcludesIdx   = 2
	numConfigOptions        = 3
)

func (m *Model) renderConfigPanel(width int) string {
	var sb strings.Builder

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
		nameStr := nameStyle.Render(opt.name)

		valueStyle := lipgloss.NewStyle().Foreground(ui.ColorHighlight)
		if selected {
			valueStyle = ui.SelectedItemStyle
		}
		valueStr := valueStyle.Render(opt.value)
		if opt.value == "" {
			valueStr = ui.SubtleStyle.Render("(not set)")
		}

		line := fmt.Sprintf("%s%-24s : %s", prefix, nameStr, valueStr)
		sb.WriteString(line + "\n")

		if selected {
			sb.WriteString("     " + ui.SubtleStyle.Render(opt.desc) + "\n\n")
		} else {
			sb.WriteString("\n")
		}
	}

	sb.WriteString("\n  " + ui.SubtleStyle.Render("Press Esc to close and return to repository dashboard."))

	return sb.String()
}

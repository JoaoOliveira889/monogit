package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/JoaoOliveira889/monogit/internal/pkg/ui"
)

func (m *Model) renderRebasePanel(width int) string {
	var sb strings.Builder
	contentWidth := width - 6
	if contentWidth < 12 {
		contentWidth = 12
	}

	if m.rebaseFetching {
		return ui.SpinnerStyle.Render("   " + m.spinnerView() + " Loading commits for rebase...")
	}

	if len(m.rebaseItems) == 0 {
		return ui.SubtleStyle.Render("   No commits available for interactive rebase.")
	}

	// Header
	sep := ui.ConfigSepStyle.Render(strings.Repeat("─", contentWidth))
	sb.WriteString("\n  " + ui.LabelStyle.Render("🔀 Interactive Rebase — Select actions for commits") + "\n")
	sb.WriteString("  " + sep + "\n\n")

	for i, item := range m.rebaseItems {
		selected := i == m.rebaseCursor

		pointer := "   "
		if selected {
			pointer = lipgloss.NewStyle().Foreground(ui.ColorCyan).Bold(true).Render(" ▶ ")
		}

		// Action badge
		var badge string
		switch item.Action {
		case "pick":
			badge = lipgloss.NewStyle().Background(ui.ColorSuccess).Foreground(ui.ColorBg).Bold(true).Render(" PICK ")
		case "squash":
			badge = lipgloss.NewStyle().Background(ui.ColorWarning).Foreground(ui.ColorBg).Bold(true).Render(" SQUASH ")
		case "fixup":
			badge = lipgloss.NewStyle().Background(ui.ColorCyan).Foreground(ui.ColorBg).Bold(true).Render(" FIXUP ")
		case "reword":
			badge = lipgloss.NewStyle().Background(ui.ColorHighlight).Foreground(ui.ColorBg).Bold(true).Render(" REWORD ")
		case "drop":
			badge = lipgloss.NewStyle().Background(ui.ColorError).Foreground(ui.ColorBg).Bold(true).Render(" DROP ")
		default:
			badge = lipgloss.NewStyle().Background(ui.ColorSubtle).Foreground(ui.ColorBg).Render(" " + strings.ToUpper(item.Action) + " ")
		}

		hashStr := ui.BranchStyle.Render(item.Hash)
		if selected {
			hashStr = ui.BranchStyle.Bold(true).Render(item.Hash)
		}

		msgLen := contentWidth - lipgloss.Width(pointer) - lipgloss.Width(badge) - lipgloss.Width(hashStr) - 6
		if msgLen < 5 {
			msgLen = 5
		}
		msgStr := truncateRunes(item.Message, msgLen)
		if selected {
			msgStr = lipgloss.NewStyle().Foreground(ui.ColorFg).Bold(true).Render(msgStr)
		} else {
			msgStr = ui.ValueStyle.Render(msgStr)
		}

		line := fmt.Sprintf("%s%s  %s  %s", pointer, badge, hashStr, msgStr)
		sb.WriteString(line + "\n")
	}

	sb.WriteString("\n  " + sep + "\n")
	actionsHelp := ui.SubtleStyle.Render("  Actions: (p)ick  •  (s)quash  •  (f)ixup  •  (r)eword  •  (d)rop")
	reorderHelp := ui.SubtleStyle.Render("  Move: (J)/(K) or (Shift+↓/↑) reorder  •  (Enter) Run Rebase  •  (Esc) Cancel")
	sb.WriteString(actionsHelp + "\n" + reorderHelp + "\n")

	return sb.String()
}

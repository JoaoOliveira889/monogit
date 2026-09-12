package tui

import (
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/JoaoOliveira889/monogit/internal/pkg/ui"
)

// renderWhichKey draws the continuations available for a half-typed sequence,
// so a prefix key is discoverable instead of something to memorise.
func (m *Model) renderWhichKey() string {
	continuations, ok := prefixKeys[m.pending.prefix]
	if !ok || len(continuations) == 0 {
		return ""
	}

	var rows []string
	for _, c := range continuations {
		rows = append(rows,
			ui.FooterKeyStyle.Render(c.Key)+"  "+ui.FooterActionStyle.Render(c.Action))
	}

	title := ui.LabelStyle.Render(m.pending.describe())
	body := strings.Join(rows, "\n")

	panel := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ui.ColorCyan).
		Padding(0, 1).
		Render(title + "\n" + body)

	return panel
}

// overlayWhichKey places the which-key panel in the bottom-right corner, above
// the footer, without disturbing the layout underneath.
func (m *Model) overlayWhichKey(view string) string {
	if m.pending.prefix == "" {
		return view
	}
	panel := m.renderWhichKey()
	if panel == "" {
		return view
	}

	lines := strings.Split(view, "\n")
	panelLines := strings.Split(panel, "\n")
	panelWidth := lipgloss.Width(panel)

	// Sit just above the footer.
	bottom := len(lines) - 2
	top := bottom - len(panelLines)
	if top < 0 {
		return view
	}

	left := m.width - panelWidth - 2
	if left < 0 {
		return view
	}

	for i, panelLine := range panelLines {
		row := top + i
		if row < 0 || row >= len(lines) {
			continue
		}
		lines[row] = overlayAt(lines[row], panelLine, left, m.width)
	}
	return strings.Join(lines, "\n")
}

// overlayAt writes replacement into line starting at column, padding the line
// first when it is shorter than the requested column.
func overlayAt(line, replacement string, column, width int) string {
	plain := lipgloss.Width(line)
	if plain < column {
		line += strings.Repeat(" ", column-plain)
		return line + replacement
	}

	// Truncating styled text by rune index would cut ANSI sequences in half, so
	// rebuild the row from the truncated prefix instead.
	prefix := truncateRunes(stripANSI(line), column)
	prefixWidth := lipgloss.Width(prefix)
	if prefixWidth < column {
		prefix += strings.Repeat(" ", column-prefixWidth)
	}
	result := prefix + replacement
	if lipgloss.Width(result) > width {
		return line
	}
	return result
}

// stripANSI removes SGR escape sequences so a styled line can be measured and
// cut by visible characters.
func stripANSI(s string) string {
	var sb strings.Builder
	sb.Grow(len(s))
	inEscape := false
	for _, r := range s {
		switch {
		case inEscape:
			if r == 'm' {
				inEscape = false
			}
		case r == '\x1b':
			inEscape = true
		default:
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

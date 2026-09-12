package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
)

// keyPress builds the v2 key message for a keystroke written the way the
// keymaps spell it, such as "q", "ctrl+c" or "enter".
func keyPress(keystroke string) tea.KeyPressMsg {
	named := map[string]rune{
		"enter":     tea.KeyEnter,
		"esc":       tea.KeyEscape,
		"tab":       tea.KeyTab,
		"space":     tea.KeySpace,
		"up":        tea.KeyUp,
		"down":      tea.KeyDown,
		"left":      tea.KeyLeft,
		"right":     tea.KeyRight,
		"home":      tea.KeyHome,
		"end":       tea.KeyEnd,
		"pgup":      tea.KeyPgUp,
		"pgdown":    tea.KeyPgDown,
		"backspace": tea.KeyBackspace,
		"delete":    tea.KeyDelete,
	}

	var mod tea.KeyMod
	for {
		switch {
		case strings.HasPrefix(keystroke, "ctrl+"):
			mod |= tea.ModCtrl
			keystroke = strings.TrimPrefix(keystroke, "ctrl+")
			continue
		case strings.HasPrefix(keystroke, "alt+"):
			mod |= tea.ModAlt
			keystroke = strings.TrimPrefix(keystroke, "alt+")
			continue
		case strings.HasPrefix(keystroke, "shift+"):
			mod |= tea.ModShift
			keystroke = strings.TrimPrefix(keystroke, "shift+")
			continue
		}
		break
	}

	if code, ok := named[keystroke]; ok {
		key := tea.KeyPressMsg{Code: code, Mod: mod}
		if code == tea.KeySpace {
			key.Text = " "
		}
		return key
	}

	runes := []rune(keystroke)
	if len(runes) != 1 {
		return tea.KeyPressMsg{Mod: mod}
	}

	key := tea.KeyPressMsg{Code: runes[0], Mod: mod}
	if mod == 0 {
		key.Text = keystroke
	}
	return key
}

// keyText builds a key message carrying literal text, used to simulate the
// escape-sequence fragments a terminal can leak as ordinary input.
func keyText(text string) tea.KeyPressMsg {
	runes := []rune(text)
	key := tea.KeyPressMsg{Text: text}
	if len(runes) > 0 {
		key.Code = runes[0]
	}
	return key
}

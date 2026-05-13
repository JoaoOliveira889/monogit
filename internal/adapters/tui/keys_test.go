package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestMatchesKey(t *testing.T) {
	tests := []struct {
		name string
		key  string
		list []string
		want bool
	}{
		{"exact match", "q", []string{"q", "ctrl+c"}, true},
		{"second match", "ctrl+c", []string{"q", "ctrl+c"}, true},
		{"no match", "x", []string{"q", "ctrl+c"}, false},
		{"empty list", "q", []string{}, false},
		{"partial match", "k", []string{"up", "k"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
			if len(tt.key) == 0 {
				msg = tea.KeyMsg{Type: tea.KeyCtrlC}
			} else {
				msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
			}
			got := matchesKey(msg, tt.list...)
			if got != tt.want {
				t.Errorf("matchesKey(%q, %v) = %v, want %v", tt.key, tt.list, got, tt.want)
			}
		})
	}
}

func TestKeyMapValues(t *testing.T) {
	if len(keys.Quit) == 0 {
		t.Error("Quit keys should not be empty")
	}
	if len(keys.Help) == 0 {
		t.Error("Help keys should not be empty")
	}
	if len(keys.Up) == 0 {
		t.Error("Up keys should not be empty")
	}
	if len(keys.Down) == 0 {
		t.Error("Down keys should not be empty")
	}
	if len(keys.Enter) == 0 {
		t.Error("Enter keys should not be empty")
	}

	if keys.Quit[0] != "q" {
		t.Errorf("expected quit key 'q', got %s", keys.Quit[0])
	}
}

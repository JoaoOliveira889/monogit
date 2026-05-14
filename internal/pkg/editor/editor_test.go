package editor

import (
	"fmt"
	"testing"
)

func TestNewLauncher(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{"vim", "*editor.TerminalLauncher"},
		{"nvim", "*editor.TerminalLauncher"},
		{"/usr/local/bin/nano", "*editor.TerminalLauncher"},
		{"code", "*editor.GUILauncher"},
		{"visual studio code", "*editor.GUILauncher"},
		{"Rider", "*editor.GUILauncher"},
		{"TextEdit", "*editor.AppLauncher"},
		{"Xcode (App)", "*editor.AppLauncher"},
	}

	for _, tt := range tests {
		l := NewLauncher(tt.name)
		actual := fmt.Sprintf("%T", l)
		if actual != tt.expected {
			t.Errorf("NewLauncher(%s) = %s, want %s", tt.name, actual, tt.expected)
		}
	}
}

func TestTerminalLauncherEscapePath(t *testing.T) {
	l := &TerminalLauncher{Editor: "vim"}
	path := "/path/to/my repo"
	escaped := l.escapePath(path)
	expected := "\"/path/to/my repo\""
	if escaped != expected {
		t.Errorf("escapePath(%s) = %s, want %s", path, escaped, expected)
	}

	pathWithQuotes := "/path/\"quoted\" repo"
	escaped = l.escapePath(pathWithQuotes)
	expected = "\"/path/\\\"quoted\\\" repo\""
	if escaped != expected {
		t.Errorf("escapePath(%s) = %s, want %s", pathWithQuotes, escaped, expected)
	}
}

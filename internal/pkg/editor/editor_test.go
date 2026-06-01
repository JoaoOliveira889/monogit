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

func TestParseCommand(t *testing.T) {
	spec, err := ParseCommand("code --reuse-window")
	if err != nil {
		t.Fatalf("ParseCommand failed: %v", err)
	}
	if spec.Name != "code" {
		t.Fatalf("expected code, got %q", spec.Name)
	}
	if len(spec.Args) != 1 || spec.Args[0] != "--reuse-window" {
		t.Fatalf("unexpected args: %+v", spec.Args)
	}
}

func TestTerminalLauncherCommandLine(t *testing.T) {
	l := &TerminalLauncher{Spec: CommandSpec{Name: "vim", Args: []string{"-p"}}}
	got := l.commandLine("/path/it's repo")
	want := "'vim' '-p' '/path/it'\"'\"'s repo'"
	if got != want {
		t.Fatalf("commandLine() = %q, want %q", got, want)
	}
}

func TestValidateAppName(t *testing.T) {
	if err := ValidateAppName("Visual Studio Code"); err != nil {
		t.Fatalf("expected valid app name, got %v", err)
	}
	if err := ValidateAppName("Bad\nApp"); err == nil {
		t.Fatal("expected invalid app name to be rejected")
	}
}

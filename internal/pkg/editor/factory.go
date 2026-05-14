package editor

import (
	"path/filepath"
	"strings"
)

func NewLauncher(editorName string) Launcher {
	if strings.HasSuffix(editorName, "(App)") {
		return &AppLauncher{AppName: strings.TrimSpace(strings.TrimSuffix(editorName, "(App)"))}
	}

	if editorName == "TextEdit" {
		return &AppLauncher{AppName: "TextEdit"}
	}

	base := strings.ToLower(filepath.Base(editorName))
	
	terminalEditors := map[string]bool{
		"vim":   true,
		"nvim":  true,
		"nano":  true,
		"vi":    true,
		"emacs": true,
	}

	if terminalEditors[base] {
		return &TerminalLauncher{Editor: editorName}
	}

	return &GUILauncher{Editor: editorName}
}

package editor

import (
	"strings"
)

func NewLauncher(editorName string) Launcher {
	if strings.HasSuffix(editorName, "(App)") {
		return &AppLauncher{AppName: strings.TrimSpace(strings.TrimSuffix(editorName, "(App)"))}
	}

	if editorName == "TextEdit" {
		return &AppLauncher{AppName: "TextEdit"}
	}

	spec, err := ParseCommand(editorName)
	if err != nil {
		return &GUILauncher{Spec: CommandSpec{Name: strings.TrimSpace(editorName)}}
	}

	if IsTerminalEditor(editorName) {
		return &TerminalLauncher{Spec: spec}
	}

	return &GUILauncher{Spec: spec}
}

package editor

import (
	"fmt"
	"path/filepath"
	"strings"
	"unicode"
)

type Launcher interface {
	Launch(path string) error
}

type CommandSpec struct {
	Name string
	Args []string
}

func ParseCommand(input string) (CommandSpec, error) {
	fields := strings.Fields(strings.TrimSpace(input))
	if len(fields) == 0 {
		return CommandSpec{}, fmt.Errorf("empty command")
	}

	name := fields[0]
	if name == "" || strings.ContainsAny(name, "\r\n\t") {
		return CommandSpec{}, fmt.Errorf("invalid command name")
	}

	args := make([]string, 0, len(fields)-1)
	for _, arg := range fields[1:] {
		if strings.ContainsAny(arg, "\r\n") {
			return CommandSpec{}, fmt.Errorf("invalid command arguments")
		}
		args = append(args, arg)
	}

	return CommandSpec{Name: name, Args: args}, nil
}

func ValidateAppName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("empty app name")
	}
	for _, r := range name {
		if r > unicode.MaxASCII || unicode.IsControl(r) {
			return fmt.Errorf("invalid app name")
		}
	}
	return nil
}

func IsTerminalEditor(editorName string) bool {
	base := strings.ToLower(filepath.Base(strings.Fields(strings.TrimSpace(editorName))[0]))
	terminalEditors := map[string]bool{
		"vim":   true,
		"nvim":  true,
		"nano":  true,
		"vi":    true,
		"emacs": true,
	}
	return terminalEditors[base]
}

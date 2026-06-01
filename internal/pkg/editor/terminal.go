package editor

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

type TerminalLauncher struct {
	Spec CommandSpec
}

func (l *TerminalLauncher) Launch(path string) error {
	if l.Spec.Name == "" {
		return fmt.Errorf("empty editor command")
	}
	switch runtime.GOOS {
	case "darwin":
		return l.launchDarwin(path)
	case "windows":
		args := append([]string{"/c", "start", "cmd", "/k", l.Spec.Name}, l.Spec.Args...)
		args = append(args, path)
		return exec.Command("cmd", args...).Start()
	default:
		return l.launchLinux(path)
	}
}

func (l *TerminalLauncher) launchDarwin(path string) error {
	termProg := os.Getenv("TERM_PROGRAM")

	if termProg == "Ghostty" {
		if _, err := exec.LookPath("ghostty"); err == nil {
			args := append([]string{"+new-tab", "-e", l.Spec.Name}, l.Spec.Args...)
			args = append(args, path)
			return exec.Command("ghostty", args...).Start()
		}
		ghosttyPath := "/Applications/Ghostty.app/Contents/MacOS/ghostty"
		if _, err := os.Stat(ghosttyPath); err == nil {
			args := append([]string{"+new-tab", "-e", l.Spec.Name}, l.Spec.Args...)
			args = append(args, path)
			return exec.Command(ghosttyPath, args...).Start()
		}
	}

	if termProg == "iTerm.app" || termProg == "iTerm" {
		return exec.Command(
			"osascript",
			"-e", iTermScript,
			l.commandLine(path),
		).Start()
	}

	return exec.Command(
		"osascript",
		"-e", terminalScript,
		l.commandLine(path),
	).Start()
}

func (l *TerminalLauncher) launchLinux(path string) error {
	terms := []string{"x-terminal-emulator", "gnome-terminal", "konsole", "alacritty", "kitty", "termite"}
	var term string
	for _, t := range terms {
		if _, err := exec.LookPath(t); err == nil {
			term = t
			break
		}
	}

	if term == "" {
		return fmt.Errorf("no terminal emulator found")
	}

	if term == "gnome-terminal" {
		args := append([]string{"--", l.Spec.Name}, l.Spec.Args...)
		args = append(args, path)
		return exec.Command(term, args...).Start()
	}
	args := append([]string{"-e", l.Spec.Name}, l.Spec.Args...)
	args = append(args, path)
	return exec.Command(term, args...).Start()
}

func (l *TerminalLauncher) commandLine(path string) string {
	parts := append([]string{l.Spec.Name}, l.Spec.Args...)
	parts = append(parts, path)
	quoted := make([]string, 0, len(parts))
	for _, part := range parts {
		quoted = append(quoted, shellQuote(part))
	}
	return strings.Join(quoted, " ")
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", `'"'"'`) + "'"
}

const terminalScript = `on run argv
	set cmdText to item 1 of argv
	tell application "Terminal" to do script cmdText
end run`

const iTermScript = `on run argv
	set cmdText to item 1 of argv
	tell application "iTerm"
		if (count of windows) = 0 then
			create window with default profile
		end if
		tell current window
			create tab with default profile
			tell current session
				write text cmdText
			end tell
		end tell
	end tell
end run`

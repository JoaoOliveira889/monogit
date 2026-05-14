package editor

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

type TerminalLauncher struct {
	Editor string
}

func (l *TerminalLauncher) Launch(path string) error {
	switch runtime.GOOS {
	case "darwin":
		return l.launchDarwin(path)
	case "windows":
		return exec.Command("cmd", "/c", "start", "cmd", "/k", l.Editor, path).Start()
	default:
		return l.launchLinux(path)
	}
}

func (l *TerminalLauncher) launchDarwin(path string) error {
	termProg := os.Getenv("TERM_PROGRAM")
	cmdString := fmt.Sprintf("%s %s", l.Editor, l.escapePath(path))

	if termProg == "Ghostty" {
		if _, err := exec.LookPath("ghostty"); err == nil {
			return exec.Command("ghostty", "+new-tab", "-e", l.Editor, path).Start()
		}
		ghosttyPath := "/Applications/Ghostty.app/Contents/MacOS/ghostty"
		if _, err := os.Stat(ghosttyPath); err == nil {
			return exec.Command(ghosttyPath, "+new-tab", "-e", l.Editor, path).Start()
		}
	}

	if termProg == "iTerm.app" || termProg == "iTerm" {
		script := fmt.Sprintf(`
			tell application "iTerm"
				tell current window
					create tab with default profile
					tell current session
						write text %q
					end tell
				end tell
			end tell
		`, cmdString)
		return exec.Command("osascript", "-e", script).Start()
	}

	script := fmt.Sprintf(`tell application "Terminal" to do script %q`, cmdString)
	return exec.Command("osascript", "-e", script).Start()
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
		return exec.Command(term, "--", l.Editor, path).Start()
	}
	return exec.Command(term, "-e", l.Editor, path).Start()
}

func (l *TerminalLauncher) escapePath(path string) string {
	path = strings.ReplaceAll(path, "\\", "\\\\")
	path = strings.ReplaceAll(path, "\"", "\\\"")
	return "\"" + path + "\""
}

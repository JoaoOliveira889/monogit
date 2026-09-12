package editor

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// TerminalKind identifies the terminal emulator to open new sessions in.
type TerminalKind int

const (
	TerminalUnknown TerminalKind = iota
	TerminalGhostty
	TerminalITerm
	TerminalWezTerm
	TerminalKitty
	TerminalAlacritty
	TerminalApple
)

// DetectTerminal resolves which terminal emulator to use, preferring an explicit
// MONOGIT_TERMINAL over the TERM_PROGRAM the session was started from.
func DetectTerminal() TerminalKind {
	preferred := strings.ToLower(os.Getenv("MONOGIT_TERMINAL"))
	current := strings.ToLower(os.Getenv("TERM_PROGRAM"))

	for _, hint := range []string{preferred, current} {
		switch {
		case hint == "":
			continue
		case strings.Contains(hint, "ghostty"):
			return TerminalGhostty
		case strings.Contains(hint, "iterm"):
			return TerminalITerm
		case strings.Contains(hint, "wezterm"):
			return TerminalWezTerm
		case strings.Contains(hint, "kitty"):
			return TerminalKitty
		case strings.Contains(hint, "alacritty"):
			return TerminalAlacritty
		case strings.Contains(hint, "apple_terminal"), strings.Contains(hint, "terminal"):
			return TerminalApple
		}
	}

	if runtime.GOOS == "darwin" && IsGhosttyInstalled() {
		return TerminalGhostty
	}
	return TerminalUnknown
}

// IsGhosttyInstalled reports whether Ghostty is installed system-wide or for the
// current user.
func IsGhosttyInstalled() bool {
	if _, err := os.Stat("/Applications/Ghostty.app"); err == nil {
		return true
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	_, err = os.Stat(filepath.Join(homeDir, "Applications", "Ghostty.app"))
	return err == nil
}

const appleTerminalCdScript = `on run argv
	set targetPath to item 1 of argv
	set cmdText to "cd " & quoted form of targetPath
	tell application "Terminal" to do script cmdText
end run`

const iTermCdScript = `on run argv
	set targetPath to item 1 of argv
	set cmdText to "cd " & quoted form of targetPath
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

// TerminalAt builds the command that opens a new terminal session rooted at
// path. The path is always passed as an argument, never interpolated into a
// script or shell string.
func TerminalAt(path string) (*exec.Cmd, error) {
	if path == "" {
		return nil, fmt.Errorf("empty terminal path")
	}

	if runtime.GOOS != "darwin" {
		return linuxTerminalAt(path)
	}

	switch DetectTerminal() {
	case TerminalGhostty:
		return exec.Command("open", "-n", "-a", "Ghostty", path), nil
	case TerminalITerm:
		return exec.Command("osascript", "-e", iTermCdScript, path), nil
	case TerminalWezTerm:
		return exec.Command("open", "-a", "WezTerm", path), nil
	case TerminalKitty:
		return exec.Command("open", "-a", "kitty", path), nil
	case TerminalAlacritty:
		return exec.Command("open", "-a", "Alacritty", path), nil
	default:
		return exec.Command("osascript", "-e", appleTerminalCdScript, path), nil
	}
}

func linuxTerminalAt(path string) (*exec.Cmd, error) {
	configured := os.Getenv("MONOGIT_TERMINAL")
	if configured == "" {
		configured = os.Getenv("TERMINAL")
	}
	if configured == "" {
		configured = "xterm"
	}

	spec, err := ParseCommand(configured)
	if err != nil {
		return nil, fmt.Errorf("invalid terminal command %q: %w", configured, err)
	}

	args := append(append([]string{}, spec.Args...), "--working-directory", path)
	return exec.Command(spec.Name, args...), nil
}

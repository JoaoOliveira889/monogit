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
		if _, err := exec.LookPath("wt.exe"); err != nil {
			return fmt.Errorf("Windows Terminal (wt.exe) is required for terminal editors")
		}
		args := []string{"new-tab", "--startingDirectory", path, "--", l.Spec.Name}
		args = append(args, l.Spec.Args...)
		return exec.Command("wt.exe", args...).Start()
	default:
		return l.launchLinux(path)
	}
}

func (l *TerminalLauncher) launchDarwin(path string) error {
	termProg := strings.ToLower(os.Getenv("TERM_PROGRAM"))
	monogitTerm := strings.ToLower(os.Getenv("MONOGIT_TERMINAL"))

	cmdLine := l.commandLine(path)

	// 1. Ghostty (check env vars first, or app existence)
	if strings.Contains(monogitTerm, "ghostty") || strings.Contains(termProg, "ghostty") || isGhosttyInstalled() {
		shCmd := fmt.Sprintf("cd %s && %s", shellQuote(path), l.commandLine(""))
		args := []string{"-a", "Ghostty", "--args", "-e", "sh", "-c", shCmd}
		return exec.Command("open", args...).Start()
	}

	// 2. iTerm / iTerm2
	if strings.Contains(monogitTerm, "iterm") || strings.Contains(termProg, "iterm") {
		return exec.Command(
			"osascript",
			"-e", iTermScript,
			cmdLine,
		).Start()
	}

	// 3. WezTerm
	if strings.Contains(monogitTerm, "wezterm") || strings.Contains(termProg, "wezterm") {
		return exec.Command("open", "-a", "WezTerm", "--args", "start", "--cwd", path, "--", l.Spec.Name).Start()
	}

	// 4. Kitty
	if strings.Contains(monogitTerm, "kitty") || strings.Contains(termProg, "kitty") {
		return exec.Command("open", "-a", "kitty", "--args", "-d", path, l.Spec.Name).Start()
	}

	// 5. Alacritty
	if strings.Contains(monogitTerm, "alacritty") || strings.Contains(termProg, "alacritty") {
		return exec.Command("open", "-a", "Alacritty", "--args", "--working-directory", path, "-e", l.Spec.Name).Start()
	}

	// 6. Default Fallback (Apple Terminal.app)
	return exec.Command(
		"osascript",
		"-e", terminalScript,
		cmdLine,
	).Start()
}

func isGhosttyInstalled() bool {
	if _, err := os.Stat("/Applications/Ghostty.app"); err == nil {
		return true
	}
	return false
}

func (l *TerminalLauncher) launchLinux(path string) error {
	termProg := strings.ToLower(os.Getenv("TERM_PROGRAM"))
	monogitTerm := strings.ToLower(os.Getenv("MONOGIT_TERMINAL"))

	if strings.Contains(monogitTerm, "ghostty") || strings.Contains(termProg, "ghostty") {
		if _, err := exec.LookPath("ghostty"); err == nil {
			args := append([]string{"-e", l.Spec.Name}, l.Spec.Args...)
			args = append(args, path)
			return exec.Command("ghostty", args...).Start()
		}
	}

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
		args := append([]string{"--tab", "--working-directory", path, "--", l.Spec.Name}, l.Spec.Args...)
		return exec.Command(term, args...).Start()
	}
	args := append([]string{"-e", l.Spec.Name}, l.Spec.Args...)
	args = append(args, path)
	return exec.Command(term, args...).Start()
}

func (l *TerminalLauncher) commandLine(path string) string {
	parts := append([]string{l.Spec.Name}, l.Spec.Args...)
	if path != "" {
		parts = append(parts, path)
	}
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
		end if
	end tell
end run`

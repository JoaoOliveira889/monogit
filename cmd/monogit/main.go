package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"monogit/internal/adapters/git"
	"monogit/internal/adapters/tui"
	"monogit/internal/usecase"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	rootPath := flag.String("path", ".", "Root directory to scan for Git repos")
	interval := flag.Duration("interval", 5*time.Minute, "Auto-fetch interval (e.g. 5m, 10m, 1h)")
	showVersion := flag.Bool("version", false, "Show version information")
	flag.Parse()

	if *showVersion {
		fmt.Printf("monogit %s\n", version)
		fmt.Printf("commit: %s\n", commit)
		fmt.Printf("built at: %s\n", date)
		return
	}

	gitAdapter := git.NewGitCLIAdapter()
	gitUseCase := usecase.NewGitUseCase(gitAdapter)

	m := tui.NewModel(*rootPath, *interval, gitUseCase)
	p := tea.NewProgram(&m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "monogit: %v\n", err)
		os.Exit(1)
	}
}

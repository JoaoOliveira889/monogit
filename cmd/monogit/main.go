package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoaoOliveira889/monogit/internal/adapters/git"
	"github.com/JoaoOliveira889/monogit/internal/adapters/tui"
	"github.com/JoaoOliveira889/monogit/internal/pkg/logging"
	"github.com/JoaoOliveira889/monogit/internal/usecase"
)

var (
	version = "0.1.0"
	commit  = "none"
	date    = "unknown"
)

func main() {
	tui.Version = version

	logging.Init()
	defer logging.Close()

	rootPath := flag.String("path", ".", "Root directory to scan for Git repos")
	interval := flag.Duration("interval", 5*time.Minute, "Auto-fetch interval (e.g. 5m, 10m, 1h)")
	showVersion := flag.Bool("version", false, "Show version information")
	flag.Parse()

	if *showVersion {
		fmt.Printf("monogit %s\n", tui.Version)
		fmt.Printf("commit: %s\n", commit)
		fmt.Printf("built at: %s\n", date)
		return
	}

	gitAdapter := git.NewGitCLIAdapter()
	gitUseCase := usecase.NewGitUseCase(gitAdapter)

	m := tui.NewModel(*rootPath, *interval, gitUseCase)
	p := tea.NewProgram(&m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		logging.Error("program exited with error", "error", err)
		fmt.Fprintf(os.Stderr, "monogit: %v\n", err)
		os.Exit(1)
	}
	logging.Info("program exited normally")
}

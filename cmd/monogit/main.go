package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoaoOliveira889/monogit/internal/adapters/git"
	"github.com/JoaoOliveira889/monogit/internal/adapters/tui"
	"github.com/JoaoOliveira889/monogit/internal/pkg/config"
	"github.com/JoaoOliveira889/monogit/internal/pkg/logging"
	"github.com/JoaoOliveira889/monogit/internal/usecase"
)

var (
	version = "0.3.5"
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
	rebaseTodo := flag.String("rebase-todo", "", "Internal: install a prepared rebase todo file (used as GIT_SEQUENCE_EDITOR)")
	flag.Parse()

	if *rebaseTodo != "" {
		if err := installRebaseTodo(*rebaseTodo, flag.Args()); err != nil {
			fmt.Fprintf(os.Stderr, "monogit: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if *showVersion {
		fmt.Printf("monogit %s\n", tui.Version)
		fmt.Printf("commit: %s\n", commit)
		fmt.Printf("built at: %s\n", date)
		return
	}

	cfg := config.LoadConfig()
	git.SetConcurrency(cfg.Concurrency)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	gitAdapter := git.NewGitCLIAdapterWithContext(ctx)
	gitUseCase := usecase.NewGitUseCase(gitAdapter)

	m := tui.NewModel(*rootPath, *interval, gitUseCase)
	p := tea.NewProgram(&m, tea.WithAltScreen(), tea.WithMouseCellMotion())

	if _, err := p.Run(); err != nil {
		logging.Error("program exited with error", "error", err)
		fmt.Fprintf(os.Stderr, "monogit: %v\n", err)
		os.Exit(1)
	}
	logging.Info("program exited normally")
}

// installRebaseTodo copies a prepared todo file over the todo file Git passes as
// the final argument. Git invokes GIT_SEQUENCE_EDITOR through a shell, so this
// indirection keeps every user-controlled value out of that shell command.
func installRebaseTodo(source string, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("rebase-todo: missing destination file")
	}
	destination := args[len(args)-1]

	data, err := os.ReadFile(source)
	if err != nil {
		return fmt.Errorf("rebase-todo: read prepared todo: %w", err)
	}
	if err := os.WriteFile(destination, data, 0600); err != nil {
		return fmt.Errorf("rebase-todo: write todo: %w", err)
	}
	return nil
}

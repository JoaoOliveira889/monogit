package tui

import (
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"monogit/internal/domain"
	"monogit/internal/pkg/scanner"
)

func (m Model) scanReposCmd(rootPath string) tea.Cmd {
	return func() tea.Msg {
		repos, err := scanner.ScanForRepos(rootPath)
		if err != nil {
			return errMsg{err: err}
		}
		return repoScannedMsg{repos: repos}
	}
}

func (m Model) refreshStatusCmd(index int, path string) tea.Cmd {
	return func() tea.Msg {
		repo, err := m.gitUC.GetRepositoryStatus(path)
		if err != nil {
			return repoStatusMsg{index: index, err: err}
		}

		return repoStatusMsg{
			index:  index,
			branch: repo.Branch,
			ahead:  repo.Ahead,
			behind: repo.Behind,
			dirty:  repo.IsDirty,
		}
	}
}

func (m Model) fetchRepoCmd(index int, path string) tea.Cmd {
	return func() tea.Msg {
		err := m.gitUC.Fetch(path)
		return fetchDoneMsg{index: index, err: err}
	}
}

func (m Model) fetchAllCmd(repos []domain.Repository) tea.Cmd {
	return func() tea.Msg {
		var wg sync.WaitGroup
		for i := range repos {
			wg.Add(1)
			go func(idx int, path string) {
				defer wg.Done()
				_ = m.gitUC.Fetch(path)
			}(i, repos[i].Path)
		}
		wg.Wait()
		return fetchAllDoneMsg{}
	}
}

func (m Model) pullRepoCmd(index int, path string) tea.Cmd {
	return func() tea.Msg {
		output, err := m.gitUC.Pull(path)
		return pullDoneMsg{index: index, output: output, err: err}
	}
}

func (m Model) pullAllCmd(repos []domain.Repository) tea.Cmd {
	return func() tea.Msg {
		var (
			wg      sync.WaitGroup
			mu      sync.Mutex
			results []pullResult
		)

		for i, r := range repos {
			if r.IsDirty {
				results = append(results, pullResult{
					index:  i,
					name:   r.Name,
					output: "Skipped: repository has local changes (dirty)",
					err:    nil,
				})
				continue
			}
			wg.Add(1)
			go func(idx int, repo domain.Repository) {
				defer wg.Done()
				output, err := m.gitUC.Pull(repo.Path)
				mu.Lock()
				results = append(results, pullResult{
					index:  idx,
					name:   repo.Name,
					output: output,
					err:    err,
				})
				mu.Unlock()
			}(i, r)
		}

		wg.Wait()
		return pullAllDoneMsg{results: results}
	}
}

func (m Model) pushAllCmd(repos []domain.Repository) tea.Cmd {
	return func() tea.Msg {
		var (
			wg      sync.WaitGroup
			mu      sync.Mutex
			results []pushResult
		)

		for i, r := range repos {
			if r.Ahead == 0 {
				continue
			}
			wg.Add(1)
			go func(idx int, repo domain.Repository) {
				defer wg.Done()
				output, err := m.gitUC.Push(repo.Path)
				mu.Lock()
				results = append(results, pushResult{
					index:  idx,
					name:   repo.Name,
					output: output,
					err:    err,
				})
				mu.Unlock()
			}(i, r)
		}

		wg.Wait()
		return pushAllDoneMsg{results: results}
	}
}

func (m Model) commitCmd(index int, path string, message string) tea.Cmd {
	return func() tea.Msg {
		output, err := m.gitUC.Commit(path, message)
		return commitDoneMsg{index: index, output: output, err: err}
	}
}

func tickCmd(interval time.Duration) tea.Cmd {
	return tea.Tick(interval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func spinnerTickCmd() tea.Cmd {
	return tea.Tick(80*time.Millisecond, func(t time.Time) tea.Msg {
		return spinnerTickMsg{}
	})
}

func spinnerTickMsgCmd() tea.Cmd {
	return tea.Tick(80*time.Millisecond, func(t time.Time) tea.Msg {
		return spinnerTickMsg{}
	})
}

func (m Model) refreshAllStatusCmd(repos []domain.Repository) tea.Cmd {
	cmds := make([]tea.Cmd, len(repos))
	for i, r := range repos {
		cmds[i] = m.refreshStatusCmd(i, r.Path)
	}
	return tea.Batch(cmds...)
}

func (m Model) fetchFilesCmd(repoPath string) tea.Cmd {
	return func() tea.Msg {
		files, err := m.gitUC.GetFiles(repoPath)
		if err != nil {
			return errMsg{err}
		}
		return gitFilesMsg{files}
	}
}

func (m Model) fetchBranchesCmd(repoPath string) tea.Cmd {
	return func() tea.Msg {
		branches, err := m.gitUC.GetBranches(repoPath)
		if err != nil {
			return errMsg{err}
		}
		return gitBranchesMsg{branches}
	}
}

func (m Model) toggleFileCmd(repoPath string, f domain.FileStatus) tea.Cmd {
	return func() tea.Msg {
		_ = m.gitUC.ToggleFile(repoPath, f)
		files, _ := m.gitUC.GetFiles(repoPath)
		return gitFilesMsg{files}
	}
}

func (m Model) unstageAllCmd(repoPath string) tea.Cmd {
	return func() tea.Msg {
		_ = m.gitUC.UnstageAll(repoPath)
		files, _ := m.gitUC.GetFiles(repoPath)
		return gitFilesMsg{files}
	}
}

func (m Model) stashCmd(index int, repoPath string) tea.Cmd {
	return func() tea.Msg {
		out, err := m.gitUC.Stash(repoPath, "MonoGit Stash")
		return stashDoneMsg{index, out, err}
	}
}

func (m Model) stashPopCmd(index int, repoPath string) tea.Cmd {
	return func() tea.Msg {
		out, err := m.gitUC.StashPop(repoPath)
		return stashPopDoneMsg{index, out, err}
	}
}

func (m Model) fetchDiffCmd(repoPath string, f domain.FileStatus) tea.Cmd {
	return func() tea.Msg {
		diff, _ := m.gitUC.GetDiff(repoPath, f)
		return gitDiffMsg{diff}
	}
}

func (m Model) pushCmd(index int, repoPath string) tea.Cmd {
	return func() tea.Msg {
		out, err := m.gitUC.Push(repoPath)
		return pushDoneMsg{index, out, err}
	}
}

func (m Model) addAllCmd(repoPath string) tea.Cmd {
	return func() tea.Msg {
		_ = m.gitUC.AddAll(repoPath)
		files, _ := m.gitUC.GetFiles(repoPath)
		return gitFilesMsg{files}
	}
}

func (m Model) undoCommitCmd(repoPath string) tea.Cmd {
	return func() tea.Msg {
		err := m.gitUC.UndoCommit(repoPath)
		if err != nil {
			return errMsg{err}
		}
		return refreshMsg{}
	}
}

func (m Model) stageByPatternCmd(repoPath string, pattern string) tea.Cmd {
	return func() tea.Msg {
		err := m.gitUC.StageByPattern(repoPath, pattern)
		if err != nil {
			return errMsg{err}
		}
		files, _ := m.gitUC.GetFiles(repoPath)
		return gitFilesMsg{files}
	}
}

func (m Model) discardChangesCmd(repoPath string, f domain.FileStatus) tea.Cmd {
	return func() tea.Msg {
		err := m.gitUC.DiscardFile(repoPath, f)
		if err != nil {
			return errMsg{err}
		}
		files, _ := m.gitUC.GetFiles(repoPath)
		return gitFilesMsg{files}
	}
}

func (m Model) checkoutBranchCmd(index int, repoPath string, branch string) tea.Cmd {
	return func() tea.Msg {
		err := m.gitUC.CheckoutBranch(repoPath, branch)
		if err != nil {
			return errMsg{err}
		}
		return repoStatusMsg{index: index, refresh: true}
	}
}

func (m Model) createBranchCmd(repoPath string, branch string) tea.Cmd {
	return func() tea.Msg {
		err := m.gitUC.CreateBranch(repoPath, branch)
		if err != nil {
			return errMsg{err}
		}
		return refreshMsg{}
	}
}

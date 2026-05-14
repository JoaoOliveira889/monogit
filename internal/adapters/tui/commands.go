package tui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"monogit/internal/domain"
	"monogit/internal/pkg/editor"
	"monogit/internal/pkg/scanner"
)

func (m Model) scanReposCmd(rootPath string) tea.Cmd {
	return func() tea.Msg {
		repos, err := scanner.ScanForRepos(rootPath)
		if err != nil {
			return errMsg{Err: err}
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

func (m Model) fetchAllCmd() tea.Cmd {
	return func() tea.Msg {
		var wg sync.WaitGroup
		sem := make(chan struct{}, 10)
		for i := range m.repos {
			wg.Add(1)
			go func(path string) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()
				_ = m.gitUC.Fetch(path)
			}(m.repos[i].Path)
		}
		wg.Wait()
		return fetchDoneMsg{all: true}
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
			results []PullResult
		)

		sem := make(chan struct{}, 5)
		for i, r := range repos {
			if r.IsDirty {
				results = append(results, PullResult{
					Index:  i,
					Name:   r.Name,
					Output: "Skipped: repository has local changes (dirty)",
				})
				continue
			}
			wg.Add(1)
			go func(idx int, repo domain.Repository) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()
				output, err := m.gitUC.Pull(repo.Path)
				mu.Lock()
				results = append(results, PullResult{
					Index:  idx,
					Name:   repo.Name,
					Output: output,
					Err:    err,
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
			results []PushResult
		)

		sem := make(chan struct{}, 5)
		for i, r := range repos {
			if r.Ahead == 0 {
				continue
			}
			wg.Add(1)
			go func(idx int, repo domain.Repository) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()
				output, err := m.gitUC.Push(repo.Path)
				mu.Lock()
				results = append(results, PushResult{
					Index:  idx,
					Name:   repo.Name,
					Output: output,
					Err:    err,
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
			return errMsg{Err: err}
		}
		return gitFilesMsg{files}
	}
}

func (m Model) fetchBranchesCmd(repoPath string) tea.Cmd {
	return func() tea.Msg {
		branches, err := m.gitUC.GetBranches(repoPath)
		if err != nil {
			return errMsg{Err: err}
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

func (m Model) addAllAndNextCmd(repoPath string) tea.Cmd {
	return func() tea.Msg {
		_ = m.gitUC.AddAll(repoPath)
		return nextStepMsg{}
	}
}

func (m Model) undoCommitCmd(repoPath string) tea.Cmd {
	return func() tea.Msg {
		err := m.gitUC.UndoCommit(repoPath)
		if err != nil {
			return errMsg{Err: err}
		}
		return refreshMsg{}
	}
}

func (m Model) stageByPatternCmd(repoPath string, pattern string) tea.Cmd {
	return func() tea.Msg {
		err := m.gitUC.StageByPattern(repoPath, pattern)
		if err != nil {
			return errMsg{Err: err}
		}
		files, _ := m.gitUC.GetFiles(repoPath)
		return gitFilesMsg{files}
	}
}

func (m Model) discardChangesCmd(repoPath string, f domain.FileStatus) tea.Cmd {
	return func() tea.Msg {
		err := m.gitUC.DiscardFile(repoPath, f)
		if err != nil {
			return errMsg{Err: err}
		}
		files, _ := m.gitUC.GetFiles(repoPath)
		return gitFilesMsg{files}
	}
}

func (m Model) checkoutBranchCmd(index int, repoPath string, branch string) tea.Cmd {
	return func() tea.Msg {
		err := m.gitUC.CheckoutBranch(repoPath, branch)
		return checkoutBranchDoneMsg{index: index, err: err}
	}
}

func (m Model) createBranchCmd(repoPath string, branch string) tea.Cmd {
	return func() tea.Msg {
		err := m.gitUC.CreateBranch(repoPath, branch)
		if err != nil {
			return errMsg{Err: err}
		}
		return refreshMsg{}
	}
}

func (m Model) deleteBranchCmd(index int, repoPath string, branch string) tea.Cmd {
	return func() tea.Msg {
		out, err := m.gitUC.DeleteBranch(repoPath, branch)
		return deleteBranchDoneMsg{index: index, output: out, err: err}
	}
}

func (m Model) deleteRemoteBranchCmd(index int, repoPath string, branch string) tea.Cmd {
	return func() tea.Msg {
		out, err := m.gitUC.DeleteRemoteBranch(repoPath, "origin", branch)
		return deleteRemoteBranchDoneMsg{index: index, output: out, err: err}
	}
}

func clearStatusCmd(id int) tea.Cmd {
	return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
		return clearStatusMsg{id: id}
	})
}

func (m Model) openEditorCmd(repoPath string, editorName string) tea.Cmd {
	return func() tea.Msg {
		if editorName == "" {
			return openEditorMsg{err: fmt.Errorf("no editor specified")}
		}

		launcher := editor.NewLauncher(editorName)
		err := launcher.Launch(repoPath)
		if err != nil {
			return openEditorMsg{err: fmt.Errorf("failed to start editor %s: %w", editorName, err)}
		}

		return openEditorMsg{editor: editorName}
	}
}

func (m Model) openInBrowserCmd(repoPath string) tea.Cmd {
	return func() tea.Msg {
		url, err := m.gitUC.GetRemoteURL(repoPath)
		if err != nil {
			return openBrowserMsg{err: err}
		}

		var cmd *exec.Cmd
		switch runtime.GOOS {
		case "darwin":
			cmd = exec.Command("open", url)
		case "windows":
			cmd = exec.Command("cmd", "/c", "start", url)
		default:
			cmd = exec.Command("xdg-open", url)
		}

		err = cmd.Run()
		if err != nil {
			return openBrowserMsg{err: fmt.Errorf("failed to open browser: %w", err)}
		}

		return openBrowserMsg{url: url}
	}
}

func (m Model) scanEditorsCmd() tea.Cmd {
	return func() tea.Msg {
		var detected []string

		for _, env := range []string{"MONOGIT_EDITOR", "VISUAL", "EDITOR"} {
			if e := os.Getenv(env); e != "" {
				detected = append(detected, e)
			}
		}

		if out, err := exec.Command("git", "config", "--get", "core.editor").Output(); err == nil {
			detected = append(detected, strings.TrimSpace(string(out)))
		}

		switch runtime.GOOS {
		case "darwin":
			queries := []string{
				"kMDItemAppStoreCategoryType == 'public.app-category.developer-tools'",
				"kMDItemContentTypeTree == 'public.source-code'",
			}
			filterKeywords := []string{"Code", "Editor", "Studio", "IDE", "Rider", "Text", "Vim", "Sublime", "Zed", "Cursor", "Antigravity", "Atom"}
			for _, query := range queries {
				if out, err := exec.Command("mdfind", query).Output(); err == nil {
					paths := strings.Split(string(out), "\n")
					for _, path := range paths {
						path = strings.TrimSpace(path)
						if path != "" && strings.HasSuffix(path, ".app") {
							name := strings.TrimSuffix(filepath.Base(path), ".app")
							keep := false
							for _, kw := range filterKeywords {
								if strings.Contains(strings.ToLower(name), strings.ToLower(kw)) {
									keep = true
									break
								}
							}
							if keep {
								detected = append(detected, name+" (App)")
							}
						}
					}
				}
			}
		case "linux":
			searchDirs := []string{"/usr/share/applications", os.Getenv("HOME") + "/.local/share/applications"}
			for _, dir := range searchDirs {
				files, err := os.ReadDir(dir)
				if err != nil {
					continue
				}
				for _, file := range files {
					if strings.HasSuffix(file.Name(), ".desktop") {
						content, err := os.ReadFile(filepath.Join(dir, file.Name()))
						if err != nil {
							continue
						}
						sContent := string(content)
						if strings.Contains(sContent, "Categories=") &&
							(strings.Contains(sContent, "Development") || strings.Contains(sContent, "IDE")) {
							// Extract Name
							for _, line := range strings.Split(sContent, "\n") {
								if strings.HasPrefix(line, "Name=") {
									name := strings.TrimPrefix(line, "Name=")
									detected = append(detected, name)
									break
								}
							}
						}
					}
				}
			}
		}

		cliTools := []string{"code", "vim", "nvim", "nano", "vi", "emacs", "rider", "subl"}
		for _, tool := range cliTools {
			if _, err := exec.LookPath(tool); err == nil {
				detected = append(detected, tool)
			}
		}

		seen := make(map[string]bool)
		unique := make([]string, 0, len(detected))
		for _, d := range detected {
			if d == "" {
				continue
			}
			if !seen[d] {
				seen[d] = true
				unique = append(unique, d)
			}
		}

		return editorsDetectedMsg{editors: unique}
	}
}


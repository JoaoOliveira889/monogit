package tui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoaoOliveira889/monogit/internal/domain"
	"github.com/JoaoOliveira889/monogit/internal/pkg/config"
	"github.com/JoaoOliveira889/monogit/internal/pkg/editor"
	"github.com/JoaoOliveira889/monogit/internal/pkg/scanner"
)

func (m Model) scanReposCmd(rootPath string) tea.Cmd {
	repoTags := m.cfg.RepoTags
	scanExcludes := append([]string(nil), m.cfg.ScanExcludes...)
	return func() tea.Msg {
		repos, err := scanner.ScanForRepos(rootPath, repoTags, scanExcludes)
		if err != nil {
			return errMsg{Err: err}
		}
		return repoScannedMsg{repos: repos}
	}
}

func (m Model) loadStartupReposCmd(rootPath string) tea.Cmd {
	repoTags := m.cfg.RepoTags
	return func() tea.Msg {
		repos, err := config.LoadStartupRepos(rootPath, repoTags)
		if err != nil || len(repos) == 0 {
			return nil
		}
		return startupReposMsg{repos: repos}
	}
}

func (m Model) saveStartupReposCmd(rootPath string, repos []domain.Repository) tea.Cmd {
	return func() tea.Msg {
		_ = config.SaveStartupRepos(rootPath, repos)
		return nil
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

func (m Model) refreshCachedRepoDetailCmd(index int, path string) tea.Cmd {
	return func() tea.Msg {
		snapshot, err := m.gitUC.GetRepositorySnapshot(path, m.viewGraph, 30)
		if err != nil {
			return repoDetailMsg{index: index, path: path, err: err, graph: m.viewGraph}
		}

		return repoDetailMsg{
			index:          index,
			path:           path,
			branch:         snapshot.Branch,
			ahead:          snapshot.Ahead,
			behind:         snapshot.Behind,
			dirty:          snapshot.IsDirty,
			detached:       snapshot.IsDetached,
			hasUpstream:    snapshot.HasUpstream,
			hasConflicts:   snapshot.HasConflicts,
			isStale:        snapshot.IsStale,
			hasUnpushedTag: snapshot.HasUnpushedTag,
			modified:       snapshot.ModifiedCount,
			untracked:      snapshot.UntrackedCount,
			lastCommit:     snapshot.LastCommit,
			log:            snapshot.Log,
			graph:          m.viewGraph,
		}
	}
}

func (m Model) fetchRepoCmd(index int, path string) tea.Cmd {
	return func() tea.Msg {
		err := m.gitUC.Fetch(path)
		return fetchDoneMsg{index: index, output: "Fetched remote updates", err: err}
	}
}

func (m Model) fetchAllCmd() tea.Cmd {
	return func() tea.Msg {
		var wg sync.WaitGroup
		sem := make(chan struct{}, 10)
		var (
			firstErr error
			results  []fetchAllResult
		)
		var mu sync.Mutex
		for i := range m.repos {
			wg.Add(1)
			go func(idx int, name, path string) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()
				err := m.gitUC.Fetch(path)
				mu.Lock()
				results = append(results, fetchAllResult{
					Index:  idx,
					Name:   name,
					Output: "Fetched remote updates",
					Err:    err,
				})
				if err != nil {
					if firstErr == nil {
						firstErr = err
					}
				}
				mu.Unlock()
			}(i, m.repos[i].Name, m.repos[i].Path)
		}
		wg.Wait()
		if len(results) > 1 {
			sort.Slice(results, func(i, j int) bool {
				return results[i].Index < results[j].Index
			})
		}
		return fetchAllDoneMsg{results: results, err: firstErr}
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

func (m Model) checkoutAllCmd(branch string) tea.Cmd {
	filtered := m.filteredRepos()
	include := make(map[string]bool, len(filtered))
	for _, r := range filtered {
		include[r.Path] = true
	}
	repos := m.repos
	return func() tea.Msg {
		var (
			wg      sync.WaitGroup
			mu      sync.Mutex
			results []BulkCheckoutResult
		)
		sem := make(chan struct{}, 5)
		for i, r := range repos {
			if !include[r.Path] {
				results = append(results, BulkCheckoutResult{
					Index:  i,
					Name:   r.Name,
					Branch: branch,
					Err:    fmt.Errorf("skipped: not in current filter"),
				})
				continue
			}
			wg.Add(1)
			go func(idx int, name, path string) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()
				err := m.gitUC.CheckoutBranch(path, branch)
				mu.Lock()
				results = append(results, BulkCheckoutResult{
					Index:  idx,
					Name:   name,
					Branch: branch,
					Err:    err,
				})
				mu.Unlock()
			}(i, r.Name, r.Path)
		}
		wg.Wait()
		sort.Slice(results, func(i, j int) bool {
			return results[i].Index < results[j].Index
		})
		return checkoutAllDoneMsg{results: results}
	}
}

func (m Model) stashAllCmd() tea.Cmd {
	filtered := m.filteredRepos()
	include := make(map[string]bool, len(filtered))
	for _, r := range filtered {
		include[r.Path] = true
	}
	repos := m.repos
	return func() tea.Msg {
		var (
			wg      sync.WaitGroup
			mu      sync.Mutex
			results []BulkStashResult
		)
		sem := make(chan struct{}, 5)
		for i, r := range repos {
			if !include[r.Path] {
				results = append(results, BulkStashResult{
					Index:  i,
					Name:   r.Name,
					Output: "Skipped: not in current filter",
				})
				continue
			}
			if !r.IsDirty {
				results = append(results, BulkStashResult{
					Index:  i,
					Name:   r.Name,
					Output: "Skipped: repository is clean (no local changes)",
				})
				continue
			}
			wg.Add(1)
			go func(idx int, name, path string) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()
				output, err := m.gitUC.Stash(path, "MonoGit Stash")
				mu.Lock()
				results = append(results, BulkStashResult{
					Index:  idx,
					Name:   name,
					Output: output,
					Err:    err,
				})
				mu.Unlock()
			}(i, r.Name, r.Path)
		}
		wg.Wait()
		sort.Slice(results, func(i, j int) bool {
			return results[i].Index < results[j].Index
		})
		return stashAllDoneMsg{results: results}
	}
}

func (m Model) commitAllCmd(index int, path string, message string) tea.Cmd {
	return func() tea.Msg {
		output, err := m.gitUC.CommitAll(path, message)
		return commitDoneMsg{index: index, output: output, err: err}
	}
}

func (m Model) commitSelectedCmd(index int, path string, files []string, message string) tea.Cmd {
	return func() tea.Msg {
		output, err := m.gitUC.CommitSelected(path, files, message)
		return commitDoneMsg{index: index, output: output, err: err}
	}
}

func tickCmd(interval time.Duration) tea.Cmd {
	return tea.Tick(interval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

const spinnerTickInterval = 80 * time.Millisecond
const splashTickInterval = 90 * time.Millisecond

func spinnerTickCmd() tea.Cmd {
	return tea.Tick(spinnerTickInterval, func(t time.Time) tea.Msg {
		return spinnerTickMsg{}
	})
}

func splashTickCmd() tea.Cmd {
	return tea.Tick(splashTickInterval, func(t time.Time) tea.Msg {
		return splashTickMsg{}
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
		if err := m.gitUC.ToggleFile(repoPath, f); err != nil {
			return errMsg{Err: err}
		}
		files, _ := m.gitUC.GetFiles(repoPath)
		return gitFilesMsg{files}
	}
}

func (m Model) unstageAllCmd(repoPath string) tea.Cmd {
	return func() tea.Msg {
		if err := m.gitUC.UnstageAll(repoPath); err != nil {
			return errMsg{Err: err}
		}
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

func (m Model) fetchStashesCmd(repoPath string) tea.Cmd {
	return func() tea.Msg {
		stashes, err := m.gitUC.GetStashes(repoPath)
		if err != nil {
			return errMsg{Err: err}
		}
		return gitStashesMsg{stashes}
	}
}

func (m Model) stashApplyCmd(index int, repoPath string, stashIndex int) tea.Cmd {
	return func() tea.Msg {
		out, err := m.gitUC.ApplyStash(repoPath, stashIndex)
		return stashApplyDoneMsg{index, out, err}
	}
}

func (m Model) stashDropCmd(index int, repoPath string, stashIndex int) tea.Cmd {
	return func() tea.Msg {
		out, err := m.gitUC.DropStash(repoPath, stashIndex)
		return stashDropDoneMsg{index, out, err}
	}
}

func (m Model) fetchStashFilesCmd(repoPath string, stashIndex int) tea.Cmd {
	return func() tea.Msg {
		files, err := m.gitUC.GetStashFiles(repoPath, stashIndex)
		if err != nil {
			return stashFilesMsg{files: []string{fmt.Sprintf("Error: %s", err)}}
		}
		if files == nil {
			files = []string{}
		}
		return stashFilesMsg{files: files}
	}
}

func (m Model) stashPopIndexCmd(index int, repoPath string, stashIndex int) tea.Cmd {
	return func() tea.Msg {
		out, err := m.gitUC.PopStash(repoPath, stashIndex)
		return stashPopIndexDoneMsg{index, out, err}
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
		if err := m.gitUC.AddAll(repoPath); err != nil {
			return errMsg{Err: err}
		}
		files, _ := m.gitUC.GetFiles(repoPath)
		return gitFilesMsg{files}
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

func (m Model) mergeCmd(index int, repoPath string, branch string) tea.Cmd {
	return func() tea.Msg {
		output, err := m.gitUC.Merge(repoPath, branch)
		return mergeDoneMsg{index: index, branch: branch, output: output, err: err}
	}
}

func (m Model) createAndPushTagCmd(index int, repoPath string, name string, message string) tea.Cmd {
	return func() tea.Msg {
		out, err := m.gitUC.CreateAndPushTag(repoPath, name, message)
		return tagDoneMsg{index: index, output: out, err: err}
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
		if strings.HasSuffix(editorName, "(App)") {
			appName := strings.TrimSpace(strings.TrimSuffix(editorName, "(App)"))
			if err := editor.ValidateAppName(appName); err != nil {
				return openEditorMsg{err: fmt.Errorf("invalid editor app: %w", err)}
			}
		} else if _, err := editor.ParseCommand(editorName); err != nil {
			return openEditorMsg{err: fmt.Errorf("invalid editor command: %w", err)}
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

func (m Model) fetchConflictFilesCmd(repoPath string) tea.Cmd {
	return func() tea.Msg {
		hasConflicts, err := m.gitUC.HasConflicts(repoPath)
		if err != nil || !hasConflicts {
			return conflictFilesMsg{files: nil}
		}
		files, err := m.gitUC.ListConflictingFiles(repoPath)
		if err != nil {
			return errMsg{Err: err}
		}
		return conflictFilesMsg{files: files}
	}
}

func (m Model) fetchCompactDiffCmd(repoPath string, f domain.FileStatus) tea.Cmd {
	return func() tea.Msg {
		changes, err := m.gitUC.GetCompactDiff(repoPath, f)
		if err != nil {
			return errMsg{Err: err}
		}
		return compactDiffMsg{changes: changes}
	}
}

func (m Model) openMergetoolCmd(index int, repoPath string, tool string, file string) tea.Cmd {
	spec, err := m.gitUC.OpenMergetool(repoPath, tool, file)
	if err != nil {
		return func() tea.Msg {
			return mergetoolDoneMsg{index: index, path: repoPath, file: file, err: err}
		}
	}

	cmd := exec.Command(spec.Name, spec.Args...)
	cmd.Dir = spec.Dir

	return tea.ExecProcess(cmd, func(execErr error) tea.Msg {
		return mergetoolDoneMsg{
			index: index,
			path:  repoPath,
			file:  file,
			err:   execErr,
		}
	})
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

		return editorsDetectedMsg{editors: dedupeStrings(detected)}
	}
}

func dedupeStrings(values []string) []string {
	seen := make(map[string]bool, len(values))
	unique := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		unique = append(unique, value)
	}
	return unique
}

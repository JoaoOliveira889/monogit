package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) handleTick() (tea.Model, tea.Cmd) {
	if len(m.repos) > 0 {
		m.statusMsg = "⟳ Auto-fetching..."
		for i := range m.repos {
			m.repos[i].Fetching = true
		}
		return m, tea.Batch(m.fetchAllCmd(), tickCmd(m.fetchInterval))
	}
	return m, tickCmd(m.fetchInterval)
}

func (m *Model) handleStartupRepos(msg startupReposMsg) (tea.Model, tea.Cmd) {
	if !m.scanning || len(msg.repos) == 0 {
		return m, nil
	}

	m.repos = msg.repos
	m.splashReady = true
	m.maybeHideSplash()
	m.refreshViewports()
	return m, nil
}

func (m *Model) handleRepoScanned(msg repoScannedMsg) (tea.Model, tea.Cmd) {
	m.repos = msg.repos
	m.scanning = false
	m.splashReady = true
	m.maybeHideSplash()
	if len(m.repos) == 0 {
		m.cachedModifiedCount = 0
		m.cachedUntrackedCount = 0
		m.cachedLastCommit = ""
		m.cachedLog = ""
		m.cachedDetailFor = ""
		m.cachedLogFor = ""
		m.detailLoading = false
		m.statusMsg = "No Git repositories found."
		return m, tea.Batch(m.saveStartupReposCmd(m.rootPath, m.repos), tickCmd(m.fetchInterval))
	}
	m.refreshViewports()
	return m, tea.Batch(
		m.refreshAllStatusCmd(m.repos),
		m.refreshSelectedRepoDetailCmd(),
		m.saveStartupReposCmd(m.rootPath, m.repos),
		tickCmd(m.fetchInterval),
	)
}

func (m *Model) maybeHideSplash() {
	if !m.showSplash || !m.splashReady {
		return
	}
	if time.Since(m.splashStartedAt) >= splashMinDuration {
		m.showSplash = false
	}
}

func (m *Model) handleRepoStatus(msg repoStatusMsg) (tea.Model, tea.Cmd) {
	if msg.index >= 0 && msg.index < len(m.repos) {
		r := &m.repos[msg.index]
		r.CheckingOut = false

		if msg.refresh {
			return m, tea.Batch(m.refreshStatusCmd(msg.index, r.Path), m.refreshCachedRepoDetailCmd(msg.index, r.Path))
		}

		if msg.err != nil {
			r.Error = msg.err.Error()
		} else {
			r.Branch = msg.branch
			r.Ahead = msg.ahead
			r.Behind = msg.behind
			r.IsDirty = msg.dirty
			r.Error = ""
		}
	}
	m.refreshViewports()
	if r := m.selectedRepo(); r != nil && msg.index == m.cursor && msg.err == nil {
		return m, m.refreshCachedRepoDetailCmd(msg.index, r.Path)
	}
	return m, nil
}

func (m *Model) refreshSelectedRepoDetailCmd() tea.Cmd {
	r := m.selectedRepo()
	if r == nil {
		m.cachedModifiedCount = 0
		m.cachedUntrackedCount = 0
		m.cachedLastCommit = ""
		m.cachedLog = ""
		m.cachedDetailFor = ""
		m.cachedLogFor = ""
		m.detailLoading = false
		return nil
	}
	m.detailLoading = true
	return m.refreshCachedRepoDetailCmd(m.cursor, r.Path)
}

func (m *Model) handleRepoDetail(msg repoDetailMsg) (tea.Model, tea.Cmd) {
	r := m.selectedRepo()
	if r == nil || r.Path != msg.path || msg.graph != m.viewGraph {
		return m, nil
	}

	m.detailLoading = false
	if msg.err != nil {
		m.statusMsg = "Failed to refresh repo details (see log 'o')"
		m.appendCommandLog(CommandLogEntry{
			Time:     time.Now(),
			RepoName: r.Name,
			Command:  "refresh repo detail",
			Output:   msg.path,
			Error:    msg.err,
		})
		return m, nil
	}

	m.cachedModifiedCount = msg.modified
	m.cachedUntrackedCount = msg.untracked
	m.cachedLastCommit = msg.lastCommit
	m.cachedLog = msg.log
	m.cachedDetailFor = msg.path
	m.cachedLogFor = msg.path
	m.cachedLogGraph = msg.graph
	m.refreshViewports()
	return m, nil
}

func (m *Model) handleFetchDone(msg tea.Msg) (tea.Model, tea.Cmd) {
	fetchMsg, ok := msg.(fetchDoneMsg)
	if !ok {
		return m, nil
	}

	if fetchMsg.index >= 0 && fetchMsg.index < len(m.repos) {
		r := &m.repos[fetchMsg.index]
		r.Fetching = false
		m.appendCommandLog(CommandLogEntry{
			Time:     time.Now(),
			RepoName: r.Name,
			Command:  "fetch",
			Output:   fetchMsg.output,
			Error:    fetchMsg.err,
		})
		if fetchMsg.err != nil {
			m.statusMsg = fmt.Sprintf("Fetch failed for %s (see log 'o')", r.Name)
		} else {
			m.statusMsg = "Fetch complete"
		}
		return m, tea.Batch(m.refreshStatusCmd(fetchMsg.index, r.Path), m.refreshCachedRepoDetailCmd(fetchMsg.index, r.Path))
	}
	return m, nil
}

func (m *Model) handleFetchAllDone(msg fetchAllDoneMsg) (tea.Model, tea.Cmd) {
	for i := range m.repos {
		m.repos[i].Fetching = false
	}

	if len(msg.results) > 0 {
		for _, res := range msg.results {
			m.appendCommandLog(CommandLogEntry{
				Time:     time.Now(),
				RepoName: res.Name,
				Command:  "fetch",
				Output:   res.Output,
				Error:    res.Err,
			})
		}
	}

	if msg.err != nil {
		m.statusMsg = "Fetch all finished with errors (see log 'o')"
	} else {
		m.statusMsg = "Fetch all complete"
	}

	return m, tea.Batch(m.refreshAllStatusCmd(m.repos), m.refreshSelectedRepoDetailCmd())
}

func (m *Model) handlePullDone(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case pullDoneMsg:
		if msg.index >= 0 && msg.index < len(m.repos) {
			r := &m.repos[msg.index]
			r.Pulling = false

			m.appendCommandLog(CommandLogEntry{
				Time:     time.Now(),
				RepoName: r.Name,
				Command:  "pull",
				Output:   msg.output,
				Error:    msg.err,
			})

			if msg.err != nil {
				m.statusMsg = fmt.Sprintf("Pull failed for %s (see log 'o')", r.Name)
			} else {
				m.statusMsg = "Pull complete"
			}
			return m, tea.Batch(m.refreshStatusCmd(msg.index, r.Path), m.refreshCachedRepoDetailCmd(msg.index, r.Path))
		}
	case pullAllDoneMsg:
		for i := range m.repos {
			m.repos[i].Pulling = false
		}

		var failedCount int
		for _, res := range msg.results {
			m.appendCommandLog(CommandLogEntry{
				Time:     time.Now(),
				RepoName: res.Name,
				Command:  "pull",
				Output:   res.Output,
				Error:    res.Err,
			})
			if res.Err != nil {
				failedCount++
			}
		}

		if failedCount > 0 {
			m.statusMsg = fmt.Sprintf("Pull all finished with %d errors (see log 'o')", failedCount)
		} else {
			m.statusMsg = "Pull all complete"
		}
		return m, tea.Batch(m.refreshAllStatusCmd(m.repos), m.refreshSelectedRepoDetailCmd())
	}
	return m, nil
}

func (m *Model) handleCommitDone(msg commitDoneMsg) (tea.Model, tea.Cmd) {
	if msg.index >= 0 && msg.index < len(m.repos) {
		r := &m.repos[msg.index]
		r.Committing = false
		m.appendCommandLog(CommandLogEntry{
			Time:     time.Now(),
			RepoName: r.Name,
			Command:  "commit",
			Output:   msg.output,
			Error:    msg.err,
		})
		m.activePanel = RepoPanel
		m.commitStep = StepAddOption
		m.commitMode = CommitModeAll
		if msg.err == nil {
			m.showConfirmModal = true
			m.confirmModalTitle = "Commit successful! Push now?"
			m.confirmModalAction = "push"
		}
		return m, tea.Batch(m.refreshStatusCmd(msg.index, r.Path), m.refreshCachedRepoDetailCmd(msg.index, r.Path))
	}
	return m, nil
}

func (m *Model) handleGitFiles(msg gitFilesMsg) (tea.Model, tea.Cmd) {
	m.files = msg.files
	m.fileSelections = make(map[int]bool)
	for i, f := range m.files {
		if f.Staged {
			m.fileSelections[i] = true
		}
	}
	m.showFiles = true
	m.statusMsg = ""
	if m.activePanel != DiffPanel {
		m.activePanel = LogPanel
	}
	if len(m.files) > 0 {
		if m.fileCursor >= len(m.files) {
			m.fileCursor = 0
		}
		r := m.selectedRepo()
		if r != nil {
			m.diffFetching = true
			m.currentDiff = ""
			m.refreshFileViewport()
			return m, m.fetchDiffCmd(r.Path, m.files[m.fileCursor])
		}
	} else {
		m.currentDiff = ""
		m.diffViewport.SetContent("")
	}
	m.refreshFileViewport()
	return m, nil
}

func (m *Model) handleGitDiff(msg gitDiffMsg) (tea.Model, tea.Cmd) {
	m.currentDiff = msg.diff
	m.diffFetching = false
	rendered := m.renderBeautifiedDiff(m.currentDiff)
	m.diffViewport.SetContent(rendered)
	m.diffViewport.GotoTop()
	return m, nil
}

func (m *Model) handleGitBranches(msg gitBranchesMsg) (tea.Model, tea.Cmd) {
	m.branches = msg.branches
	m.showBranches = true
	m.activePanel = LogPanel

	if m.branchCursor >= len(m.branches) {
		m.branchCursor = 0
		if len(m.branches) > 0 {
			m.branchCursor = len(m.branches) - 1
		}
	}
	if m.branchCursor < 0 {
		m.branchCursor = 0
	}

	return m, nil
}

func (m *Model) handleStashFiles(msg stashFilesMsg) (tea.Model, tea.Cmd) {
	m.stashFiles = msg.files
	return m, nil
}

func (m *Model) handleGitStashes(msg gitStashesMsg) (tea.Model, tea.Cmd) {
	m.stashes = msg.stashes
	m.showStashes = true
	m.activePanel = LogPanel
	m.statusMsg = ""
	m.stashFiles = nil

	if m.stashCursor >= len(m.stashes) {
		m.stashCursor = 0
		if len(m.stashes) > 0 {
			m.stashCursor = len(m.stashes) - 1
		}
	}
	if m.stashCursor < 0 {
		m.stashCursor = 0
	}

	r := m.selectedRepo()
	if r != nil && len(m.stashes) > 0 && m.stashCursor < len(m.stashes) {
		return m, m.fetchStashFilesCmd(r.Path, m.stashes[m.stashCursor].Index)
	}
	return m, nil
}

func (m *Model) handleConflictFiles(msg conflictFilesMsg) (tea.Model, tea.Cmd) {
	m.conflictFiles = msg.files
	if len(m.conflictFiles) == 0 {
		m.showConflicts = false
		m.statusMsg = "No merge conflicts"
		return m, nil
	}
	m.showConflicts = true
	m.showFiles = false
	m.showBranches = false
	m.showStashes = false
	m.activePanel = ConflictPanel
	m.statusMsg = ""
	m.conflictCursor = 0
	m.refreshViewports()
	return m, nil
}

func (m *Model) handleCompactDiff(msg compactDiffMsg) (tea.Model, tea.Cmd) {
	m.compactChanges = msg.changes
	m.compactFetching = false
	if len(m.compactChanges) == 0 {
		m.compactDiff = false
		m.statusMsg = "No function-level changes detected"
		return m, nil
	}
	return m, nil
}

func (m *Model) handleGitOperationDone(msg any) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case pushDoneMsg:
		if msg.index >= 0 && msg.index < len(m.repos) {
			r := &m.repos[msg.index]
			r.Pushing = false
			m.appendCommandLog(CommandLogEntry{
				Time:     time.Now(),
				RepoName: r.Name,
				Command:  "push",
				Output:   msg.output,
				Error:    msg.err,
			})
		}
		if msg.err != nil {
			m.statusMsg = "Push failed (see log 'o')"
		} else {
			m.statusMsg = "Push done"
		}
	case pushAllDoneMsg:
		for i := range m.repos {
			m.repos[i].Pushing = false
		}
		for _, res := range msg.results {
			m.appendCommandLog(CommandLogEntry{
				Time:     time.Now(),
				RepoName: res.Name,
				Command:  "push",
				Output:   res.Output,
				Error:    res.Err,
			})
		}
		m.statusMsg = "Push all done"
	case stashDoneMsg:
		if msg.index >= 0 && msg.index < len(m.repos) {
			r := &m.repos[msg.index]
			r.Stashing = false
			m.appendCommandLog(CommandLogEntry{
				Time:     time.Now(),
				RepoName: r.Name,
				Command:  "stash",
				Output:   msg.output,
				Error:    msg.err,
			})
		}
		m.statusMsg = "Stashed"
	case stashPopDoneMsg:
		if msg.index >= 0 && msg.index < len(m.repos) {
			r := &m.repos[msg.index]
			r.Stashing = false
			m.appendCommandLog(CommandLogEntry{
				Time:     time.Now(),
				RepoName: r.Name,
				Command:  "stash pop",
				Output:   msg.output,
				Error:    msg.err,
			})
		}
		m.statusMsg = "Stash popped"
	case stashApplyDoneMsg:
		if msg.index >= 0 && msg.index < len(m.repos) {
			r := &m.repos[msg.index]
			r.Stashing = false
			m.appendCommandLog(CommandLogEntry{
				Time:     time.Now(),
				RepoName: r.Name,
				Command:  "stash apply",
				Output:   msg.output,
				Error:    msg.err,
			})
			if msg.err != nil {
				m.statusMsg = "Stash apply failed (see log 'o')"
			} else {
				m.statusMsg = "Stash applied successfully"
				cmd = tea.Batch(
					m.refreshStatusCmd(msg.index, r.Path),
					m.refreshCachedRepoDetailCmd(msg.index, r.Path),
					m.fetchStashesCmd(r.Path),
				)
			}
		}
	case stashDropDoneMsg:
		if msg.index >= 0 && msg.index < len(m.repos) {
			r := &m.repos[msg.index]
			r.Stashing = false
			m.appendCommandLog(CommandLogEntry{
				Time:     time.Now(),
				RepoName: r.Name,
				Command:  "stash drop",
				Output:   msg.output,
				Error:    msg.err,
			})
			if msg.err != nil {
				m.statusMsg = "Stash drop failed (see log 'o')"
			} else {
				m.statusMsg = "Stash dropped successfully"
				cmd = tea.Batch(
					m.refreshStatusCmd(msg.index, r.Path),
					m.refreshCachedRepoDetailCmd(msg.index, r.Path),
					m.fetchStashesCmd(r.Path),
				)
			}
		}
	case stashPopIndexDoneMsg:
		if msg.index >= 0 && msg.index < len(m.repos) {
			r := &m.repos[msg.index]
			r.Stashing = false
			m.appendCommandLog(CommandLogEntry{
				Time:     time.Now(),
				RepoName: r.Name,
				Command:  "stash pop",
				Output:   msg.output,
				Error:    msg.err,
			})
			if msg.err != nil {
				m.statusMsg = "Stash pop failed (see log 'o')"
			} else {
				m.statusMsg = "Stash popped successfully"
				cmd = tea.Batch(
					m.refreshStatusCmd(msg.index, r.Path),
					m.refreshCachedRepoDetailCmd(msg.index, r.Path),
					m.fetchStashesCmd(r.Path),
				)
			}
		}
	case deleteBranchDoneMsg:
		if msg.index >= 0 && msg.index < len(m.repos) {
			r := &m.repos[msg.index]
			r.CheckingOut = false
			m.appendCommandLog(CommandLogEntry{
				Time:     time.Now(),
				RepoName: r.Name,
				Command:  "delete branch",
				Output:   msg.output,
				Error:    msg.err,
			})
			if msg.err != nil {
				m.statusMsg = "Delete branch failed (see log 'o')"
			} else {
				m.statusMsg = "Branch deleted"
				cmd = m.fetchBranchesCmd(r.Path)
			}
		}
	case deleteRemoteBranchDoneMsg:
		if msg.index >= 0 && msg.index < len(m.repos) {
			r := &m.repos[msg.index]
			r.CheckingOut = false
			m.appendCommandLog(CommandLogEntry{
				Time:     time.Now(),
				RepoName: r.Name,
				Command:  "delete remote branch",
				Output:   msg.output,
				Error:    msg.err,
			})
			if msg.err != nil {
				m.statusMsg = "Delete remote branch failed (see log 'o')"
			} else {
				m.statusMsg = "Remote branch deleted"
				cmd = m.fetchBranchesCmd(r.Path)
			}
		}
	case checkoutBranchDoneMsg:
		if msg.index >= 0 && msg.index < len(m.repos) {
			r := &m.repos[msg.index]
			r.CheckingOut = false
			branchName := ""
			if m.branchCursor >= 0 && m.branchCursor < len(m.branches) {
				branchName = m.branches[m.branchCursor].Name
			}
			m.appendCommandLog(CommandLogEntry{
				Time:     time.Now(),
				RepoName: r.Name,
				Command:  "checkout branch",
				Output:   branchName,
				Error:    msg.err,
			})
			if msg.err != nil {
				m.statusMsg = "Checkout failed: " + msg.err.Error()
			} else {
				m.statusMsg = "Checked out successfully"
				cmd = tea.Batch(
					m.refreshStatusCmd(msg.index, r.Path),
					m.refreshCachedRepoDetailCmd(msg.index, r.Path),
					m.fetchBranchesCmd(r.Path),
				)
			}
		}
	case mergeDoneMsg:
		if msg.index >= 0 && msg.index < len(m.repos) {
			r := &m.repos[msg.index]
			r.Merging = false
			m.appendCommandLog(CommandLogEntry{
				Time:     time.Now(),
				RepoName: r.Name,
				Command:  "merge " + msg.branch,
				Output:   msg.output,
				Error:    msg.err,
			})
			if msg.err != nil {
				m.statusMsg = "Merge failed (see log 'o')"
			} else {
				m.statusMsg = "Merge complete"
				cmd = tea.Batch(
					m.refreshStatusCmd(msg.index, r.Path),
					m.refreshCachedRepoDetailCmd(msg.index, r.Path),
					m.fetchBranchesCmd(r.Path),
					m.fetchConflictFilesCmd(r.Path),
				)
			}
		}
	case openBrowserMsg:
		if msg.err != nil {
			m.statusMsg = "Browser error: " + msg.err.Error()
			m.appendCommandLog(CommandLogEntry{
				Time:     time.Now(),
				RepoName: "browser",
				Command:  "open browser",
				Output:   msg.url,
				Error:    msg.err,
			})
		} else {
			m.statusMsg = "Opened: " + msg.url
			m.appendCommandLog(CommandLogEntry{
				Time:     time.Now(),
				RepoName: "browser",
				Command:  "open browser",
				Output:   msg.url,
			})
		}
	case openEditorMsg:
		if msg.err != nil {
			m.statusMsg = "Editor error: " + msg.err.Error()
			m.appendCommandLog(CommandLogEntry{
				Time:     time.Now(),
				RepoName: "editor",
				Command:  "open editor",
				Output:   msg.editor,
				Error:    msg.err,
			})
		} else {
			m.statusMsg = "Opened in " + msg.editor
			m.appendCommandLog(CommandLogEntry{
				Time:     time.Now(),
				RepoName: "editor",
				Command:  "open editor",
				Output:   msg.editor,
			})
		}
	case editorsDetectedMsg:
		m.availableEditors = msg.editors
		if len(m.availableEditors) == 0 {
			m.statusMsg = "No editors found. Try setting $MONOGIT_EDITOR"
		} else {
			m.showEditorModal = true
			m.editorCursor = 0
			m.statusMsg = ""
		}
	case tagDoneMsg:
		if msg.index >= 0 && msg.index < len(m.repos) {
			r := &m.repos[msg.index]
			r.Tagging = false
			m.appendCommandLog(CommandLogEntry{
				Time:     time.Now(),
				RepoName: r.Name,
				Command:  "tag & push",
				Output:   msg.output,
				Error:    msg.err,
			})
			if msg.err != nil {
				m.statusMsg = "Tag deploy failed (see log 'o')"
			} else {
				m.statusMsg = "Tag deployed successfully"
			}
		}
	case mergetoolDoneMsg:
		if msg.index >= 0 && msg.index < len(m.repos) {
			r := &m.repos[msg.index]
			m.appendCommandLog(CommandLogEntry{
				Time:     time.Now(),
				RepoName: r.Name,
				Command:  "mergetool",
				Output:   "Opened mergetool for " + msg.file,
				Error:    msg.err,
			})
			if msg.err != nil {
				m.statusMsg = "Mergetool failed (see log 'o')"
			} else {
				m.statusMsg = "Merge resolution complete"
				cmd = m.fetchConflictFilesCmd(msg.path)
			}
		}
	case checkoutAllDoneMsg:
		for i := range m.repos {
			m.repos[i].CheckingOut = false
		}
		var failedCount int
		for _, res := range msg.results {
			var err error
			if res.Err != nil && res.Err.Error() == "skipped: not in current filter" {
				err = nil
			} else {
				err = res.Err
			}
			m.appendCommandLog(CommandLogEntry{
				Time:     time.Now(),
				RepoName: res.Name,
				Command:  "checkout " + res.Branch,
				Output:   res.Branch,
				Error:    err,
			})
			if err != nil {
				failedCount++
			}
		}
		if failedCount > 0 {
			m.statusMsg = fmt.Sprintf("Checkout all finished with %d errors (see log 'o')", failedCount)
		} else {
			m.statusMsg = "Checkout all complete"
		}
	case stashAllDoneMsg:
		for i := range m.repos {
			m.repos[i].Stashing = false
		}
		for _, res := range msg.results {
			m.appendCommandLog(CommandLogEntry{
				Time:     time.Now(),
				RepoName: res.Name,
				Command:  "stash",
				Output:   res.Output,
				Error:    res.Err,
			})
		}
		m.statusMsg = "Stash all done"
	}

	m.refreshViewports()
	return m, tea.Batch(cmd, m.refreshAllStatusCmd(m.repos))
}

func (m *Model) handleRefreshMsg() (tea.Model, tea.Cmd) {
	r := m.selectedRepo()
	if r != nil {
		if m.showBranches {
			return m, tea.Batch(m.refreshStatusCmd(m.cursor, r.Path), m.refreshCachedRepoDetailCmd(m.cursor, r.Path), m.fetchBranchesCmd(r.Path))
		}
		return m, tea.Batch(m.refreshStatusCmd(m.cursor, r.Path), m.refreshCachedRepoDetailCmd(m.cursor, r.Path))
	}
	return m, nil
}

func (m *Model) handleNextStepMsg() (tea.Model, tea.Cmd) {
	if m.commitStep == StepMessage {
		r := m.selectedRepo()
		if r != nil {
			m.inputMode = true
			m.inputAction = "commit"
			m.commitInput.Reset()
			m.commitInput.Focus()
			m.statusMsg = "Enter commit message..."
			m.activePanel = CommitWizardPanel
			m.showFiles = false
			return m, m.commitInput.Focus()
		}
	}
	return m, nil
}

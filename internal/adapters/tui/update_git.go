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
		return m, tea.Batch(m.fetchAllCmd(m.repos), tickCmd(m.fetchInterval))
	}
	return m, tickCmd(m.fetchInterval)
}

func (m *Model) handleRepoScanned(msg repoScannedMsg) (tea.Model, tea.Cmd) {
	m.repos = msg.repos
	m.scanning = false
	if len(m.repos) == 0 {
		m.statusMsg = "No Git repositories found."
		return m, tickCmd(m.fetchInterval)
	}
	m.refreshCachedRepoDetail()
	m.refreshViewports()
	return m, tea.Batch(m.refreshAllStatusCmd(m.repos), tickCmd(m.fetchInterval))
}

func (m *Model) handleRepoStatus(msg repoStatusMsg) (tea.Model, tea.Cmd) {
	if msg.index >= 0 && msg.index < len(m.repos) {
		r := &m.repos[msg.index]
		r.CheckingOut = false

		if msg.refresh {
			return m, m.refreshStatusCmd(msg.index, r.Path)
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
	m.refreshCachedRepoDetail()
	m.refreshViewports()
	return m, nil
}

func (m *Model) handleFetchDone(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case fetchDoneMsg:
		if msg.index >= 0 && msg.index < len(m.repos) {
			r := &m.repos[msg.index]
			r.Fetching = false
			return m, m.refreshStatusCmd(msg.index, r.Path)
		}
	case fetchAllDoneMsg:
		for i := range m.repos {
			m.repos[i].Fetching = false
		}
		return m, m.refreshAllStatusCmd(m.repos)
	}
	return m, nil
}

func (m *Model) handlePullDone(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case pullDoneMsg:
		if msg.index >= 0 && msg.index < len(m.repos) {
			r := &m.repos[msg.index]
			r.Pulling = false

			m.commandLogs = append(m.commandLogs, CommandLogEntry{
				Time:     time.Now(),
				RepoName: r.Name,
				Command:  "pull",
				Output:   msg.output,
				Error:    msg.err,
			})

			if msg.err != nil {
				m.statusMsg = fmt.Sprintf("Pull failed for %s (see log 'l')", r.Name)
			} else {
				m.statusMsg = "Pull complete"
			}
			return m, m.refreshStatusCmd(msg.index, r.Path)
		}
	case pullAllDoneMsg:
		for i := range m.repos {
			m.repos[i].Pulling = false
		}
		
		var failedCount int
		for _, res := range msg.results {
			m.commandLogs = append(m.commandLogs, CommandLogEntry{
				Time:     time.Now(),
				RepoName: res.name,
				Command:  "pull",
				Output:   res.output,
				Error:    res.err,
			})
			if res.err != nil {
				failedCount++
			}
		}

		if failedCount > 0 {
			m.statusMsg = fmt.Sprintf("Pull all finished with %d errors (see log 'l')", failedCount)
		} else {
			m.statusMsg = "Pull all complete"
		}
		return m, m.refreshAllStatusCmd(m.repos)
	}
	return m, nil
}

func (m *Model) handleCommitDone(msg commitDoneMsg) (tea.Model, tea.Cmd) {
	if msg.index >= 0 && msg.index < len(m.repos) {
		r := &m.repos[msg.index]
		r.Committing = false
		m.activePanel = RepoPanel
		m.commitStep = StepAddOption
		if msg.err == nil {
			m.showConfirmModal = true
			m.confirmModalTitle = "Commit successful! Push now?"
			m.confirmModalAction = "push"
		}
		return m, m.refreshStatusCmd(msg.index, r.Path)
	}
	return m, nil
}

func (m *Model) handleGitFiles(msg gitFilesMsg) (tea.Model, tea.Cmd) {
	m.files = msg.files
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

func (m *Model) handleGitOperationDone(msg any) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case pushDoneMsg:
		if msg.index >= 0 && msg.index < len(m.repos) {
			r := &m.repos[msg.index]
			r.Pushing = false
			m.commandLogs = append(m.commandLogs, CommandLogEntry{
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
			cmd = clearStatusCmd()
		}
	case pushAllDoneMsg:
		for i := range m.repos {
			m.repos[i].Pushing = false
		}
		for _, res := range msg.results {
			m.commandLogs = append(m.commandLogs, CommandLogEntry{
				Time:     time.Now(),
				RepoName: res.name,
				Command:  "push",
				Output:   res.output,
				Error:    res.err,
			})
		}
		m.statusMsg = "Push all done"
		cmd = clearStatusCmd()
	case stashDoneMsg:
		if msg.index >= 0 && msg.index < len(m.repos) {
			r := &m.repos[msg.index]
			r.Stashing = false
			m.commandLogs = append(m.commandLogs, CommandLogEntry{
				Time:     time.Now(),
				RepoName: r.Name,
				Command:  "stash",
				Output:   msg.output,
				Error:    msg.err,
			})
		}
		m.statusMsg = "Stashed"
		cmd = clearStatusCmd()
	case stashPopDoneMsg:
		if msg.index >= 0 && msg.index < len(m.repos) {
			r := &m.repos[msg.index]
			r.Stashing = false
			m.commandLogs = append(m.commandLogs, CommandLogEntry{
				Time:     time.Now(),
				RepoName: r.Name,
				Command:  "stash pop",
				Output:   msg.output,
				Error:    msg.err,
			})
		}
		m.statusMsg = "Stash popped"
		cmd = clearStatusCmd()
	case deleteBranchDoneMsg:
		if msg.index >= 0 && msg.index < len(m.repos) {
			r := &m.repos[msg.index]
			r.CheckingOut = false // Safety reset
			m.commandLogs = append(m.commandLogs, CommandLogEntry{
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
				cmd = tea.Batch(m.fetchBranchesCmd(r.Path), clearStatusCmd())
			}
		}
	case deleteRemoteBranchDoneMsg:
		if msg.index >= 0 && msg.index < len(m.repos) {
			r := &m.repos[msg.index]
			r.CheckingOut = false // Safety reset
			m.commandLogs = append(m.commandLogs, CommandLogEntry{
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
				cmd = tea.Batch(m.fetchBranchesCmd(r.Path), clearStatusCmd())
			}
		}
	case checkoutBranchDoneMsg:
		if msg.index >= 0 && msg.index < len(m.repos) {
			r := &m.repos[msg.index]
			r.CheckingOut = false
			if msg.err != nil {
				m.statusMsg = "Checkout failed: " + msg.err.Error()
			} else {
				m.statusMsg = "Checked out successfully"
				cmd = tea.Batch(
					m.refreshStatusCmd(msg.index, r.Path),
					m.fetchBranchesCmd(r.Path),
					clearStatusCmd(),
				)
			}
		}
	}

	m.refreshViewports()
	return m, tea.Batch(cmd, m.refreshAllStatusCmd(m.repos))
}

func (m *Model) handleRefreshMsg() (tea.Model, tea.Cmd) {
	r := m.selectedRepo()
	if r != nil {
		if m.showBranches {
			return m, tea.Batch(m.refreshStatusCmd(m.cursor, r.Path), m.fetchBranchesCmd(r.Path))
		}
		return m, m.refreshStatusCmd(m.cursor, r.Path)
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

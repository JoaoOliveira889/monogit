package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) handleConfirmModalKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y", "enter":
		if m.confirmModalAction == "delete_branch_options" {
			return m, nil
		}
		m.showConfirmModal = false
		action := m.confirmModalAction
		m.confirmModalAction = ""
		r := m.selectedRepo()
		if r == nil {
			return m, nil
		}

		switch action {
		case "add_all":
			m.commitStep = StepMessage
			return m, m.addAllAndNextCmd(r.Path)
		case "pull":
			m.statusMsg = "Pulling..."
			r.Pulling = true
			return m, m.pullRepoCmd(m.cursor, r.Path)
		case "push":
			m.statusMsg = "Pushing..."
			r.Pushing = true
			return m, m.pushCmd(m.cursor, r.Path)
		case "create_branch":
			val := m.commitInput.Value()
			m.inputMode = false
			m.commitInput.Reset()
			return m, m.createBranchCmd(r.Path, val)
		case "checkout_branch":
			val := m.branches[m.branchCursor].Name
			r.CheckingOut = true
			return m, m.checkoutBranchCmd(m.cursor, r.Path, val)
		case "discard":
			if len(m.files) > 0 && m.fileCursor < len(m.files) {
				return m, m.discardChangesCmd(r.Path, m.files[m.fileCursor])
			}
		}
		return m, nil

	case "l", "L":
		if m.confirmModalAction == "delete_branch_options" {
			m.showConfirmModal = false
			m.confirmModalAction = ""
			r := m.selectedRepo()
			if r != nil && len(m.branches) > 0 {
				branch := m.branches[m.branchCursor].Name
				m.statusMsg = "Deleting local branch '" + branch + "'..."
				return m, m.deleteBranchCmd(m.cursor, r.Path, branch)
			}
		}
		return m, nil

	case "r", "R":
		if m.confirmModalAction == "delete_branch_options" {
			m.showConfirmModal = false
			m.confirmModalAction = ""
			r := m.selectedRepo()
			if r != nil && len(m.branches) > 0 {
				branch := m.branches[m.branchCursor].Name
				m.statusMsg = "Deleting remote branch 'origin/" + branch + "'..."
				return m, m.deleteRemoteBranchCmd(m.cursor, r.Path, branch)
			}
		}
		return m, nil

	case "n", "N", "esc":
		m.showConfirmModal = false
		m.confirmModalAction = ""
		return m, nil
	}
	return m, nil
}

func (m *Model) handleEditorModalKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		m.editorCursor = clamp(m.editorCursor-1, 0, len(m.availableEditors)-1)
	case "down", "j":
		m.editorCursor = clamp(m.editorCursor+1, 0, len(m.availableEditors)-1)
	case "enter":
		m.showEditorModal = false
		r := m.selectedRepo()
		if r != nil && len(m.availableEditors) > 0 {
			editor := m.availableEditors[m.editorCursor]
			return m, m.openEditorCmd(r.Path, editor)
		}
	case "esc", "q":
		m.showEditorModal = false
	}
	return m, nil
}

func (m *Model) handleNormalKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case matchesKey(msg, keys.Quit...):
		m.quitting = true
		return m, tea.Quit

	case matchesKey(msg, keys.Help...) || matchesKey(msg, keys.HelpAlt...):
		m.showHelp = !m.showHelp
		if m.showHelp {
			m.activePanel = HelpPanel
		} else {
			m.activePanel = RepoPanel
		}
		return m, nil

	case matchesKey(msg, keys.Panel1...):
		return m.handleNumericPanel(0)

	case matchesKey(msg, keys.Panel2...):
		return m.handleNumericPanel(1)

	case matchesKey(msg, keys.Panel3...):
		return m.handleNumericPanel(2)

	case matchesKey(msg, keys.Esc...):
		if m.showHelp {
			m.showHelp = false
			m.activePanel = RepoPanel
			return m, nil
		}
		if m.activePanel == CommandLogPanel {
			m.activePanel = m.previousPanel
			if m.activePanel == RepoPanel {
				m.showBranches = false
				m.showFiles = false
			}
			m.refreshViewports()
			return m, nil
		}
		if m.showFiles || m.showBranches || m.inputMode || m.activePanel == CommitWizardPanel {
			m.cancelSpecialModes()
			m.activePanel = RepoPanel
			m.refreshViewports()
			return m, nil
		}
		return m, nil

	case matchesKey(msg, keys.Tab...):
		if m.activePanel == CommitWizardPanel {
			m.cancelSpecialModes()
			m.activePanel = RepoPanel
			m.refreshViewports()
			return m, nil
		}
		if m.showFiles {
			if m.activePanel == LogPanel {
				m.activePanel = DiffPanel
			} else if m.activePanel == DiffPanel {
				m.cancelSpecialModes()
				m.activePanel = RepoPanel
			} else {
				m.activePanel = LogPanel
			}
		} else {
			if m.activePanel == RepoPanel {
				m.activePanel = LogPanel
			} else {
				m.cancelSpecialModes()
				m.activePanel = RepoPanel
			}
		}
		m.refreshViewports()
		return m, nil

	case matchesKey(msg, keys.Left...):
		if m.activePanel == CommitWizardPanel || m.showFiles || m.showBranches {
			m.cancelSpecialModes()
		}
		m.activePanel = RepoPanel
		m.refreshViewports()
		return m, nil

	case matchesKey(msg, keys.Right...):
		if m.activePanel == CommitWizardPanel {
			m.cancelSpecialModes()
		}
		m.activePanel = LogPanel
		m.refreshViewports()
		return m, nil

	case matchesKey(msg, keys.Up...):
		return m.handleUpKey()

	case matchesKey(msg, keys.Down...):
		return m.handleDownKey()

	case msg.String() == "enter":
		return m.handleEnterKey()

	case msg.String() == " ":
		if m.showFiles && len(m.files) > 0 && m.activePanel != DiffPanel {
			r := m.selectedRepo()
			if r != nil {
				m.fileSelections[m.fileCursor] = !m.fileSelections[m.fileCursor]
				m.refreshFileViewport()
				return m, m.toggleFileCmd(r.Path, m.files[m.fileCursor])
			}
		}
		return m, nil

	case matchesKey(msg, keys.SelectAll...):
		if m.showFiles || m.activePanel == CommitWizardPanel {
			return m.handleSelectAll()
		}
		return m, nil

	case matchesKey(msg, keys.CreateBranch...) || matchesKey(msg, keys.DeselectAll...):
		if m.activePanel == LogPanel && m.showBranches {
			r := m.selectedRepo()
			if r != nil {
				m.inputMode = true
				m.inputAction = "create_branch"
				m.commitInput.Reset()
				m.commitInput.Placeholder = "New branch name..."
				m.commitInput.Focus()
				m.statusMsg = "Enter new branch name..."
				return m, m.commitInput.Focus()
			}
		}
		if m.showFiles && m.commitStep == StepSelectFiles {
			r := m.selectedRepo()
			if r != nil {
				m.fileSelections = make(map[int]bool)
				m.refreshFileViewport()
				return m, m.unstageAllCmd(r.Path)
			}
		}
		return m, nil

	case matchesKey(msg, keys.DeleteBranch...):
		if m.activePanel == LogPanel && m.showBranches && len(m.branches) > 0 {
			branch := m.branches[m.branchCursor].Name
			m.showConfirmModal = true
			m.confirmModalTitle = "Delete branch '" + branch + "'?"
			m.confirmModalAction = "delete_branch_options"
			return m, nil
		}
		return m, nil

	case msg.String() == "b":
		r := m.selectedRepo()
		if r != nil {
			return m, m.fetchBranchesCmd(r.Path)
		}
		return m, nil

	case matchesKey(msg, keys.CommandLog...):
		if m.activePanel == CommandLogPanel {
			m.activePanel = m.previousPanel
		} else {
			m.previousPanel = m.activePanel
			m.activePanel = CommandLogPanel
			m.refreshLogViewport()
			m.logViewport.GotoBottom()
		}
		return m, nil

	case msg.String() == "c":
		r := m.selectedRepo()
		if r != nil {
			m.commitStep = StepAddOption
			m.activePanel = CommitWizardPanel
			m.statusMsg = "Commit: [a] Add All, [s] Select Files"
			return m, nil
		}
		return m, nil

	case msg.String() == "s":
		if m.activePanel == CommitWizardPanel && m.commitStep == StepAddOption {
			r := m.selectedRepo()
			if r != nil {
				m.commitStep = StepSelectFiles
				m.showFiles = true
				m.activePanel = LogPanel
				m.fileCursor = 0
				m.files = nil
				m.fileSelections = make(map[int]bool)
				m.currentDiff = ""
				m.statusMsg = ""
				return m, m.unstageAllCmd(r.Path)
			}
		}
		return m, nil

	case matchesKey(msg, keys.Stash...):
		r := m.selectedRepo()
		if r != nil {
			m.statusMsg = "Stashing..."
			r.Stashing = true
			return m, m.stashCmd(m.cursor, r.Path)
		}
		return m, nil

	case matchesKey(msg, keys.StashPop...):
		r := m.selectedRepo()
		if r != nil {
			m.statusMsg = "Popping stash..."
			r.Stashing = true
			return m, m.stashPopCmd(m.cursor, r.Path)
		}
		return m, nil

	case matchesKey(msg, keys.Discard...):
		if m.showFiles && len(m.files) > 0 {
			m.showConfirmModal = true
			m.confirmModalTitle = "Discard all changes in this file?"
			m.confirmModalAction = "discard"
			return m, nil
		}
		r := m.selectedRepo()
		if r != nil && m.activePanel == LogPanel {
			m.statusMsg = "Undoing commit..."
			return m, m.undoCommitCmd(r.Path)
		}
		return m, nil

	case matchesKey(msg, keys.Fetch...):
		r := m.selectedRepo()
		if r != nil && !r.Fetching {
			r.Fetching = true
			m.statusMsg = "Fetching..."
			return m, m.fetchRepoCmd(m.cursor, r.Path)
		}
		return m, nil

	case matchesKey(msg, keys.FetchAll...):
		if len(m.repos) > 0 {
			m.statusMsg = "Fetching all..."
			for i := range m.repos {
				m.repos[i].Fetching = true
			}
			return m, m.fetchAllCmd()
		}
		return m, nil

	case matchesKey(msg, keys.Pull...):
		r := m.selectedRepo()
		if r != nil && !r.Pulling {
			r.Pulling = true
			m.statusMsg = "Pulling..."
			return m, m.pullRepoCmd(m.cursor, r.Path)
		}
		return m, nil

	case matchesKey(msg, keys.PullAll...):
		if len(m.repos) > 0 {
			for i := range m.repos {
				if !m.repos[i].IsDirty {
					m.repos[i].Pulling = true
				}
			}
			return m, m.pullAllCmd(m.repos)
		}
		return m, nil

	case matchesKey(msg, keys.Push...):
		r := m.selectedRepo()
		if r != nil {
			m.statusMsg = "Pushing..."
			r.Pushing = true
			return m, m.pushCmd(m.cursor, r.Path)
		}
		return m, nil

	case matchesKey(msg, keys.PushAll...):
		if len(m.repos) > 0 {
			for i := range m.repos {
				if m.repos[i].Ahead > 0 {
					m.repos[i].Pushing = true
				}
			}
			return m, m.pushAllCmd(m.repos)
		}
		return m, nil

	case matchesKey(msg, keys.Graph...):
		m.viewGraph = !m.viewGraph
		m.refreshViewports()
		return m, nil

	case matchesKey(msg, keys.OpenEditor...):
		r := m.selectedRepo()
		if r != nil {
			m.statusMsg = "Scanning for editors..."
			return m, m.scanEditorsCmd()
		}
		return m, nil

	case matchesKey(msg, keys.OpenBrowser...):
		r := m.selectedRepo()
		if r != nil {
			m.statusMsg = "Opening in browser..."
			return m, m.openInBrowserCmd(r.Path)
		}
		return m, nil

	case matchesKey(msg, keys.Tag...):
		r := m.selectedRepo()
		if r != nil {
			m.inputMode = true
			m.inputAction = "create_tag_version"
			m.commitInput.Reset()
			m.commitInput.Placeholder = "v1.0.0"
			m.commitInput.Focus()
			m.statusMsg = "Enter tag version..."
			return m, m.commitInput.Focus()
		}
		return m, nil
	}

	return m, nil
}

func (m *Model) handleCursorMove(delta int) (tea.Model, tea.Cmd) {
	if m.activePanel == DiffPanel {
		if delta < 0 {
			m.diffViewport.LineUp(1)
		} else {
			m.diffViewport.LineDown(1)
		}
		return m, nil
	}

	if m.activePanel == RepoPanel {
		maxIdx := len(m.repos) - 1
		newCursor := clamp(m.cursor+delta, 0, maxIdx)
		if newCursor != m.cursor {
			m.cursor = newCursor
			m.refreshCachedRepoDetail()
			m.refreshViewports()
		}
		return m, nil
	}

	if m.showFiles {
		maxIdx := len(m.files) - 1
		newCursor := clamp(m.fileCursor+delta, 0, maxIdx)
		if newCursor != m.fileCursor {
			m.fileCursor = newCursor
			m.refreshFileViewport()
			r := m.selectedRepo()
			if r != nil && len(m.files) > 0 {
				m.diffFetching = true
				return m, m.fetchDiffCmd(r.Path, m.files[m.fileCursor])
			}
		}
		return m, nil
	}

	if m.showBranches {
		maxIdx := len(m.branches) - 1
		m.branchCursor = clamp(m.branchCursor+delta, 0, maxIdx)
		return m, nil
	}

	if m.activePanel == CommandLogPanel {
		if delta < 0 {
			m.logViewport.LineUp(1)
		} else {
			m.logViewport.LineDown(1)
		}
		return m, nil
	}

	if delta < 0 {
		m.viewport.LineUp(1)
	} else {
		m.viewport.LineDown(1)
	}
	return m, nil
}

func (m *Model) handleUpKey() (tea.Model, tea.Cmd) {
	return m.handleCursorMove(-1)
}

func (m *Model) handleDownKey() (tea.Model, tea.Cmd) {
	return m.handleCursorMove(1)
}

func (m *Model) handleEnterKey() (tea.Model, tea.Cmd) {
	if m.activePanel != LogPanel {
		return m, nil
	}

	if m.activePanel == LogPanel && m.showBranches && len(m.branches) > 0 {
		r := m.selectedRepo()
		if r != nil {
			m.showConfirmModal = true
			m.confirmModalTitle = "Checkout branch '" + m.branches[m.branchCursor].Name + "'?"
			m.confirmModalAction = "checkout_branch"
			return m, nil
		}
	}

	if m.showFiles {
		if m.commitStep == StepSelectFiles {
			if len(m.getStagedFiles()) == 0 {
				m.statusMsg = "No files selected"
				return m, nil
			}
			m.commitStep = StepMessage
			m.showFiles = false
			return m, func() tea.Msg { return nextStepMsg{} }
		}
		m.showFiles = false
		m.activePanel = RepoPanel
		return m, nil
	}

	return m, nil
}

func (m *Model) handleSelectAll() (tea.Model, tea.Cmd) {
	if m.showFiles && m.commitStep == StepSelectFiles {
		r := m.selectedRepo()
		if r != nil {
			for i := range m.files {
				m.fileSelections[i] = true
			}
			m.refreshFileViewport()
			return m, m.addAllCmd(r.Path)
		}
	}
	if m.activePanel == CommitWizardPanel && m.commitStep == StepAddOption {
		m.showConfirmModal = true
		m.confirmModalTitle = "Add all files?"
		m.confirmModalAction = "add_all"
		return m, nil
	}
	return m, nil
}

func (m *Model) handleInputKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.inputMode = false
		m.commitInput.Reset()
		m.statusMsg = ""
		return m, nil
	case "enter":
		val := m.commitInput.Value()
		if val == "" {
			m.statusMsg = "Input cannot be empty"
			return m, nil
		}
		r := m.selectedRepo()
		if r == nil {
			return m, nil
		}
		if m.inputAction == "commit" {
			m.inputMode = false
			m.statusMsg = "Committing..."
			r.Committing = true
			return m, m.commitCmd(m.cursor, r.Path, val)
		} else if m.inputAction == "pattern_stage" {
			m.inputMode = false
			m.statusMsg = "Staging..."
			return m, m.stageByPatternCmd(r.Path, val)
		} else if m.inputAction == "create_branch" {
			for _, b := range m.branches {
				if b.Name == val && !b.IsRemote {
					m.statusMsg = "Branch '" + val + "' already exists locally!"
					return m, nil
				}
			}
			m.inputMode = false
			m.statusMsg = "Creating branch '" + val + "'..."
			return m, m.createBranchCmd(r.Path, val)
		} else if m.inputAction == "create_tag_version" {
			m.tagVersion = val
			m.inputAction = "create_tag_message"
			m.commitInput.Reset()
			m.commitInput.Placeholder = "Tag message..."
			m.statusMsg = "Enter tag message..."
			return m, m.commitInput.Focus()
		} else if m.inputAction == "create_tag_message" {
			m.inputMode = false
			m.statusMsg = "Deploying tag " + m.tagVersion + "..."
			r.Tagging = true
			return m, m.createAndPushTagCmd(m.cursor, r.Path, m.tagVersion, val)
		}
		return m, nil
	}
	var cmd tea.Cmd
	m.commitInput, cmd = m.commitInput.Update(msg)
	return m, cmd
}

func (m *Model) handleNumericPanel(index int) (tea.Model, tea.Cmd) {
	visible := m.GetVisiblePanels()
	if index >= 0 && index < len(visible) {
		m.activePanel = visible[index]
		m.refreshViewports()
	}
	return m, nil
}

func clamp(val, minVal, maxVal int) int {
	if val < minVal {
		return minVal
	}
	if val > maxVal {
		return maxVal
	}
	return val
}

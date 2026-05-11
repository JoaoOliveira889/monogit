package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) handleConfirmModalKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y", "enter":
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
			val := m.branches[m.branchCursor]
			r.CheckingOut = true
			return m, m.checkoutBranchCmd(m.cursor, r.Path, val)
		case "discard":
			if len(m.files) > 0 && m.fileCursor < len(m.files) {
				return m, m.discardChangesCmd(r.Path, m.files[m.fileCursor])
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
		if m.showFiles || m.showBranches || m.inputMode || m.activePanel == CommitWizardPanel || m.activePanel == CommandLogPanel {
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
		return m.handleSelectAll()

	case matchesKey(msg, keys.DeselectAll...):
		if m.showFiles && m.commitStep == StepSelectFiles {
			r := m.selectedRepo()
			if r != nil {
				return m, m.unstageAllCmd(r.Path)
			}
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
			m.activePanel = RepoPanel
		} else {
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
			return m, m.fetchAllCmd(m.repos)
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
	}

	return m, nil
}

func (m *Model) handleUpKey() (tea.Model, tea.Cmd) {
	if m.activePanel == DiffPanel {
		m.diffViewport.LineUp(1)
		return m, nil
	}

	if m.activePanel == RepoPanel {
		if m.cursor > 0 {
			m.cursor--
			m.refreshCachedRepoDetail()
			m.refreshViewports()
		}
		return m, nil
	}

	if m.showFiles {
		if m.fileCursor > 0 {
			m.fileCursor--
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
		if m.branchCursor > 0 {
			m.branchCursor--
		}
		return m, nil
	}

	if m.activePanel == CommandLogPanel {
		m.logViewport.LineUp(1)
		return m, nil
	}

	m.viewport.LineUp(1)
	return m, nil
}

func (m *Model) handleDownKey() (tea.Model, tea.Cmd) {
	if m.activePanel == DiffPanel {
		m.diffViewport.LineDown(1)
		return m, nil
	}

	if m.activePanel == RepoPanel {
		if m.cursor < len(m.repos)-1 {
			m.cursor++
			m.refreshCachedRepoDetail()
			m.refreshViewports()
		}
		return m, nil
	}

	if m.showFiles {
		if m.fileCursor < len(m.files)-1 {
			m.fileCursor++
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
		if m.branchCursor < len(m.branches)-1 {
			m.branchCursor++
		}
		return m, nil
	}

	if m.activePanel == CommandLogPanel {
		m.logViewport.LineDown(1)
		return m, nil
	}

	m.viewport.LineDown(1)
	return m, nil
}

func (m *Model) handleEnterKey() (tea.Model, tea.Cmd) {
	if m.activePanel != LogPanel {
		return m, nil
	}

	if m.showBranches && len(m.branches) > 0 {
		r := m.selectedRepo()
		if r != nil {
			m.showConfirmModal = true
			m.confirmModalTitle = "Checkout branch '" + m.branches[m.branchCursor] + "'?"
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
			return m, m.addAllCmd(r.Path)
		}
	}
	if m.activePanel == CommitWizardPanel && m.commitStep == StepAddOption {
		m.showConfirmModal = true
		m.confirmModalTitle = "Add all files?"
		m.confirmModalAction = "add_all"
		return m, nil
	}
	r := m.selectedRepo()
	if r != nil {
		m.statusMsg = ""
		return m, m.fetchFilesCmd(r.Path)
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

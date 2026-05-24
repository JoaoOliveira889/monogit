package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"monogit/internal/pkg/config"
)

func (m *Model) promptConfirm(title, detail, action string) (tea.Model, tea.Cmd) {
	m.showConfirmModal = true
	m.confirmModalTitle = title
	m.confirmModalDetail = detail
	m.confirmModalAction = action
	return m, nil
}

func (m *Model) clearConfirmModal() {
	m.showConfirmModal = false
	m.confirmModalTitle = ""
	m.confirmModalDetail = ""
	m.confirmModalAction = ""
}

func (m *Model) clearPendingActionValues() {
	m.pendingCommitMessage = ""
	m.pendingBranchName = ""
	m.pendingTagVersion = ""
	m.pendingTagMessage = ""
	m.pendingPattern = ""
}

func (m *Model) executeConfirmedAction(action string) (tea.Model, tea.Cmd) {
	r := m.selectedRepo()
	if r == nil {
		return m, nil
	}

	switch action {
	case "pull":
		m.statusMsg = "Pulling..."
		r.Pulling = true
		return m, m.pullRepoCmd(m.cursor, r.Path)
	case "pull_all":
		if len(m.repos) > 0 {
			for i := range m.repos {
				if !m.repos[i].IsDirty {
					m.repos[i].Pulling = true
				}
			}
			return m, m.pullAllCmd(m.repos)
		}
	case "push":
		m.statusMsg = "Pushing..."
		r.Pushing = true
		return m, m.pushCmd(m.cursor, r.Path)
	case "push_all":
		if len(m.repos) > 0 {
			for i := range m.repos {
				if m.repos[i].Ahead > 0 {
					m.repos[i].Pushing = true
				}
			}
			return m, m.pushAllCmd(m.repos)
		}
	case "add_all_commit":
		m.commitStep = StepMessage
		return m, m.addAllAndNextCmd(r.Path)
	case "stage_all_files":
		return m, m.addAllCmd(r.Path)
	case "toggle_file":
		if len(m.files) > 0 && m.fileCursor < len(m.files) {
			return m, m.toggleFileCmd(r.Path, m.files[m.fileCursor])
		}
	case "unstage_all":
		return m, m.unstageAllCmd(r.Path)
	case "prepare_select_files":
		if err := m.gitUC.UnstageAll(r.Path); err != nil {
			return m, func() tea.Msg { return errMsg{Err: err} }
		}
		m.commitStep = StepSelectFiles
		m.showFiles = true
		m.activePanel = LogPanel
		m.fileCursor = 0
		m.files = nil
		m.fileSelections = make(map[int]bool)
		m.currentDiff = ""
		m.statusMsg = ""
		return m, m.fetchFilesCmd(r.Path)
	case "commit":
		if m.pendingCommitMessage != "" {
			m.inputMode = false
			m.statusMsg = "Committing..."
			r.Committing = true
			msg := m.pendingCommitMessage
			m.pendingCommitMessage = ""
			return m, m.commitCmd(m.cursor, r.Path, msg)
		}
	case "create_branch":
		if m.pendingBranchName != "" {
			m.statusMsg = "Creating branch '" + m.pendingBranchName + "'..."
			branch := m.pendingBranchName
			m.pendingBranchName = ""
			return m, m.createBranchCmd(r.Path, branch)
		}
	case "checkout_branch":
		if len(m.branches) > 0 && m.branchCursor < len(m.branches) {
			val := m.branches[m.branchCursor].Name
			r.CheckingOut = true
			return m, m.checkoutBranchCmd(m.cursor, r.Path, val)
		}
	case "discard":
		if len(m.files) > 0 && m.fileCursor < len(m.files) {
			return m, m.discardChangesCmd(r.Path, m.files[m.fileCursor])
		}
	case "pop_stash":
		if len(m.stashes) > 0 && m.stashCursor < len(m.stashes) {
			stashIdx := m.stashes[m.stashCursor].Index
			m.statusMsg = "Popping stash..."
			r.Stashing = true
			return m, m.stashPopIndexCmd(m.cursor, r.Path, stashIdx)
		}
	case "apply_stash":
		if len(m.stashes) > 0 && m.stashCursor < len(m.stashes) {
			stashIdx := m.stashes[m.stashCursor].Index
			m.statusMsg = "Applying stash..."
			r.Stashing = true
			return m, m.stashApplyCmd(m.cursor, r.Path, stashIdx)
		}
	case "stash":
		m.statusMsg = "Stashing..."
		r.Stashing = true
		return m, m.stashCmd(m.cursor, r.Path)
	case "drop_stash":
		if len(m.stashes) > 0 && m.stashCursor < len(m.stashes) {
			stashIdx := m.stashes[m.stashCursor].Index
			m.statusMsg = "Dropping stash..."
			r.Stashing = true
			return m, m.stashDropCmd(m.cursor, r.Path, stashIdx)
		}
	case "undo":
		m.statusMsg = "Undoing commit..."
		return m, m.undoCommitCmd(r.Path)
	case "create_tag":
		if m.pendingTagVersion != "" {
			m.statusMsg = "Deploying tag " + m.pendingTagVersion + "..."
			r.Tagging = true
			name := m.pendingTagVersion
			message := m.pendingTagMessage
			m.pendingTagVersion = ""
			m.pendingTagMessage = ""
			return m, m.createAndPushTagCmd(m.cursor, r.Path, name, message)
		}
	case "stage_pattern":
		if m.pendingPattern != "" {
			m.statusMsg = "Staging..."
			pattern := m.pendingPattern
			m.pendingPattern = ""
			return m, m.stageByPatternCmd(r.Path, pattern)
		}
	case "delete_branch_local":
		if len(m.branches) > 0 {
			branch := m.branches[m.branchCursor].Name
			m.statusMsg = "Deleting local branch '" + branch + "'..."
			return m, m.deleteBranchCmd(m.cursor, r.Path, branch)
		}
	case "delete_branch_remote":
		if len(m.branches) > 0 {
			branch := m.branches[m.branchCursor].Name
			m.statusMsg = "Deleting remote branch 'origin/" + branch + "'..."
			return m, m.deleteRemoteBranchCmd(m.cursor, r.Path, branch)
		}
	}

	return m, nil
}

func (m *Model) handleConfirmModalKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y", "enter":
		if m.confirmModalAction == "delete_branch_options" {
			return m, nil
		}
		action := m.confirmModalAction
		m.clearConfirmModal()
		return m.executeConfirmedAction(action)

	case "l", "L":
		if m.confirmModalAction == "delete_branch_options" {
			m.clearConfirmModal()
			return m.executeConfirmedAction("delete_branch_local")
		}
		return m, nil

	case "r", "R":
		if m.confirmModalAction == "delete_branch_options" {
			m.clearConfirmModal()
			return m.executeConfirmedAction("delete_branch_remote")
		}
		return m, nil

	case "n", "N", "esc":
		m.clearConfirmModal()
		m.clearPendingActionValues()
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
	if m.showStashes {
		switch {
		case matchesKey(msg, keys.StashPop...) || msg.String() == "enter":
			r := m.selectedRepo()
			if r != nil && len(m.stashes) > 0 && m.stashCursor < len(m.stashes) {
				m.showConfirmModal = true
				m.confirmModalTitle = fmt.Sprintf("Pop stash@{%d}?", m.stashes[m.stashCursor].Index)
				m.confirmModalAction = "pop_stash"
				return m, nil
			}
			return m, nil
		case matchesKey(msg, keys.StashApply...):
			r := m.selectedRepo()
			if r != nil && len(m.stashes) > 0 && m.stashCursor < len(m.stashes) {
				m.showConfirmModal = true
				m.confirmModalTitle = fmt.Sprintf("Apply stash@{%d}?", m.stashes[m.stashCursor].Index)
				m.confirmModalAction = "apply_stash"
				return m, nil
			}
			return m, nil
		case matchesKey(msg, keys.StashDrop...):
			r := m.selectedRepo()
			if r != nil && len(m.stashes) > 0 && m.stashCursor < len(m.stashes) {
				m.showConfirmModal = true
				m.confirmModalTitle = fmt.Sprintf("Drop stash@{%d}?", m.stashes[m.stashCursor].Index)
				m.confirmModalAction = "drop_stash"
				return m, nil
			}
			return m, nil
		}
	}

	switch {
	case matchesKey(msg, keys.Quit...):
		m.clearCommandLogs()
		m.clearSelection()
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
		m.clearSelection()
		return m.handleNumericPanel(0)

	case matchesKey(msg, keys.Panel2...):
		m.clearSelection()
		return m.handleNumericPanel(1)

	case matchesKey(msg, keys.Panel3...):
		m.clearSelection()
		return m.handleNumericPanel(2)

	case matchesKey(msg, keys.Esc...):
		if m.showHelp {
			m.showHelp = false
			m.activePanel = RepoPanel
			return m, nil
		}
		if m.activePanel == CommandLogPanel {
			m.clearCommandLogs()
			m.clearSelection()
			m.activePanel = m.previousPanel
			if m.activePanel == RepoPanel {
				m.showBranches = false
				m.showFiles = false
			}
			m.refreshViewports()
			return m, nil
		}
		if m.showFiles || m.showBranches || m.showStashes || m.inputMode || m.activePanel == CommitWizardPanel {
			m.cancelSpecialModes()
			m.activePanel = RepoPanel
			m.refreshViewports()
			return m, nil
		}
		return m, nil

	case matchesKey(msg, keys.Tab...):
		m.clearSelection()
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
		m.clearSelection()
		if m.activePanel == CommitWizardPanel || m.showFiles || m.showBranches || m.showStashes {
			m.cancelSpecialModes()
		}
		m.activePanel = RepoPanel
		m.refreshViewports()
		return m, nil

	case matchesKey(msg, keys.Right...):
		m.clearSelection()
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
			if r := m.selectedRepo(); r != nil {
				file := m.files[m.fileCursor]
				label := file.Name
				if label == "" {
					label = "selected file"
				}
				return m.promptConfirm("Toggle file '"+label+"'?", "This will stage or unstage the file in Git.", "toggle_file")
			}
		}
		return m, nil

	case matchesKey(msg, keys.SelectAll...):
		if m.showFiles || m.activePanel == CommitWizardPanel {
			return m.handleSelectAll()
		}
		return m, nil

	case matchesKey(msg, keys.CreateBranch...):
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
		return m, nil
	case matchesKey(msg, keys.DeselectAll...):
		if m.showFiles && m.commitStep == StepSelectFiles {
			r := m.selectedRepo()
			if r != nil {
				return m.promptConfirm("Unstage all files?", "This will clear the staged selection in Git.", "unstage_all")
			}
		}
		return m, nil

	case matchesKey(msg, keys.DeleteBranch...):
		if m.activePanel == LogPanel && m.showBranches && len(m.branches) > 0 {
			branch := m.branches[m.branchCursor].Name
			m.showConfirmModal = true
			m.confirmModalTitle = "Delete branch '" + branch + "'?"
			m.confirmModalDetail = "Choose `l` for the local branch or `r` for the remote branch."
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
			m.clearCommandLogs()
			m.clearSelection()
			m.activePanel = m.previousPanel
		} else {
			m.clearCommandLogs()
			m.clearSelection()
			m.previousPanel = m.activePanel
			m.activePanel = CommandLogPanel
			m.refreshLogViewport()
			m.logViewport.GotoTop()
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

	case msg.String() == "s" && m.activePanel == CommitWizardPanel && m.commitStep == StepAddOption:
		r := m.selectedRepo()
		if r != nil {
			return m.promptConfirm("Unstage all files first?", "This clears staged files so you can pick individual changes.", "prepare_select_files")
		}
		return m, nil

	case matchesKey(msg, keys.Stash...):
		r := m.selectedRepo()
		if r != nil {
			return m.promptConfirm("Stash all changes?", "This will save all current work into the stash.", "stash")
		}
		return m, nil

	case matchesKey(msg, keys.StashList...):
		r := m.selectedRepo()
		if r != nil {
			m.statusMsg = "Fetching stashes..."
			return m, m.fetchStashesCmd(r.Path)
		}
		return m, nil

	case matchesKey(msg, keys.Discard...):
		if m.showFiles && len(m.files) > 0 {
			file := m.files[m.fileCursor]
			return m.promptConfirm("Discard changes in '"+file.Name+"'?", "This will restore the file from Git.", "discard")
		}
		return m, nil

	case matchesKey(msg, keys.Undo...):
		r := m.selectedRepo()
		if r != nil && m.activePanel == LogPanel {
			return m.promptConfirm("Undo last commit?", "This will perform a soft reset of HEAD~1.", "undo")
		}
		return m, nil

	case matchesKey(msg, keys.Fetch...):
		r := m.selectedRepo()
		if r != nil && !r.Fetching {
			m.statusMsg = "Fetching..."
			r.Fetching = true
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
			return m.promptConfirm("Pull '"+r.Name+"'?", "This will merge remote changes into the working tree.", "pull")
		}
		return m, nil

	case matchesKey(msg, keys.PullAll...):
		if len(m.repos) > 0 {
			return m.promptConfirm("Pull all repositories?", "Dirty repositories will be skipped.", "pull_all")
		}
		return m, nil

	case matchesKey(msg, keys.Push...):
		r := m.selectedRepo()
		if r != nil {
			return m.promptConfirm(fmt.Sprintf("Push '%s' to remote?", r.Name), "This will send the current branch to the remote.", "push")
		}
		return m, nil

	case matchesKey(msg, keys.PushAll...):
		if len(m.repos) > 0 {
			return m.promptConfirm("Push all repositories with pending commits?", "Only repositories ahead of their upstream will be pushed.", "push_all")
		}
		return m, nil

	case matchesKey(msg, keys.Graph...):
		m.viewGraph = !m.viewGraph
		if r := m.selectedRepo(); r != nil {
			m.detailLoading = true
			m.refreshViewports()
			return m, m.refreshCachedRepoDetailCmd(m.cursor, r.Path)
		}
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

	case matchesKey(msg, keys.ResizeLeft...):
		m.leftPanelRatio -= 0.05
		if m.leftPanelRatio < 0.1 {
			m.leftPanelRatio = 0.1
		}
		_ = config.SaveConfig(config.Config{LeftPanelRatio: m.leftPanelRatio})
		return m.handleResize(tea.WindowSizeMsg{Width: m.width, Height: m.height})

	case matchesKey(msg, keys.ResizeRight...):
		m.leftPanelRatio += 0.05
		if m.leftPanelRatio > 0.9 {
			m.leftPanelRatio = 0.9
		}
		_ = config.SaveConfig(config.Config{LeftPanelRatio: m.leftPanelRatio})
		return m.handleResize(tea.WindowSizeMsg{Width: m.width, Height: m.height})

	case matchesKey(msg, keys.Copy...):
		return m.copyCurrentSelection()

	case matchesKey(msg, keys.Paste...):
		if m.inputMode {
			return m.pasteClipboard()
		}
		return m, nil

	case msg.String() == "v":
		return m.toggleSelection()
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
			m.updateSelection(RepoPanel, m.cursor)
			m.refreshViewports()
			r := m.selectedRepo()
			if r != nil {
				m.detailLoading = true
				return m, m.refreshCachedRepoDetailCmd(m.cursor, r.Path)
			}
		}
		return m, nil
	}

	if m.showFiles {
		maxIdx := len(m.files) - 1
		newCursor := clamp(m.fileCursor+delta, 0, maxIdx)
		if newCursor != m.fileCursor {
			m.fileCursor = newCursor
			m.updateSelection(LogPanel, m.fileCursor)
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
		m.updateSelection(LogPanel, m.branchCursor)
		return m, nil
	}

	if m.showStashes {
		maxIdx := len(m.stashes) - 1
		m.stashCursor = clamp(m.stashCursor+delta, 0, maxIdx)
		m.updateSelection(LogPanel, m.stashCursor)
		r := m.selectedRepo()
		if r != nil && len(m.stashes) > 0 && m.stashCursor < len(m.stashes) {
			return m, m.fetchStashFilesCmd(r.Path, m.stashes[m.stashCursor].Index)
		}
		return m, nil
	}

	if m.activePanel == CommandLogPanel {
		maxIdx := len(m.commandLogs) - 1
		newCursor := clamp(m.commandLogCursor+delta, 0, maxIdx)
		if newCursor != m.commandLogCursor {
			m.commandLogCursor = newCursor
			m.updateSelection(CommandLogPanel, m.commandLogCursor)
		}
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

	if m.activePanel == LogPanel && m.showBranches && len(m.branches) > 0 {
		r := m.selectedRepo()
		if r != nil {
			m.showConfirmModal = true
			m.confirmModalTitle = "Checkout branch '" + m.branches[m.branchCursor].Name + "'?"
			m.confirmModalAction = "checkout_branch"
			return m, nil
		}
	}

	return m, nil
}

func (m *Model) handleSelectAll() (tea.Model, tea.Cmd) {
	if m.showFiles && m.commitStep == StepSelectFiles {
		r := m.selectedRepo()
		if r != nil {
			return m.promptConfirm("Stage all files?", "This will stage every file in the current repository.", "stage_all_files")
		}
	}
	if m.activePanel == CommitWizardPanel && m.commitStep == StepAddOption {
		return m.promptConfirm("Add all files before commit?", "This will stage every file and move to the commit message step.", "add_all_commit")
	}
	return m, nil
}

func (m *Model) handleInputKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+v":
		return m.pasteClipboard()
	case "esc":
		m.inputMode = false
		m.commitInput.Reset()
		m.statusMsg = ""
		m.pendingCommitMessage = ""
		m.pendingBranchName = ""
		m.pendingTagVersion = ""
		m.pendingTagMessage = ""
		m.pendingPattern = ""
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
			m.pendingCommitMessage = val
			m.commitInput.Reset()
			return m.promptConfirm("Commit changes?", "This will create a commit with the message you entered.", "commit")
		} else if m.inputAction == "pattern_stage" {
			m.inputMode = false
			m.pendingPattern = val
			m.commitInput.Reset()
			return m.promptConfirm("Stage by pattern?", "This will stage files that match the pattern.", "stage_pattern")
		} else if m.inputAction == "create_branch" {
			for _, b := range m.branches {
				if b.Name == val && !b.IsRemote {
					m.statusMsg = "Branch '" + val + "' already exists locally!"
					return m, nil
				}
			}
			m.inputMode = false
			m.pendingBranchName = val
			m.commitInput.Reset()
			return m.promptConfirm("Create branch '"+val+"'?", "This will create a new local branch in Git.", "create_branch")
		} else if m.inputAction == "create_tag_version" {
			m.pendingTagVersion = val
			m.inputAction = "create_tag_message"
			m.commitInput.Reset()
			m.commitInput.Placeholder = "Tag message..."
			m.statusMsg = "Enter tag message..."
			return m, m.commitInput.Focus()
		} else if m.inputAction == "create_tag_message" {
			m.inputMode = false
			m.pendingTagMessage = val
			m.commitInput.Reset()
			return m.promptConfirm("Create and push tag '"+m.pendingTagVersion+"'?", "This will create an annotated tag and push it to origin.", "create_tag")
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

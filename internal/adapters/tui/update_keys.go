package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoaoOliveira889/monogit/internal/pkg/config"
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
	m.pendingTagName = ""
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
		m.commitMode = CommitModeAll
		m.commitStep = StepMessage
		return m, func() tea.Msg { return nextStepMsg{} }
	case "prepare_select_files":
		m.commitMode = CommitModeSelected
		m.commitStep = StepSelectFiles
		m.showFiles = true
		m.showBranches = false
		m.showStashes = false
		m.showConflicts = false
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
			if m.commitMode == CommitModeSelected {
				selected := m.selectedFiles()
				return m, m.commitSelectedCmd(m.cursor, r.Path, selected, msg)
			}
			return m, m.commitAllCmd(m.cursor, r.Path, msg)
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
	case "merge":
		if len(m.branches) > 0 && m.branchCursor < len(m.branches) {
			branch := m.branches[m.branchCursor].Name
			m.statusMsg = "Merging '" + branch + "'..."
			r.Merging = true
			return m, m.mergeCmd(m.cursor, r.Path, branch)
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
	case "delete_repo_tag":
		if m.pendingTagName != "" {
			tag := m.pendingTagName
			m.pendingTagName = ""
			m.statusMsg = "Removing tag '" + tag + "'..."
			m.removeTagFromRepo(r.Path, tag)
			return m, nil
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
	case "resolve_conflict":
		if len(m.conflictFiles) > 0 && m.conflictCursor < len(m.conflictFiles) {
			m.statusMsg = "Opening mergetool..."
			file := m.conflictFiles[m.conflictCursor].Name
			return m, m.openMergetoolCmd(m.cursor, r.Path, m.cfg.MergeTool, file)
		}
	case "checkout_all":
		branch := m.pendingBranchName
		m.pendingBranchName = ""
		if len(m.repos) > 0 {
			m.statusMsg = "Checking out '" + branch + "' on all filtered repos..."
			filtered := m.filteredRepos()
			filteredPaths := make(map[string]bool, len(filtered))
			for _, fr := range filtered {
				filteredPaths[fr.Path] = true
			}
			for i := range m.repos {
				if filteredPaths[m.repos[i].Path] {
					m.repos[i].CheckingOut = true
				}
			}
			return m, m.checkoutAllCmd(branch)
		}
	case "stash_all":
		if len(m.repos) > 0 {
			m.statusMsg = "Stashing all dirty filtered repos..."
			filtered := m.filteredRepos()
			filteredPaths := make(map[string]bool, len(filtered))
			for _, fr := range filtered {
				filteredPaths[fr.Path] = true
			}
			for i := range m.repos {
				if filteredPaths[m.repos[i].Path] && m.repos[i].IsDirty {
					m.repos[i].Stashing = true
				}
			}
			return m, m.stashAllCmd()
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
	if m.showConflicts {
		switch {
		case msg.String() == "enter":
			r := m.selectedRepo()
			if r != nil && len(m.conflictFiles) > 0 && m.conflictCursor < len(m.conflictFiles) {
				m.showConfirmModal = true
				m.confirmModalTitle = fmt.Sprintf("Resolve '%s'?", m.conflictFiles[m.conflictCursor].Name)
				m.confirmModalDetail = "This will open the configured mergetool and take over the terminal."
				m.confirmModalAction = "resolve_conflict"
				return m, nil
			}
			return m, nil
		case msg.String() == "esc":
			m.showConflicts = false
			m.conflictFiles = nil
			m.activePanel = RepoPanel
			m.refreshViewports()
			return m, nil
		}
	}

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
		if m.searchQuery != "" && m.activePanel == RepoPanel {
			m.searchQuery = ""
			m.searchInput.Reset()
			m.syncCursorToFilter()
			_, _ = m.handleResize(tea.WindowSizeMsg{Width: m.width, Height: m.height})
			m.refreshViewports()
			return m, nil
		}
		if m.tagAssignModal {
			m.tagAssignModal = false
			m.tagEditorRepo = ""
			if m.previousPanel != 0 {
				m.activePanel = m.previousPanel
			} else {
				m.activePanel = RepoPanel
			}
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
		if m.showFiles || m.showBranches || m.showStashes || m.showConflicts || m.inputMode || m.activePanel == CommitWizardPanel {
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
		if m.activePanel == CommitWizardPanel || m.showFiles || m.showBranches || m.showStashes || m.showConflicts {
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

	case m.showHelp && matchesKey(msg, keys.Up...):
		m.helpViewport.LineUp(3)
		return m, nil

	case m.showHelp && matchesKey(msg, keys.Down...):
		m.helpViewport.LineDown(3)
		return m, nil

	case matchesKey(msg, keys.Up...):
		return m.handleUpKey()

	case matchesKey(msg, keys.Down...):
		return m.handleDownKey()

	case msg.String() == "enter":
		return m.handleEnterKey()

	case msg.String() == " ":
		if m.showFiles && len(m.files) > 0 && m.activePanel != DiffPanel {
			m.fileSelections[m.fileCursor] = !m.fileSelections[m.fileCursor]
			m.refreshFileViewport()
		}
		return m, nil

	case matchesKey(msg, keys.SelectAll...):
		if m.showFiles || m.activePanel == CommitWizardPanel {
			return m.handleSelectAll()
		}
		return m, nil

	case matchesKey(msg, keys.DeselectAll...):
		if m.showFiles && m.commitStep == StepSelectFiles {
			m.fileSelections = make(map[int]bool)
			m.refreshFileViewport()
			return m, nil
		}
		return m, nil

	case matchesKey(msg, keys.CreateBranch...) && m.activePanel == LogPanel && m.showBranches:
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

	case matchesKey(msg, keys.Merge...):
		if m.showBranches && len(m.branches) > 0 {
			r := m.selectedRepo()
			if r != nil && !r.Merging {
				branch := m.branches[m.branchCursor].Name
				return m.promptConfirm(
					"Merge '"+branch+"' into current branch?",
					"This will merge the selected branch into HEAD.",
					"merge",
				)
			}
		}
		return m, nil

	case msg.String() == "b":
		r := m.selectedRepo()
		if r != nil {
			return m, m.fetchBranchesCmd(r.Path)
		}
		return m, nil

	case matchesKey(msg, keys.ResolveConflicts...):
		r := m.selectedRepo()
		if r != nil {
			m.statusMsg = "Checking for merge conflicts..."
			return m, m.fetchConflictFilesCmd(r.Path)
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
			m.commitMode = CommitModeAll
			m.showFiles = false
			m.showBranches = false
			m.showStashes = false
			m.showConflicts = false
			m.activePanel = CommitWizardPanel
			m.statusMsg = "Commit: [a] Add All, [v] Select Files"
			return m, nil
		}
		return m, nil

	case msg.String() == "v" && m.activePanel == CommitWizardPanel && m.commitStep == StepAddOption:
		r := m.selectedRepo()
		if r != nil {
			return m.executeConfirmedAction("prepare_select_files")
		}
		return m, nil

	case matchesKey(msg, keys.Stash...):
		r := m.selectedRepo()
		if r != nil {
			return m.promptConfirm("Stash all changes?", "This will save all current work into the stash.", "stash")
		}
		return m, nil

	case matchesKey(msg, keys.BulkCheckout...):
		filtered := m.filteredRepos()
		if len(filtered) == 0 {
			m.statusMsg = "No repositories to checkout"
			return m, nil
		}
		m.inputMode = true
		m.inputAction = "checkout_all_branch"
		m.commitInput.Reset()
		m.commitInput.Placeholder = "Branch name (e.g. main)..."
		m.commitInput.Focus()
		m.statusMsg = "Enter branch name to checkout in " + fmt.Sprintf("%d", len(filtered)) + " repos..."
		return m, m.commitInput.Focus()

	case matchesKey(msg, keys.BulkStash...):
		filtered := m.filteredRepos()
		dirtyCount := 0
		for _, r := range filtered {
			if r.IsDirty {
				dirtyCount++
			}
		}
		if dirtyCount == 0 {
			m.statusMsg = "No dirty repositories to stash"
			return m, nil
		}
		return m.promptConfirm(
			fmt.Sprintf("Stash changes in %d dirty repos?", dirtyCount),
			"Clean repositories will be skipped.",
			"stash_all",
		)

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

	case matchesKey(msg, keys.CompactDiff...):
		if m.showFiles && m.activePanel == DiffPanel {
			m.compactDiff = !m.compactDiff
			if m.compactDiff {
				r := m.selectedRepo()
				if r != nil && len(m.files) > 0 && m.fileCursor < len(m.files) {
					m.compactFetching = true
					return m, m.fetchCompactDiffCmd(r.Path, m.files[m.fileCursor])
				}
			}
			m.refreshViewports()
		}
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
		_ = config.SaveConfig(config.Config{LeftPanelRatio: m.leftPanelRatio, RepoTags: m.cfg.RepoTags})
		return m.handleResize(tea.WindowSizeMsg{Width: m.width, Height: m.height})

	case matchesKey(msg, keys.ResizeRight...):
		m.leftPanelRatio += 0.05
		if m.leftPanelRatio > 0.9 {
			m.leftPanelRatio = 0.9
		}
		_ = config.SaveConfig(config.Config{LeftPanelRatio: m.leftPanelRatio, RepoTags: m.cfg.RepoTags})
		return m.handleResize(tea.WindowSizeMsg{Width: m.width, Height: m.height})

	case matchesKey(msg, keys.TagFilter...):
		return m.toggleTagFilter()

	case matchesKey(msg, keys.TagAssign...):
		return m.toggleTagAssign()

	case matchesKey(msg, keys.Search...):
		if !m.searchMode {
			m.tagAssignModal = false
			m.tagEditorRepo = ""
			m.searchMode = true
			m.activePanel = RepoPanel
			m.searchInput.Reset()
			if m.searchQuery != "" {
				m.searchInput.SetValue(m.searchQuery)
			}
			m.searchInput.Focus()
			_, _ = m.handleResize(tea.WindowSizeMsg{Width: m.width, Height: m.height})
			m.refreshViewports()
			return m, m.searchInput.Focus()
		}

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
		filtered := m.filteredRepos()
		if len(filtered) == 0 {
			return m, nil
		}

		currentFilteredIdx := -1
		for i, r := range filtered {
			if r.Path == m.repos[m.cursor].Path {
				currentFilteredIdx = i
				break
			}
		}
		if currentFilteredIdx < 0 {
			currentFilteredIdx = 0
		}

		newFilteredIdx := clamp(currentFilteredIdx+delta, 0, len(filtered)-1)
		if newFilteredIdx == currentFilteredIdx {
			return m, nil
		}

		newRepo := &filtered[newFilteredIdx]
		for i := range m.repos {
			if m.repos[i].Path == newRepo.Path {
				m.cursor = i
				break
			}
		}

		m.updateSelection(RepoPanel, m.cursor)
		m.refreshViewports()
		r := m.selectedRepo()
		if r != nil {
			m.detailLoading = true
			return m, m.refreshCachedRepoDetailCmd(m.cursor, r.Path)
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
			m.compactDiff = false
			m.compactChanges = nil
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

	if m.showConflicts {
		maxIdx := len(m.conflictFiles) - 1
		m.conflictCursor = clamp(m.conflictCursor+delta, 0, maxIdx)
		m.updateSelection(ConflictPanel, m.conflictCursor)
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
			if len(m.selectedFiles()) == 0 {
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
		for i := range m.files {
			m.fileSelections[i] = true
		}
		m.refreshFileViewport()
		return m, nil
	}
	if m.activePanel == CommitWizardPanel && m.commitStep == StepAddOption {
		return m.executeConfirmedAction("add_all_commit")
	}
	return m, nil
}

func (m *Model) handleInputKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+v":
		return m.pasteClipboard()
	case "esc":
		if m.inputAction == "new_tag" && m.tagAssignModal {
			m.inputMode = false
			m.commitInput.Reset()
			m.commitInput.Placeholder = "Commit message..."
			m.statusMsg = ""
			return m, nil
		}
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
		} else if m.inputAction == "checkout_all_branch" {
			m.inputMode = false
			m.pendingBranchName = val
			m.commitInput.Reset()
			return m.promptConfirm("Checkout '"+val+"' in all filtered repos?", "This will switch branches in every visible repository.", "checkout_all")
		} else if m.inputAction == "new_tag" {
			m.inputMode = false
			m.commitInput.Reset()
			if m.repoHasTag(r.Path, val) {
				m.statusMsg = "Tag already assigned"
				return m, nil
			}
			m.addTagToRepo(r.Path, val)
			m.statusMsg = ""
			return m, nil
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

func (m *Model) toggleTagFilter() (tea.Model, tea.Cmd) {
	if m.tagFilterModal {
		m.tagFilterModal = false
		return m, nil
	}
	m.tagFilterModal = true
	m.tagFilterActive = false
	m.tagFilter = nil
	m.tagModalCursor = 0
	m.tagModalSelections = make(map[int]bool)
	m.refreshAvailableTags()
	return m, nil
}

func (m *Model) toggleTagAssign() (tea.Model, tea.Cmd) {
	if m.tagAssignModal {
		m.tagAssignModal = false
		m.tagEditorRepo = ""
		if m.previousPanel != 0 {
			m.activePanel = m.previousPanel
		} else {
			m.activePanel = RepoPanel
		}
		return m, nil
	}
	if m.selectedRepo() == nil {
		return m, nil
	}
	m.previousPanel = m.activePanel
	m.searchMode = false
	m.tagAssignModal = true
	m.activePanel = LogPanel
	m.tagModalCursor = 0
	m.refreshAvailableTags()
	if r := m.selectedRepo(); r != nil {
		m.tagEditorRepo = r.Path
	}
	return m, nil
}

func (m *Model) handleTagFilterKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.tagFilterModal = false
		m.tagModalSelections = make(map[int]bool)
		return m, nil
	case "up", "k":
		maxIdx := len(m.availableTags) - 1
		m.tagModalCursor = clamp(m.tagModalCursor-1, 0, maxIdx)
		return m, nil
	case "down", "j":
		maxIdx := len(m.availableTags) - 1
		m.tagModalCursor = clamp(m.tagModalCursor+1, 0, maxIdx)
		return m, nil
	case " ":
		if m.tagModalCursor < len(m.availableTags) {
			idx := m.tagModalCursor
			m.tagModalSelections[idx] = !m.tagModalSelections[idx]
		}
		return m, nil
	case "enter":
		m.tagFilter = nil
		for i, selected := range m.tagModalSelections {
			if selected && i < len(m.availableTags) {
				m.tagFilter = append(m.tagFilter, m.availableTags[i])
			}
		}
		m.tagFilterActive = len(m.tagFilter) > 0
		m.tagFilterModal = false
		m.syncCursorToFilter()
		m.refreshViewports()
		return m, nil
	}
	return m, nil
}

func (m *Model) handleTagAssignKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	r := m.selectedRepo()
	if r == nil {
		m.tagAssignModal = false
		m.tagEditorRepo = ""
		return m, nil
	}
	if m.tagEditorRepo != "" && m.tagEditorRepo != r.Path {
		m.tagEditorRepo = r.Path
	}
	switch msg.String() {
	case "esc":
		m.tagAssignModal = false
		m.tagEditorRepo = ""
		if m.previousPanel != 0 {
			m.activePanel = m.previousPanel
		} else {
			m.activePanel = RepoPanel
		}
		return m, nil
	case "up", "k":
		maxIdx := len(r.Tags)
		m.tagModalCursor = clamp(m.tagModalCursor-1, 0, maxIdx)
		return m, nil
	case "down", "j":
		maxIdx := len(r.Tags)
		m.tagModalCursor = clamp(m.tagModalCursor+1, 0, maxIdx)
		return m, nil
	case " ", "enter":
		if m.tagModalCursor < len(r.Tags) {
			m.statusMsg = "Use d to remove the selected tag"
			return m, nil
		}
		if m.repoTagCount(r.Path) >= maxTagsPerRepo {
			m.statusMsg = "Tag limit reached: max 4 per repo"
			return m, nil
		}
		m.inputMode = true
		m.inputAction = "new_tag"
		m.commitInput.Reset()
		m.commitInput.Placeholder = "New tag name..."
		m.commitInput.Focus()
		m.statusMsg = "Enter new tag name..."
		return m, m.commitInput.Focus()
	case "d":
		if m.tagModalCursor < len(r.Tags) {
			tag := r.Tags[m.tagModalCursor]
			m.showConfirmModal = true
			m.confirmModalTitle = "Remove tag '" + tag + "'?"
			m.confirmModalDetail = "This will remove the tag from this repository only."
			m.confirmModalAction = "delete_repo_tag"
			m.pendingTagName = tag
			return m, nil
		}
	}
	return m, nil
}

func (m *Model) addTagToRepo(repoPath, tag string) {
	currentTags := m.cfg.RepoTags[repoPath]
	found := -1
	for i, t := range currentTags {
		if t == tag {
			found = i
			break
		}
	}
	if found >= 0 {
		return
	}
	if len(currentTags) >= maxTagsPerRepo {
		m.statusMsg = "Tag limit reached: max 4 per repo"
		return
	}
	currentTags = append(currentTags, tag)

	if m.cfg.RepoTags == nil {
		m.cfg.RepoTags = make(map[string][]string)
	}
	m.cfg.RepoTags[repoPath] = currentTags

	_ = config.SaveConfig(m.cfg)

	for i := range m.repos {
		if m.repos[i].Path == repoPath {
			tags := m.cfg.RepoTags[repoPath]
			m.repos[i].Tags = tags
			break
		}
	}

	m.refreshAvailableTags()
	m.refreshViewports()
}

func (m *Model) removeTagFromRepo(repoPath, tag string) {
	currentTags := m.cfg.RepoTags[repoPath]
	found := -1
	for i, t := range currentTags {
		if t == tag {
			found = i
			break
		}
	}
	if found < 0 {
		m.statusMsg = "Tag not assigned"
		return
	}

	currentTags = append(currentTags[:found], currentTags[found+1:]...)
	if len(currentTags) == 0 {
		delete(m.cfg.RepoTags, repoPath)
	} else {
		if m.cfg.RepoTags == nil {
			m.cfg.RepoTags = make(map[string][]string)
		}
		m.cfg.RepoTags[repoPath] = currentTags
	}

	_ = config.SaveConfig(m.cfg)

	for i := range m.repos {
		if m.repos[i].Path == repoPath {
			m.repos[i].Tags = m.cfg.RepoTags[repoPath]
			break
		}
	}

	m.refreshAvailableTags()
	m.refreshViewports()
}

func (m *Model) handleSearchKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.searchMode = false
		m.searchInput.Reset()
		if m.searchQuery != "" {
			m.searchInput.SetValue(m.searchQuery)
		}
		m.syncCursorToFilter()
		_, _ = m.handleResize(tea.WindowSizeMsg{Width: m.width, Height: m.height})
		m.refreshViewports()
		return m, nil
	case "enter":
		m.searchQuery = strings.TrimSpace(m.searchInput.Value())
		m.searchInput.Reset()
		if m.searchQuery != "" {
			m.searchInput.SetValue(m.searchQuery)
		}
		m.searchMode = false
		m.syncCursorToFilter()
		_, _ = m.handleResize(tea.WindowSizeMsg{Width: m.width, Height: m.height})
		m.refreshViewports()
		return m, nil
	case "up", "k":
		return m.handleCursorMove(-1)
	case "down", "j":
		return m.handleCursorMove(1)
	}
	var cmd tea.Cmd
	m.searchInput, cmd = m.searchInput.Update(msg)
	m.syncCursorToFilter()
	m.refreshViewports()
	return m, cmd
}

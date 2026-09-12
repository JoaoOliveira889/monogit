package tui

import (
	"fmt"
	"time"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
)

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	oldStatusMsg := m.statusMsg
	var nextModel tea.Model
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		nextModel, cmd = m.handleResize(msg)
	case spinnerTickMsg:
		m.spinnerFrame++
		nextModel, cmd = m, spinnerTickCmdFor(m.isBusy())
	case splashTickMsg:
		if m.showSplash {
			m.splashFrame++
			m.maybeHideSplash()
			if m.showSplash {
				nextModel, cmd = m, splashTickCmd()
			} else {
				nextModel, cmd = m, nil
			}
		} else {
			nextModel, cmd = m, nil
		}
	case tickMsg:
		nextModel, cmd = m.handleTick()
	case repoScannedMsg:
		nextModel, cmd = m.handleRepoScanned(msg)
	case startupReposMsg:
		nextModel, cmd = m.handleStartupRepos(msg)
	case repoStatusMsg:
		nextModel, cmd = m.handleRepoStatus(msg)
	case repoDetailMsg:
		nextModel, cmd = m.handleRepoDetail(msg)
	case repoUnpushedTagMsg:
		nextModel, cmd = m.handleRepoUnpushedTag(msg)
	case fetchDoneMsg:
		nextModel, cmd = m.handleFetchDone(msg)
	case fetchAllDoneMsg:
		nextModel, cmd = m.handleFetchAllDone(msg)
	case pullDoneMsg, pullAllDoneMsg:
		nextModel, cmd = m.handlePullDone(msg)
	case commitDoneMsg:
		nextModel, cmd = m.handleCommitDone(msg)
	case gitFilesMsg:
		nextModel, cmd = m.handleGitFiles(msg)
	case gitDiffMsg:
		nextModel, cmd = m.handleGitDiff(msg)
	case gitBranchesMsg:
		nextModel, cmd = m.handleGitBranches(msg)
	case gitStashesMsg:
		nextModel, cmd = m.handleGitStashes(msg)
	case stashFilesMsg:
		nextModel, cmd = m.handleStashFiles(msg)
	case conflictFilesMsg:
		nextModel, cmd = m.handleConflictFiles(msg)
	case compactDiffMsg:
		nextModel, cmd = m.handleCompactDiff(msg)
	case pushDoneMsg, pushAllDoneMsg, stashDoneMsg, stashPopDoneMsg, deleteBranchDoneMsg, deleteRemoteBranchDoneMsg, checkoutBranchDoneMsg, mergeDoneMsg, openBrowserMsg, openEditorMsg, openWorktreeTerminalMsg, editorsDetectedMsg, tagDoneMsg, stashApplyDoneMsg, stashDropDoneMsg, stashPopIndexDoneMsg, mergetoolDoneMsg, checkoutAllDoneMsg, stashAllDoneMsg, cherryPickDoneMsg, revertDoneMsg:
		nextModel, cmd = m.handleGitOperationDone(msg)
	case refreshMsg:
		nextModel, cmd = m.handleRefreshMsg()
	case nextStepMsg:
		nextModel, cmd = m.handleNextStepMsg()
	case errMsg:
		m.statusMsg = fmt.Sprintf("Error: %s", msg.Err)
		nextModel, cmd = m, nil
	case clearStatusMsg:
		if m.statusMsgID == msg.id {
			m.statusMsg = ""
		}
		nextModel, cmd = m, nil
	case configSavedMsg:
		if msg.err != nil {
			m.statusMsg = fmt.Sprintf("Config save failed: %s", msg.err)
		} else {
			m.statusMsg = "✓ Config saved"
		}
		nextModel, cmd = m, nil
	case exportLogMsg:
		if msg.err != nil {
			m.statusMsg = fmt.Sprintf("Export failed: %s", msg.err)
		} else {
			m.statusMsg = "Command log exported to " + msg.path
		}
		nextModel, cmd = m, nil
	case rebaseCommitsMsg:
		nextModel, cmd = m.handleRebaseCommitsMsg(msg)
	case rebaseDoneMsg:
		nextModel, cmd = m.handleRebaseDoneMsg(msg)
	case worktreePathResolvedMsg:
		title := "Open terminal for worktree branch '" + msg.branch + "'?"
		detail := "Branch is active in another worktree."
		if msg.path != "" {
			title = "Open terminal at worktree for '" + msg.branch + "'?"
			detail = msg.path
		}
		nextModel, cmd = m.promptConfirm(title, detail, "open_worktree_terminal")
	case tea.KeyMsg:
		if m.showSplash && m.splashReady {
			m.showSplash = false
			nextModel, cmd = m, nil
			break
		}
		nextModel, cmd = m.routeKey(msg)
	case tea.MouseMsg:
		nextModel, cmd = m.handleMouse(msg)
	default:
		nextModel, cmd = m, nil
	}

	if updatedModel, ok := nextModel.(*Model); ok {
		if updatedModel.statusMsg != "" && updatedModel.statusMsg != oldStatusMsg && !updatedModel.isStatusPersistent() {
			updatedModel.statusMsgID++
			cmd = tea.Batch(cmd, clearStatusCmd(updatedModel.statusMsgID))
		}
	}

	return nextModel, cmd
}

// routeKey hands the key to whichever layer currently owns the keyboard: the
// innermost open overlay, otherwise the active detail view, otherwise the
// dashboard itself.
func (m *Model) routeKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if overlay, ok := m.topOverlay(); ok {
		switch overlay {
		case OverlayConfirm:
			return m.handleConfirmModalKeys(msg)
		case OverlayHelp:
			return m.handleHelpKeys(msg)
		case OverlayEditorPicker:
			return m.handleEditorModalKeys(msg)
		case OverlaySearch:
			return m.handleSearchKeys(msg)
		case OverlayStatusFilter:
			return m.handleFilterModalKeys(msg)
		case OverlayTagFilter:
			return m.handleTagFilterKeys(msg)
		case OverlayInput:
			return m.handleInputKeys(msg)
		case OverlayPalette:
			return m.handlePaletteKeys(msg)
		case OverlayTagAssign:
			return m.handleTagAssignKeys(msg)
		}
	}

	if m.showDiff() {
		return m.handleDiffKeys(msg)
	}
	if m.showRebase() {
		return m.handleRebaseKeys(msg)
	}
	return m.handleNormalKeys(msg)
}

func (m *Model) handleResize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.width = msg.Width
	m.height = msg.Height

	lpWidth := m.leftPanelWidth()
	lpInternalWidth := lpWidth - 2
	if lpInternalWidth < 0 {
		lpInternalWidth = 0
	}

	vpWidth := m.rightPanelWidth()
	vpInternalWidth := vpWidth - 2
	if vpInternalWidth < 0 {
		vpInternalWidth = 0
	}

	overhead := footerOverhead

	contentHeight := m.height - overhead
	if contentHeight < 0 {
		contentHeight = 0
	}
	repoContentHeight := contentHeight - 2
	if m.searchMode() {
		repoContentHeight -= searchSectionHeight
	}
	if repoContentHeight < 0 {
		repoContentHeight = 0
	}

	detailContentHeight := contentHeight - 2
	if detailContentHeight < 0 {
		detailContentHeight = 0
	}

	lpViewportWidth := lpInternalWidth - 1
	if lpViewportWidth < 0 {
		lpViewportWidth = 0
	}

	vpViewportWidth := vpInternalWidth - 1
	if vpViewportWidth < 0 {
		vpViewportWidth = 0
	}

	if m.repoViewport.Width() == 0 {
		m.repoViewport = viewport.New(viewport.WithWidth(lpViewportWidth), viewport.WithHeight(repoContentHeight))
	} else {
		m.repoViewport.SetWidth(lpViewportWidth)
		m.repoViewport.SetHeight(repoContentHeight)
	}

	if m.viewport.Width() == 0 {
		m.viewport = viewport.New(viewport.WithWidth(vpViewportWidth), viewport.WithHeight(detailContentHeight))
	} else {
		m.viewport.SetWidth(vpViewportWidth)
		m.viewport.SetHeight(detailContentHeight)
	}

	fileViewportWidth := vpViewportWidth
	diffViewportWidth := vpViewportWidth
	fileListHeight := detailContentHeight * fileListHeightPercent / 100
	if fileListHeight < minFileListHeight {
		fileListHeight = minFileListHeight
	}
	diffHeight := detailContentHeight - fileListHeight - diffFileHeaderGap
	if diffHeight < minDiffHeight {
		diffHeight = minDiffHeight
	}
	if m.usesSideBySideDiff() {
		filePaneWidth := vpInternalWidth * 32 / 100
		if filePaneWidth < minPanelWidth {
			filePaneWidth = minPanelWidth
		}
		diffPaneWidth := vpInternalWidth - filePaneWidth - 1
		if diffPaneWidth < minPanelWidth {
			diffPaneWidth = minPanelWidth
			filePaneWidth = vpInternalWidth - diffPaneWidth - 1
		}
		fileViewportWidth = filePaneWidth - 1
		diffViewportWidth = diffPaneWidth - 1
		fileListHeight = detailContentHeight - 1
		diffHeight = detailContentHeight - 1
	}
	if m.fileViewport.Width() == 0 {
		m.fileViewport = viewport.New(viewport.WithWidth(fileViewportWidth), viewport.WithHeight(fileListHeight))
	} else {
		m.fileViewport.SetWidth(fileViewportWidth)
		m.fileViewport.SetHeight(fileListHeight)
	}

	if m.diffViewport.Width() == 0 {
		m.diffViewport = viewport.New(viewport.WithWidth(diffViewportWidth), viewport.WithHeight(diffHeight))
	} else {
		m.diffViewport.SetWidth(diffViewportWidth)
		m.diffViewport.SetHeight(diffHeight)
	}
	if m.logViewport.Width() == 0 {
		m.logViewport = viewport.New(viewport.WithWidth(vpViewportWidth), viewport.WithHeight(detailContentHeight))
	} else {
		m.logViewport.SetWidth(vpViewportWidth)
		m.logViewport.SetHeight(detailContentHeight)
	}

	if m.showDiff() {
		m.refreshDiffViewport()
	}

	m.refreshViewports()
	return m, nil
}

func (m *Model) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if m.showSplash {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.MouseWheelMsg:
		return m.handleMouseWheel(msg)
	case tea.MouseClickMsg:
		return m.handleMouseClick(msg)
	}
	return m, nil
}

func (m *Model) handleMouseWheel(msg tea.MouseWheelMsg) (tea.Model, tea.Cmd) {
	if msg.Button != tea.MouseWheelUp && msg.Button != tea.MouseWheelDown {
		return m, nil
	}

	if m.showHelp() {
		if msg.Button == tea.MouseWheelUp {
			m.helpViewport.ScrollUp(2)
		} else {
			m.helpViewport.ScrollDown(2)
		}
		return m, nil
	}

	now := time.Now()
	if now.Sub(m.lastWheelTime) < wheelDebounce {
		return m, nil
	}
	m.lastWheelTime = now

	delta := 1
	if msg.Button == tea.MouseWheelUp {
		delta = -1
	}

	if _, open := m.topOverlay(); open {
		return m, nil
	}

	// Over the repository list.
	if msg.X < m.leftPanelWidth() {
		if m.activePanel != RepoPanel {
			m.focusRepoPanel()
		}
		return m.handleCursorMove(delta)
	}

	// Over the detail, file or diff panels.
	if m.activePanel == DiffPanel {
		if delta < 0 {
			m.diffViewport.ScrollUp(1)
		} else {
			m.diffViewport.ScrollDown(1)
		}
		return m, nil
	}
	if m.activePanel == RepoPanel {
		m.activePanel = LogPanel
		m.refreshViewports()
	}
	return m.handleCursorMove(delta)
}

func (m *Model) handleMouseClick(msg tea.MouseClickMsg) (tea.Model, tea.Cmd) {
	if msg.Button != tea.MouseLeft {
		return m, nil
	}
	if overlay, open := m.topOverlay(); open && overlay.replacesFrame() {
		return m, nil
	}

	if msg.X < m.leftPanelWidth() {
		if m.activePanel != RepoPanel {
			m.focusRepoPanel()
		}
		return m, nil
	}

	if m.activePanel == RepoPanel {
		m.activePanel = LogPanel
		m.refreshViewports()
	}
	return m, nil
}

// focusRepoPanel moves focus back to the repository list, closing any detail
// view that only makes sense while the right-hand panel has focus.
func (m *Model) focusRepoPanel() {
	switch m.detailView {
	case DetailBranches, DetailStashes, DetailConflicts:
		m.cancelSpecialModes()
	}
	m.activePanel = RepoPanel
	m.refreshViewports()
}

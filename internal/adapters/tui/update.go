package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
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
		nextModel, cmd = m, spinnerTickCmd()
	case tickMsg:
		nextModel, cmd = m.handleTick()
	case repoScannedMsg:
		nextModel, cmd = m.handleRepoScanned(msg)
	case repoStatusMsg:
		nextModel, cmd = m.handleRepoStatus(msg)
	case fetchDoneMsg:
		nextModel, cmd = m.handleFetchDone(msg)
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
	case pushDoneMsg, pushAllDoneMsg, stashDoneMsg, stashPopDoneMsg, deleteBranchDoneMsg, deleteRemoteBranchDoneMsg, checkoutBranchDoneMsg, openBrowserMsg, openEditorMsg, editorsDetectedMsg:
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
	case tea.KeyMsg:
		if m.showConfirmModal {
			nextModel, cmd = m.handleConfirmModalKeys(msg)
		} else if m.showEditorModal {
			nextModel, cmd = m.handleEditorModalKeys(msg)
		} else if m.inputMode {
			nextModel, cmd = m.handleInputKeys(msg)
		} else {
			nextModel, cmd = m.handleNormalKeys(msg)
		}
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

	overhead := 6
	if m.statusMsg != "" {
		overhead = 7
	}

	contentHeight := m.height - overhead
	if contentHeight < 0 {
		contentHeight = 0
	}

	if m.repoViewport.Width == 0 {
		m.repoViewport = viewport.New(lpInternalWidth, contentHeight)
	} else {
		m.repoViewport.Width = lpInternalWidth
		m.repoViewport.Height = contentHeight
	}

	rightContentHeight := contentHeight
	if m.inputMode {
		rightContentHeight -= 4
	}
	if rightContentHeight < 0 {
		rightContentHeight = 0
	}

	if m.viewport.Width == 0 {
		m.viewport = viewport.New(vpInternalWidth, rightContentHeight)
	} else {
		m.viewport.Width = vpInternalWidth
		m.viewport.Height = rightContentHeight
	}

	fileListHeight := rightContentHeight * 30 / 100
	if fileListHeight < 5 {
		fileListHeight = 5
	}
	if m.fileViewport.Width == 0 {
		m.fileViewport = viewport.New(vpInternalWidth, fileListHeight)
	} else {
		m.fileViewport.Width = vpInternalWidth
		m.fileViewport.Height = fileListHeight
	}

	diffHeight := rightContentHeight - fileListHeight - 2
	if diffHeight < 5 {
		diffHeight = 5
	}
	if m.diffViewport.Width == 0 {
		m.diffViewport = viewport.New(vpInternalWidth, diffHeight)
	} else {
		m.diffViewport.Width = vpInternalWidth
		m.diffViewport.Height = diffHeight
	}
	if m.logViewport.Width == 0 {
		m.logViewport = viewport.New(vpInternalWidth, contentHeight)
	} else {
		m.logViewport.Width = vpInternalWidth
		m.logViewport.Height = contentHeight
	}

	m.refreshCachedRepoDetail()
	m.refreshViewports()
	return m, nil
}

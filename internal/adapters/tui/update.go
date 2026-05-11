package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.handleResize(msg)
	case spinnerTickMsg:
		m.spinnerFrame++
		return m, spinnerTickMsgCmd()
	case tickMsg:
		return m.handleTick()
	case repoScannedMsg:
		return m.handleRepoScanned(msg)
	case repoStatusMsg:
		return m.handleRepoStatus(msg)
	case fetchDoneMsg, fetchAllDoneMsg:
		return m.handleFetchDone(msg)
	case pullDoneMsg, pullAllDoneMsg:
		return m.handlePullDone(msg)
	case commitDoneMsg:
		return m.handleCommitDone(msg)
	case gitFilesMsg:
		return m.handleGitFiles(msg)
	case gitDiffMsg:
		return m.handleGitDiff(msg)
	case gitBranchesMsg:
		return m.handleGitBranches(msg)
	case pushDoneMsg, pushAllDoneMsg, stashDoneMsg, stashPopDoneMsg:
		return m.handleGitOperationDone(msg)
	case refreshMsg:
		return m.handleRefreshMsg()
	case nextStepMsg:
		return m.handleNextStepMsg()
	case errMsg:
		m.statusMsg = fmt.Sprintf("Error: %s", msg.err)
		return m, nil
	case tea.KeyMsg:
		if m.showConfirmModal {
			return m.handleConfirmModalKeys(msg)
		}
		if m.inputMode {
			return m.handleInputKeys(msg)
		}
		return m.handleNormalKeys(msg)
	}

	return m, nil
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

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
	case pushDoneMsg, pushAllDoneMsg, stashDoneMsg, stashPopDoneMsg, deleteBranchDoneMsg, deleteRemoteBranchDoneMsg, checkoutBranchDoneMsg, mergeDoneMsg, openBrowserMsg, openEditorMsg, editorsDetectedMsg, tagDoneMsg, stashApplyDoneMsg, stashDropDoneMsg, stashPopIndexDoneMsg, mergetoolDoneMsg, checkoutAllDoneMsg, stashAllDoneMsg:
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
		}
		nextModel, cmd = m, nil
	case exportLogMsg:
		if msg.err != nil {
			m.statusMsg = fmt.Sprintf("Export failed: %s", msg.err)
		} else {
			m.statusMsg = "Command log exported to " + msg.path
		}
		nextModel, cmd = m, nil
	case tea.KeyMsg:
		if m.showConfirmModal {
			nextModel, cmd = m.handleConfirmModalKeys(msg)
		} else if m.showEditorModal {
			nextModel, cmd = m.handleEditorModalKeys(msg)
		} else if m.searchMode {
			nextModel, cmd = m.handleSearchKeys(msg)
		} else if m.tagFilterModal {
			nextModel, cmd = m.handleTagFilterKeys(msg)
		} else if m.inputMode {
			nextModel, cmd = m.handleInputKeys(msg)
		} else if m.tagAssignModal {
			nextModel, cmd = m.handleTagAssignKeys(msg)
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

	overhead := footerOverhead + 1
	if m.statusMsg != "" {
		overhead = footerOverhead + 2
	}

	contentHeight := m.height - overhead
	if contentHeight < 0 {
		contentHeight = 0
	}
	repoContentHeight := contentHeight
	if m.searchMode {
		repoContentHeight -= searchSectionHeight
	}
	if repoContentHeight < 0 {
		repoContentHeight = 0
	}

	detailContentHeight := contentHeight
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

	if m.repoViewport.Width == 0 {
		m.repoViewport = viewport.New(lpViewportWidth, repoContentHeight)
	} else {
		m.repoViewport.Width = lpViewportWidth
		m.repoViewport.Height = repoContentHeight
	}

	if m.viewport.Width == 0 {
		m.viewport = viewport.New(vpViewportWidth, detailContentHeight)
	} else {
		m.viewport.Width = vpViewportWidth
		m.viewport.Height = detailContentHeight
	}

	fileListHeight := detailContentHeight * fileListHeightPercent / 100
	if fileListHeight < minFileListHeight {
		fileListHeight = minFileListHeight
	}
	if m.fileViewport.Width == 0 {
		m.fileViewport = viewport.New(vpViewportWidth, fileListHeight)
	} else {
		m.fileViewport.Width = vpViewportWidth
		m.fileViewport.Height = fileListHeight
	}

	diffHeight := detailContentHeight - fileListHeight - diffFileHeaderGap
	if diffHeight < minDiffHeight {
		diffHeight = minDiffHeight
	}
	if m.diffViewport.Width == 0 {
		m.diffViewport = viewport.New(vpViewportWidth, diffHeight)
	} else {
		m.diffViewport.Width = vpViewportWidth
		m.diffViewport.Height = diffHeight
	}
	if m.logViewport.Width == 0 {
		m.logViewport = viewport.New(vpViewportWidth, detailContentHeight)
	} else {
		m.logViewport.Width = vpViewportWidth
		m.logViewport.Height = detailContentHeight
	}

	m.refreshViewports()
	return m, nil
}

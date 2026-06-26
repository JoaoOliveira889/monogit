package tui

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoaoOliveira889/monogit/internal/pkg/ui"
)

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
			cmds := []tea.Cmd{m.refreshCachedRepoDetailCmd(m.cursor, r.Path)}

			now := time.Now()
			if now.Sub(m.rerenderDebounce) >= adjacentPrefetchDelay {
				m.rerenderDebounce = now
				for _, offset := range []int{1, -1} {
					adjIdx := newFilteredIdx + offset
					if adjIdx < 0 || adjIdx >= len(filtered) {
						continue
					}
					adjRepo := filtered[adjIdx]
					if _, ok := m.detailCache[adjRepo.Path]; ok {
						continue
					}
					for i := range m.repos {
						if m.repos[i].Path == adjRepo.Path {
							cmds = append(cmds, m.refreshCachedRepoDetailCmd(i, adjRepo.Path))
							break
						}
					}
				}
			}

			return m, tea.Batch(cmds...)
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
		if m.stashFilesFocus {
			maxIdx := len(m.stashFiles) - 1
			m.stashFileCursor = clamp(m.stashFileCursor+delta, 0, maxIdx)
			r := m.selectedRepo()
			if r != nil && len(m.stashFiles) > 0 && m.stashFileCursor < len(m.stashFiles) {
				m.diffFetching = true
				return m, m.fetchStashFileDiffCmd(r.Path, m.stashes[m.stashCursor].Index, m.stashFiles[m.stashFileCursor])
			}
			return m, nil
		}
		maxIdx := len(m.stashes) - 1
		m.stashCursor = clamp(m.stashCursor+delta, 0, maxIdx)
		m.updateSelection(LogPanel, m.stashCursor)
		r := m.selectedRepo()
		if r != nil && len(m.stashes) > 0 && m.stashCursor < len(m.stashes) {
			m.stashFileCursor = 0
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

	if m.activePanel == ConfigPanel {
		m.configCursor = clamp(m.configCursor+delta, 0, numConfigOptions-1)
		m.refreshViewports()
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
	if m.activePanel == ConfigPanel {
		if m.configCursor == configMergeToolIdx {
			m.inputMode = true
			m.inputAction = "config_edit_merge_tool"
			m.commitInput.Reset()
			m.commitInput.Placeholder = "vimdiff, meld, kdiff3..."
			m.commitInput.SetValue(m.cfg.MergeTool)
			m.commitInput.Focus()
			m.statusMsg = "Enter default merge tool command..."
			return m, m.commitInput.Focus()
		} else if m.configCursor == configScanExcludesIdx {
			m.inputMode = true
			m.inputAction = "config_edit_scan_excludes"
			m.commitInput.Reset()
			m.commitInput.Placeholder = "node_modules, vendor, dist..."
			m.commitInput.SetValue(strings.Join(m.cfg.ScanExcludes, ", "))
			m.commitInput.Focus()
			m.statusMsg = "Enter exclude directories (comma separated)..."
			return m, m.commitInput.Focus()
		} else if m.configCursor == configThemeIdx {
			nextTheme := getNextTheme(m.cfg.Theme)
			m.cfg.Theme = nextTheme
			ui.ApplyTheme(nextTheme)
			m.refreshViewports()
			return m, saveConfigCmd(m.cfg)
		}
		return m, nil
	}

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

	if m.activePanel == LogPanel && m.showStashes && len(m.stashes) > 0 {
		r := m.selectedRepo()
		if r != nil && len(m.stashFiles) > 0 && !m.stashFilesFocus {
			m.stashFilesFocus = true
			m.stashFileCursor = 0
			m.diffFetching = true
			return m, m.fetchStashFileDiffCmd(r.Path, m.stashes[m.stashCursor].Index, m.stashFiles[0])
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

func isValidCommitHash(s string) bool {
	if len(s) < 7 || len(s) > 40 {
		return false
	}
	for _, r := range s {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')) {
			return false
		}
	}
	return true
}

func getNextTheme(current string) string {
	for i, t := range ui.Themes {
		if strings.EqualFold(t.Name, current) {
			nextIdx := (i + 1) % len(ui.Themes)
			return ui.Themes[nextIdx].Name
		}
	}
	return ui.Themes[0].Name
}

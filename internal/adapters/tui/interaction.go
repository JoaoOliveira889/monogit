package tui

import (
	"fmt"
	"strings"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"

	"monogit/internal/domain"
)

func (m *Model) toggleSelection() (tea.Model, tea.Cmd) {
	panel, index, ok := m.selectionSource()
	if !ok {
		m.statusMsg = "Selection is only available in list-based panels"
		return m, nil
	}

	if m.selectionActive && m.selectionPanel == panel {
		m.clearSelection()
		m.statusMsg = "Selection cleared"
		return m, nil
	}

	m.beginSelection(panel, index)
	m.statusMsg = "Selection started"
	return m, nil
}

func (m *Model) selectionSource() (Panel, int, bool) {
	switch {
	case m.activePanel == RepoPanel:
		return RepoPanel, m.cursor, len(m.repos) > 0
	case m.activePanel == CommandLogPanel:
		return CommandLogPanel, m.commandLogCursor, len(m.commandLogs) > 0
	case m.activePanel == LogPanel && m.showFiles && len(m.files) > 0:
		return LogPanel, m.fileCursor, true
	case m.activePanel == LogPanel && m.showBranches && len(m.branches) > 0:
		return LogPanel, m.branchCursor, true
	case m.activePanel == LogPanel && m.showStashes && len(m.stashes) > 0:
		return LogPanel, m.stashCursor, true
	}
	return RepoPanel, 0, false
}

func (m *Model) copyCurrentSelection() (tea.Model, tea.Cmd) {
	text := m.selectedText()
	if strings.TrimSpace(text) == "" {
		m.statusMsg = "Nothing to copy"
		return m, nil
	}

	if err := clipboard.WriteAll(text); err != nil {
		m.statusMsg = "Clipboard unavailable: " + err.Error()
		return m, nil
	}

	m.statusMsg = "Copied to clipboard"
	return m, nil
}

func (m *Model) pasteClipboard() (tea.Model, tea.Cmd) {
	text, err := clipboard.ReadAll()
	if err != nil {
		m.statusMsg = "Clipboard unavailable: " + err.Error()
		return m, nil
	}
	if text == "" {
		m.statusMsg = "Clipboard is empty"
		return m, nil
	}

	m.commitInput.SetValue(m.commitInput.Value() + text)
	m.commitInput.CursorEnd()
	m.statusMsg = "Pasted from clipboard"
	return m, nil
}

func (m *Model) selectedText() string {
	panel, start, end, ok := m.selectionBounds()
	if ok {
		switch panel {
		case RepoPanel:
			return m.repoSelectionText(start, end)
		case LogPanel:
			if m.showFiles {
				return m.fileSelectionText(start, end)
			}
			if m.showBranches {
				return m.branchSelectionText(start, end)
			}
			if m.showStashes {
				return m.stashSelectionText(start, end)
			}
			return m.repoDetailPlainText()
		case CommandLogPanel:
			return m.commandLogSelectionText(start, end)
		}
	}

	switch {
	case m.activePanel == RepoPanel:
		return m.repoSelectionText(m.cursor, m.cursor)
	case m.activePanel == CommandLogPanel:
		return m.commandLogSelectionText(m.commandLogCursor, m.commandLogCursor)
	case m.activePanel == LogPanel && m.showFiles:
		return m.fileSelectionText(m.fileCursor, m.fileCursor)
	case m.activePanel == LogPanel && m.showBranches:
		return m.branchSelectionText(m.branchCursor, m.branchCursor)
	case m.activePanel == LogPanel && m.showStashes:
		return m.stashSelectionText(m.stashCursor, m.stashCursor)
	case m.activePanel == LogPanel:
		return m.repoDetailPlainText()
	case m.activePanel == DiffPanel:
		return strings.TrimSpace(m.currentDiff)
	default:
		return ""
	}
}

func selectionText[T any](items []T, start, end int, render func(T) string) string {
	if len(items) == 0 {
		return ""
	}
	if start < 0 {
		start = 0
	}
	if end >= len(items) {
		end = len(items) - 1
	}
	if start > end {
		start, end = end, start
	}
	var lines []string
	for i := start; i <= end; i++ {
		lines = append(lines, render(items[i]))
	}
	return strings.Join(lines, "\n")
}

func (m *Model) repoSelectionText(start, end int) string {
	return selectionText(m.repos, start, end, func(r domain.Repository) string { return m.repoLinePlain(r) })
}

func (m *Model) fileSelectionText(start, end int) string {
	return selectionText(m.files, start, end, func(f domain.FileStatus) string { return m.fileLinePlain(f) })
}

func (m *Model) branchSelectionText(start, end int) string {
	return selectionText(m.branches, start, end, func(b domain.BranchInfo) string { return m.branchLinePlain(b) })
}

func (m *Model) stashSelectionText(start, end int) string {
	return selectionText(m.stashes, start, end, func(s domain.StashInfo) string { return m.stashLinePlain(s) })
}

func (m *Model) commandLogSelectionText(start, end int) string {
	return selectionText(m.commandLogs, start, end, func(e CommandLogEntry) string { return m.commandLogLinePlain(e) })
}

func (m *Model) repoLinePlain(r domain.Repository) string {
	parts := []string{r.Name}
	if r.Branch != "" {
		parts = append(parts, fmt.Sprintf("branch=%s", r.Branch))
	}
	parts = append(parts, fmt.Sprintf("ahead=%d", r.Ahead))
	parts = append(parts, fmt.Sprintf("behind=%d", r.Behind))
	if r.IsDirty {
		parts = append(parts, "dirty")
	} else {
		parts = append(parts, "clean")
	}
	return strings.Join(parts, " | ")
}

func (m *Model) fileLinePlain(f domain.FileStatus) string {
	status := "modified"
	switch {
	case f.Untracked:
		status = "untracked"
	case f.Modified:
		status = "modified"
	case f.Staged:
		status = "staged"
	}
	return fmt.Sprintf("%s | %s", status, f.Name)
}

func (m *Model) branchLinePlain(b domain.BranchInfo) string {
	role := []string{}
	if b.IsLocal {
		role = append(role, "local")
	}
	if b.IsRemote {
		role = append(role, "remote")
	}
	if b.IsCurrent {
		role = append(role, "current")
	}
	if len(role) == 0 {
		role = append(role, "branch")
	}
	return fmt.Sprintf("%s | %s", b.Name, strings.Join(role, ", "))
}

func (m *Model) stashLinePlain(s domain.StashInfo) string {
	return fmt.Sprintf("stash@{%d} | %s", s.Index, s.Message)
}

func (m *Model) commandLogLinePlain(entry CommandLogEntry) string {
	status := "ok"
	if entry.Error != nil {
		status = "error=" + entry.Error.Error()
	}
	out := strings.TrimSpace(entry.Output)
	if out == "" {
		out = "no output"
	}
	return fmt.Sprintf("%s | %s | %s | %s", entry.Time.Format("15:04:05"), entry.RepoName, entry.Command, status+" | "+out)
}

func (m *Model) repoDetailPlainText() string {
	r := m.selectedRepo()
	if r == nil {
		return "No repository selected"
	}

	var lines []string
	lines = append(lines, fmt.Sprintf("Branch: %s", r.Branch))
	if r.IsDirty {
		lines = append(lines, "Status: Modified")
	} else {
		lines = append(lines, "Status: Clean")
	}
	lines = append(lines, fmt.Sprintf("Modified: %d", m.cachedModifiedCount))
	lines = append(lines, fmt.Sprintf("Untracked: %d", m.cachedUntrackedCount))
	lines = append(lines, fmt.Sprintf("Ahead: %d", r.Ahead))
	lines = append(lines, fmt.Sprintf("Behind: %d", r.Behind))

	if m.cachedLastCommit != "" {
		lines = append(lines, "", "Last Commit:", "  "+m.cachedLastCommit)
	}

	if r.Error != "" {
		lines = append(lines, "", "Error: "+r.Error)
	}

	if m.cachedLog != "" {
		lines = append(lines, "", "Recent Commits:", m.cachedLog)
	}

	if r.LastOutput != "" {
		lines = append(lines, "", "Last Output:", r.LastOutput)
	}

	return strings.Join(lines, "\n")
}

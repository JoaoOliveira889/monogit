package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

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

func (m *Model) addTagToRepo(repoPath, tag string) tea.Cmd {
	currentTags := m.cfg.RepoTags[repoPath]
	found := -1
	for i, t := range currentTags {
		if t == tag {
			found = i
			break
		}
	}
	if found >= 0 {
		return nil
	}
	if len(currentTags) >= maxTagsPerRepo {
		m.statusMsg = "Tag limit reached: max 4 per repo"
		return nil
	}
	currentTags = append(currentTags, tag)

	if m.cfg.RepoTags == nil {
		m.cfg.RepoTags = make(map[string][]string)
	}
	m.cfg.RepoTags[repoPath] = currentTags

	for i := range m.repos {
		if m.repos[i].Path == repoPath {
			m.repos[i].Tags = m.cfg.RepoTags[repoPath]
			break
		}
	}

	m.refreshAvailableTags()
	m.refreshViewports()
	return saveConfigCmd(m.cfg)
}

func (m *Model) removeTagFromRepo(repoPath, tag string) tea.Cmd {
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
		return nil
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

	for i := range m.repos {
		if m.repos[i].Path == repoPath {
			m.repos[i].Tags = m.cfg.RepoTags[repoPath]
			break
		}
	}

	m.refreshAvailableTags()
	m.refreshViewports()
	return saveConfigCmd(m.cfg)
}

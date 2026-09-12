package tui

import (
	tea "charm.land/bubbletea/v2"

	"github.com/JoaoOliveira889/monogit/internal/domain"
)

// maxJumpList bounds the jump history, matching the order of magnitude of Vim's
// jumplist rather than growing without limit.
const maxJumpList = 100

// recordJump appends the current repository to the jump history, so ctrl+o can
// return to it after a search, filter or motion moves the cursor far away.
func (m *Model) recordJump() {
	r := m.selectedRepo()
	if r == nil {
		return
	}
	if len(m.jumpList) > 0 && m.jumpList[len(m.jumpList)-1] == r.Path {
		return
	}

	// A new jump discards anything ahead of the current position, exactly as
	// editing after an undo discards the redo branch.
	m.jumpList = append(m.jumpList[:m.jumpIndex], r.Path)
	if len(m.jumpList) > maxJumpList {
		m.jumpList = m.jumpList[len(m.jumpList)-maxJumpList:]
	}
	m.jumpIndex = len(m.jumpList)
}

// jumpTo moves the cursor by steps through the jump history. Negative steps go
// back, positive forward.
func (m *Model) jumpTo(steps int) (tea.Model, tea.Cmd) {
	if len(m.jumpList) == 0 {
		m.statusMsg = "Jump list is empty"
		return m, nil
	}

	// jumpIndex sits one past the newest entry when the cursor has not moved
	// through the history yet, so clamping alone would turn a forward jump from
	// that position into a backward one.
	target := clamp(m.jumpIndex+steps, 0, len(m.jumpList)-1)
	if steps > 0 && target <= m.jumpIndex {
		m.statusMsg = "Already at the newest jump"
		return m, nil
	}
	if steps < 0 && target >= m.jumpIndex && m.jumpIndex == 0 {
		m.statusMsg = "Already at the oldest jump"
		return m, nil
	}
	if target == m.jumpIndex {
		return m, nil
	}
	m.jumpIndex = target

	path := m.jumpList[target]
	for i := range m.repos {
		if m.repos[i].Path == path {
			m.cursor = i
			m.refreshViewports()
			return m, m.refreshSelectedRepoDetailCmd()
		}
	}

	m.statusMsg = "Jump target is no longer in the list"
	return m, nil
}

// needsAttention reports whether a repository is one the user probably wants to
// stop at: dirty, out of sync, or conflicted.
func needsAttention(r domain.Repository) bool {
	return r.IsDirty || r.HasConflicts || r.Ahead > 0 || r.Behind > 0
}

// jumpToAttention moves to the next or previous repository needing attention,
// wrapping around the filtered list.
func (m *Model) jumpToAttention(direction, count int) (tea.Model, tea.Cmd) {
	filtered := m.filteredRepos()
	if len(filtered) == 0 {
		return m, nil
	}

	current := 0
	if r := m.selectedRepo(); r != nil {
		for i := range filtered {
			if filtered[i].Path == r.Path {
				current = i
				break
			}
		}
	}

	found := -1
	for range count {
		next := -1
		for step := 1; step <= len(filtered); step++ {
			idx := ((current+direction*step)%len(filtered) + len(filtered)) % len(filtered)
			if needsAttention(filtered[idx]) {
				next = idx
				break
			}
		}
		if next < 0 {
			break
		}
		current = next
		found = next
	}

	if found < 0 {
		m.statusMsg = "No repository needs attention"
		return m, nil
	}

	m.recordJump()
	for i := range m.repos {
		if m.repos[i].Path == filtered[found].Path {
			m.cursor = i
			break
		}
	}
	m.refreshViewports()
	return m, m.refreshSelectedRepoDetailCmd()
}

// repeatableAction records the last mutating action so "." can replay it.
type repeatableAction struct {
	action string
	label  string
}

// rememberRepeatable stores an action for ".". Actions that depend on text the
// user typed are not stored, because replaying them silently would apply a
// stale message or branch name.
func (m *Model) rememberRepeatable(action string) {
	label, ok := repeatableActions[action]
	if !ok {
		return
	}
	m.lastAction = repeatableAction{action: action, label: label}
}

// repeatableActions are the actions "." can replay: those that act on whatever
// repository is selected now, with no further input. Actions needing a selected
// file or a typed value are deliberately absent, since replaying them would
// either do nothing or reuse a stale value.
var repeatableActions = map[string]string{
	"fetch": "fetch",
	"pull":  "pull",
	"push":  "push",
	"stash": "stash",
	"undo":  "undo last commit",
}

func (m *Model) repeatLastAction() (tea.Model, tea.Cmd) {
	if m.lastAction.action == "" {
		m.statusMsg = "Nothing to repeat"
		return m, nil
	}
	r := m.selectedRepo()
	if r == nil {
		return m, nil
	}
	return m.promptConfirm(
		"Repeat "+m.lastAction.label+" on '"+r.Name+"'?",
		"Repeats the last action on the selected repository.",
		m.lastAction.action,
	)
}

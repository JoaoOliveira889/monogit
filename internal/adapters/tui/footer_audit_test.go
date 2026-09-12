package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/JoaoOliveira889/monogit/internal/domain"
)

// footerHints extracts the key names the footer advertises.
func footerHints(m *Model) []string {
	text := stripANSI(m.renderFooter())
	text = strings.TrimSuffix(text, "? help  MonoGit "+Version)
	var keysShown []string
	for _, part := range strings.Split(text, "•") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		// A hint reads "<keys> <action>", where <keys> may list alternatives
		// separated by "|", as in "d | esc close".
		fields := strings.Fields(part)
		for _, f := range fields {
			if f == "|" {
				continue
			}
			keysShown = append(keysShown, f)
			// The action follows the last key, so stop at the first field that
			// is not followed by a "|".
			break
		}
		for i := 1; i+1 < len(fields); i++ {
			if fields[i] == "|" {
				keysShown = append(keysShown, fields[i+1])
			}
		}
	}
	return keysShown
}

func TestFooterAdvertisesOnlyWorkingKeys(t *testing.T) {
	states := []struct {
		name  string
		setup func(*Model)
		// banned keys must not appear: they do nothing in this state.
		banned []string
		// wanted keys must appear: they are this state's primary actions.
		wanted []string
	}{
		{
			name:   "repo panel",
			setup:  func(m *Model) { m.setDetailView(DetailLog); m.activePanel = RepoPanel },
			wanted: []string{"jk", "d", "f"},
		},
		{
			name:   "diff viewer",
			setup:  func(m *Model) { m.setDetailView(DetailDiff); m.activePanel = DiffPanel },
			banned: []string{"enter", "gl", "b", "f"},
			wanted: []string{"jk", "J/K", "esc"},
		},
		{
			name:   "branches",
			setup:  func(m *Model) { m.setDetailView(DetailBranches); m.activePanel = LogPanel },
			banned: []string{"gl"},
			wanted: []string{"enter", "M", "d"},
		},
		{
			name:   "rebase",
			setup:  func(m *Model) { m.setDetailView(DetailRebase); m.activePanel = RebasePanel },
			banned: []string{"f", "gl"},
			wanted: []string{"enter", "esc"},
		},
		{
			name:   "config",
			setup:  func(m *Model) { m.activePanel = ConfigPanel },
			banned: []string{"d", "f", "gl"},
			wanted: []string{"enter", "esc"},
		},
	}

	for _, st := range states {
		t.Run(st.name, func(t *testing.T) {
			m := NewModel("/tmp", 0, nil)
			m.showSplash = false
			m.repos = []domain.Repository{{Name: "r", Path: "/r"}}
			m.invalidateFilterCache()
			m.handleResize(tea.WindowSizeMsg{Width: 200, Height: 40})
			st.setup(&m)

			shown := footerHints(&m)
			joined := strings.Join(shown, " ")

			for _, b := range st.banned {
				for _, s := range shown {
					if s == b {
						t.Errorf("footer offers %q, which does nothing here: %v", b, shown)
					}
				}
			}
			for _, w := range st.wanted {
				if !strings.Contains(joined, w) {
					t.Errorf("footer is missing %q: %v", w, shown)
				}
			}
		})
	}
}

// expandHint turns a footer hint into the keystrokes it stands for. "jk" means
// j and k, "p/u" means p and u, "ctrl+d/u" means ctrl+d and ctrl+u.
func expandHint(hint string) []string {
	switch hint {
	case "jk":
		return []string{"j", "k"}
	case "J/K":
		return []string{"J", "K"}
	case "p/u":
		return []string{"p", "u"}
	case "ctrl+d/u":
		return []string{"ctrl+d", "ctrl+u"}
	case "↑↓", "jk/↑↓", "↑↓/jk":
		return []string{"up", "down"}
	case "p/s/f/r/d":
		return []string{"p", "s", "f", "r", "d"}
	case "ctrl+w":
		return nil // a prefix, meaningless on its own
	}
	if strings.Contains(hint, "/") && !strings.HasPrefix(hint, "ctrl") {
		return strings.Split(hint, "/")
	}
	return []string{hint}
}

// pressSequence sends a multi-key binding such as "gg" or "gl" one key at a
// time, the way a person types it.
func pressSequence(m *Model, keystroke string) tea.Cmd {
	if len(keystroke) > 1 && !strings.Contains(keystroke, "+") &&
		!strings.Contains(keystroke, "ctrl") && isPlainRunes(keystroke) {
		var last tea.Cmd
		for _, r := range keystroke {
			_, last = m.routeKey(keyPress(string(r)))
		}
		return last
	}
	_, cmd := m.routeKey(keyPress(keystroke))
	return cmd
}

func isPlainRunes(s string) bool {
	for _, r := range s {
		if r < 32 || r > 126 {
			return false
		}
	}
	// Named keys are words, not sequences.
	switch s {
	case "enter", "esc", "tab", "space", "up", "down", "left", "right",
		"home", "end", "pgup", "pgdown", "backspace", "delete":
		return false
	}
	return true
}

// TestEveryAdvertisedKeyDoesSomething presses each key the footer offers and
// requires it to change the model or dispatch a command. A hint that does
// nothing is worse than no hint: it teaches a binding that is not there.
func TestEveryAdvertisedKeyDoesSomething(t *testing.T) {
	states := []struct {
		name  string
		setup func(*Model)
	}{
		{"repo panel", func(m *Model) { m.setDetailView(DetailLog); m.activePanel = RepoPanel }},
		{"log panel", func(m *Model) { m.setDetailView(DetailLog); m.activePanel = LogPanel }},
		{"file list", func(m *Model) { m.setDetailView(DetailFiles); m.activePanel = LogPanel }},
		{"diff viewer", func(m *Model) { m.setDetailView(DetailDiff); m.activePanel = DiffPanel }},
		{"branches", func(m *Model) { m.setDetailView(DetailBranches); m.activePanel = LogPanel }},
		{"stashes", func(m *Model) { m.setDetailView(DetailStashes); m.activePanel = LogPanel }},
		{"conflicts", func(m *Model) { m.setDetailView(DetailConflicts); m.activePanel = ConflictPanel }},
		{"rebase", func(m *Model) { m.setDetailView(DetailRebase); m.activePanel = RebasePanel }},
		{"config", func(m *Model) { m.activePanel = ConfigPanel }},
		{"command log", func(m *Model) { m.activePanel = CommandLogPanel }},
	}

	for _, st := range states {
		t.Run(st.name, func(t *testing.T) {
			base := func() *Model {
				m := NewModel("/tmp", 0, nil)
				m.showSplash = false
				for i := range 8 {
					m.repos = append(m.repos, domain.Repository{
						Name: fmt.Sprintf("repo-%d", i), Path: fmt.Sprintf("/r%d", i),
						IsDirty: i%2 == 0, Ahead: i % 3,
					})
				}
				m.invalidateFilterCache()
				m.handleResize(tea.WindowSizeMsg{Width: 200, Height: 40})
				fillEveryList(&m)
				st.setup(&m)

				// Park every cursor mid-list: a key that cannot move because it
				// is already at a boundary is not a broken binding.
				m.cursor = len(m.repos) / 2
				m.cachedLogFor = m.repos[m.cursor].Path
				m.cachedDetailFor = m.cachedLogFor
				m.fileCursor = len(m.files) / 2
				m.branchCursor = len(m.branches) / 2
				m.stashCursor = len(m.stashes) / 2
				m.conflictCursor = len(m.conflictFiles) / 2
				m.rebaseCursor = len(m.rebaseItems) / 2
				m.commandLogCursor = len(m.commandLogs) / 2
				m.configCursor = 1
				m.refreshViewports()
				m.syncViewports()
				m.diffViewport.SetHeight(20)
				m.diffViewport.SetWidth(80)
				m.refreshDiffViewport()
				m.viewport.SetYOffset(3)
				m.diffViewport.SetYOffset(3)
				m.logViewport.SetYOffset(3)
				return &m
			}

			for _, hint := range footerHints(base()) {
				for _, key := range expandHint(hint) {
					if isFooterLabel(key) {
						continue
					}
					m := base()
					prepareForKey(m, key)
					before := snapshotState(m)

					cmd := pressSequence(m, key)

					if cmd == nil && snapshotState(m) == before {
						t.Errorf("footer offers %q (from hint %q) but it does nothing", key, hint)
					}
				}
			}
		})
	}
}

// snapshotState captures the parts of the model a keystroke is expected to
// move, so "did anything happen" can be answered without listing every field.
func snapshotState(m *Model) string {
	return fmt.Sprintf("%d|%v|%d|%d|%d|%d|%d|%v|%q|%v|%d|%d|%d|%v|%v|%v|%v|%v|%v|%d|%v|%d|%d",
		m.activePanel, m.detailView, m.cursor, m.fileCursor, m.branchCursor,
		m.stashCursor, m.commandLogCursor, m.overlays, m.statusMsg, m.quitting,
		m.viewport.YOffset(), m.diffViewport.YOffset(), m.logViewport.YOffset(),
		m.viewGraph, m.stashFilesFocus, m.confirmModalAction,
		m.rebaseItems, m.fileSelections, m.configCursor, m.rebaseCursor, m.commitStep,
		m.conflictCursor, m.stashFileCursor)
}

// prepareForKey puts the model in a state where the key has something to do.
// Rebase action keys are idempotent — pressing "p" on a commit already marked
// pick correctly changes nothing — so the commit under the cursor is set to a
// different action first.
func prepareForKey(m *Model, key string) {
	if !m.showRebase() || m.rebaseCursor >= len(m.rebaseItems) {
		return
	}
	actions := map[string]string{"p": "pick", "s": "squash", "f": "fixup", "r": "reword", "d": "drop"}
	if want, ok := actions[key]; ok && m.rebaseItems[m.rebaseCursor].Action == want {
		m.rebaseItems[m.rebaseCursor].Action = "pick"
		if want == "pick" {
			m.rebaseItems[m.rebaseCursor].Action = "drop"
		}
	}
}

// isFooterLabel filters out words that are part of a hint's description rather
// than a key, such as the "none" in "a | n  all | none".
func isFooterLabel(s string) bool {
	switch s {
	case "none", "all", "copy", "back", "nav", "page", "scroll":
		return true
	}
	return false
}

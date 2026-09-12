package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/JoaoOliveira889/monogit/internal/domain"
)

func layoutModel(t *testing.T, w, h int) *Model {
	t.Helper()
	m := NewModel("/tmp", 0, nil)
	m.showSplash = false
	m.repos = []domain.Repository{{Name: "repo", Path: "/tmp/repo", Branch: "main", IsDirty: true}}
	m.invalidateFilterCache()
	m.handleResize(tea.WindowSizeMsg{Width: w, Height: h})
	return &m
}

func TestFrameNeverExceedsTerminal(t *testing.T) {
	for _, size := range [][2]int{{100, 30}, {140, 44}, {80, 24}, {60, 20}} {
		m := layoutModel(t, size[0], size[1])
		m.setDetailView(DetailFiles)
		m.files = []domain.FileStatus{
			{Name: "internal/adapters/tui/a_very_long_path_that_should_be_truncated_somewhere.go", Modified: true},
		}
		m.currentDiff = strings.Repeat("+ a line of diff that is far wider than any panel could ever be\n", 20)
		m.diffViewport.SetContent(m.renderBeautifiedDiff(m.currentDiff))
		m.activePanel = DiffPanel
		m.refreshViewports()

		frame := m.render()
		lines := strings.Split(frame, "\n")
		for i, l := range lines {
			if got := lipgloss.Width(l); got > size[0] {
				t.Errorf("w=%d h=%d: line %d width %d exceeds terminal: %q",
					size[0], size[1], i, got, stripANSI(l))
				break
			}
		}
		if len(lines) > size[1] {
			t.Errorf("w=%d h=%d: frame has %d lines, exceeds terminal height",
				size[0], size[1], len(lines))
		}
	}
}

func TestEveryViewKeepsTheFooterVisible(t *testing.T) {
	sizes := [][2]int{{200, 50}, {140, 44}, {120, 30}, {100, 24}, {90, 20}, {80, 40}, {60, 15}}
	views := []struct {
		name  string
		setup func(*Model)
	}{
		{"log", func(m *Model) { m.setDetailView(DetailLog); m.activePanel = LogPanel }},
		{"files", func(m *Model) { m.setDetailView(DetailFiles); m.activePanel = LogPanel }},
		{"diff", func(m *Model) { m.setDetailView(DetailFiles); m.activePanel = DiffPanel }},
		{"branches", func(m *Model) { m.setDetailView(DetailBranches); m.activePanel = LogPanel }},
		{"stashes", func(m *Model) { m.setDetailView(DetailStashes); m.activePanel = LogPanel }},
		{"conflicts", func(m *Model) { m.setDetailView(DetailConflicts); m.activePanel = ConflictPanel }},
		{"rebase", func(m *Model) { m.setDetailView(DetailRebase); m.activePanel = RebasePanel }},
		{"config", func(m *Model) { m.activePanel = ConfigPanel }},
		{"commandlog", func(m *Model) { m.activePanel = CommandLogPanel }},
	}

	for _, size := range sizes {
		for _, v := range views {
			m := layoutModel(t, size[0], size[1])
			fillEveryList(m)
			v.setup(m)
			m.refreshViewports()

			frame := m.render()
			lines := strings.Split(frame, "\n")
			maxW := 0
			for _, l := range lines {
				if w := lipgloss.Width(l); w > maxW {
					maxW = w
				}
			}
			if maxW > size[0] || len(lines) > size[1] {
				t.Errorf("%s at %dx%d -> %d lines, max width %d",
					v.name, size[0], size[1], len(lines), maxW)
			}
			// The footer is the last thing joined, so anything that overflows
			// pushes it off the bottom of the frame.
			last := stripANSI(lines[len(lines)-1])
			if !strings.Contains(last, "? help") {
				t.Errorf("%s at %dx%d: footer missing, last line is %q",
					v.name, size[0], size[1], last)
			}
		}
	}
}

func fillEveryList(m *Model) {
	for i := range 40 {
		name := "file-" + strings.Repeat("x", i%30) + ".go"
		m.files = append(m.files, domain.FileStatus{Name: name, Modified: true})
		m.branches = append(m.branches, domain.BranchInfo{Name: "feature/branch-" + strings.Repeat("y", i%25), IsLocal: true})
		m.stashes = append(m.stashes, domain.StashInfo{Index: i, Message: "WIP " + strings.Repeat("z", i%30)})
		m.conflictFiles = append(m.conflictFiles, domain.ConflictFile{Name: name, Status: "both modified"})
		actions := []string{"pick", "squash", "fixup", "reword", "drop"}
		m.rebaseItems = append(m.rebaseItems, domain.RebaseItem{
			Hash: "abc1234", Action: actions[i%len(actions)],
			Message: "commit " + strings.Repeat("w", i%30),
		})
		m.commandLogs = append(m.commandLogs, CommandLogEntry{RepoName: "repo", Command: "fetch", Output: strings.Repeat("output ", i%10)})
	}
	m.currentDiff = strings.Repeat("+ a diff line that is quite long indeed and keeps going\n", 40)
	m.parsedDiff = parseUnifiedDiff(m.currentDiff)
	m.cachedLog = strings.Repeat("* abc1234 some commit subject line here\n", 40)
	m.cachedLogFor = "/tmp/repo"
	m.stashFiles = []string{"a.go", "b.go"}
	for i := range m.files {
		if i%3 == 0 {
			m.fileSelections[i] = true
		}
	}
}

func TestFooterPerPanel(t *testing.T) {
	m := layoutModel(t, 140, 44)
	m.activePanel = RepoPanel
	t.Logf("RepoPanel footer:  %s", stripANSI(m.renderFooter()))

	m.setDetailView(DetailLog)
	m.activePanel = LogPanel
	t.Logf("LogPanel footer:   %s", stripANSI(m.renderFooter()))

	m.setDetailView(DetailFiles)
	m.activePanel = DiffPanel
	t.Logf("DiffPanel footer:  %s", stripANSI(m.renderFooter()))
}

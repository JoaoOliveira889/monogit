package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoaoOliveira889/monogit/internal/domain"
)

func TestCherryPickDoneIsDispatched(t *testing.T) {
	m := Model{repos: []domain.Repository{{Name: "repo", Path: "/tmp", Committing: true}}}
	m.Update(cherryPickDoneMsg{index: 0, hash: "abc1234", output: "ok"})

	if m.repos[0].Committing {
		t.Error("cherry-pick left the repo marked as committing")
	}
	if m.statusMsg == "" {
		t.Error("cherry-pick produced no status message")
	}
}

func TestRevertDoneIsDispatched(t *testing.T) {
	m := Model{repos: []domain.Repository{{Name: "repo", Path: "/tmp", Committing: true}}}
	m.Update(revertDoneMsg{index: 0, hash: "abc1234", output: "ok"})

	if m.repos[0].Committing {
		t.Error("revert left the repo marked as committing")
	}
	if m.statusMsg == "" {
		t.Error("revert produced no status message")
	}
}

func TestSearchAcceptsJAndK(t *testing.T) {
	m := NewModel("/tmp", 0, nil)
	m.pushOverlay(OverlaySearch)
	m.searchInput.Focus()

	for _, r := range []rune{'j', 'k'} {
		m.handleSearchKeys(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}

	if got := m.searchInput.Value(); got != "jk" {
		t.Errorf("search input = %q, want %q", got, "jk")
	}
}

func TestPushAllReportsErrors(t *testing.T) {
	m := Model{repos: []domain.Repository{{Name: "a"}, {Name: "b"}}}
	m.Update(pushAllDoneMsg{results: []PushResult{
		{Index: 0, Name: "a"},
		{Index: 1, Name: "b", Err: errFake{}},
	}})

	if !strings.Contains(m.statusMsg, "1 errors") {
		t.Errorf("status = %q, want it to report the failure", m.statusMsg)
	}
}

func TestRebaseEnterAsksForConfirmation(t *testing.T) {
	m := NewModel("/tmp", 0, nil)
	m.repos = []domain.Repository{{Name: "repo", Path: "/tmp"}}
	m.setDetailView(DetailRebase)
	m.rebaseItems = []domain.RebaseItem{{Hash: "abc1234", Action: "pick", Message: "x"}}

	m.handleRebaseKeys(tea.KeyMsg{Type: tea.KeyEnter})

	if !m.showConfirmModal() {
		t.Fatal("enter executed the rebase without a confirmation modal")
	}
	if m.confirmModalAction != "execute_rebase" {
		t.Errorf("confirm action = %q, want %q", m.confirmModalAction, "execute_rebase")
	}
}

type errFake struct{}

func (errFake) Error() string { return "boom" }

// Search and tag assignment render inside the dashboard. If they were treated
// as frame-replacing overlays the view would go blank while they are open.
func TestInlineOverlaysKeepTheDashboardVisible(t *testing.T) {
	for _, tc := range []struct {
		name    string
		overlay Overlay
	}{
		{"search", OverlaySearch},
		{"tag assign", OverlayTagAssign},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := NewModel("/tmp", 0, nil)
			m.showSplash = false
			m.repos = []domain.Repository{{Name: "repo", Path: "/tmp/repo", Branch: "main"}}
			m.invalidateFilterCache()
			m.handleResize(tea.WindowSizeMsg{Width: 120, Height: 40})
			m.pushOverlay(tc.overlay)

			view := m.View()

			if !strings.Contains(view, "repo") {
				t.Errorf("dashboard not rendered while %s is open:\n%s", tc.name, view)
			}
		})
	}
}

// Overlays stack: the innermost one owns the keyboard, and closing it hands
// control back to the one beneath.
func TestOverlayStackUnwinds(t *testing.T) {
	m := NewModel("/tmp", 0, nil)
	m.pushOverlay(OverlayTagAssign)
	m.pushOverlay(OverlayInput)

	if top, _ := m.topOverlay(); top != OverlayInput {
		t.Fatalf("top overlay = %v, want OverlayInput", top)
	}

	m.popOverlay()

	if top, ok := m.topOverlay(); !ok || top != OverlayTagAssign {
		t.Fatalf("top overlay after pop = %v (open=%v), want OverlayTagAssign", top, ok)
	}
	if !m.tagAssignModal() || m.inputMode() {
		t.Error("accessors disagree with the stack")
	}
}

func TestCountedMotionMovesByCount(t *testing.T) {
	m := NewModel("/tmp", 0, nil)
	m.showSplash = false
	for i := range 10 {
		m.repos = append(m.repos, domain.Repository{
			Name: string(rune('a' + i)),
			Path: "/tmp/" + string(rune('a'+i)),
		})
	}
	m.invalidateFilterCache()
	m.handleResize(tea.WindowSizeMsg{Width: 120, Height: 40})

	// "4j"
	m.handleNormalKeys(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("4")})
	if m.cursor != 0 {
		t.Fatalf("the count digit moved the cursor to %d; it should only be buffered", m.cursor)
	}
	m.handleNormalKeys(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})

	if m.cursor != 4 {
		t.Errorf("cursor = %d after 4j, want 4", m.cursor)
	}
	if !m.pending.empty() {
		t.Error("the pending count survived the motion")
	}
}

func TestGPrefixSequences(t *testing.T) {
	m := NewModel("/tmp", 0, nil)
	m.showSplash = false
	m.repos = []domain.Repository{{Name: "a", Path: "/a"}, {Name: "b", Path: "/b"}}
	m.invalidateFilterCache()
	m.handleResize(tea.WindowSizeMsg{Width: 120, Height: 40})
	m.cursor = 1

	// A lone "g" is a prefix and must not act on its own.
	m.handleNormalKeys(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")})
	if m.pending.prefix != "g" {
		t.Fatalf("pending prefix = %q, want \"g\"", m.pending.prefix)
	}
	if m.cursor != 1 {
		t.Error("a lone g moved the cursor")
	}

	m.handleNormalKeys(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")})
	if m.cursor != 0 {
		t.Errorf("cursor = %d after gg, want 0", m.cursor)
	}

	// "gl" toggles the graph, which "g" alone used to do.
	before := m.viewGraph
	m.handleNormalKeys(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")})
	m.handleNormalKeys(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
	if m.viewGraph == before {
		t.Error("gl did not toggle the graph view")
	}
}

func TestJumpToAttentionSkipsCleanRepositories(t *testing.T) {
	m := NewModel("/tmp", 0, nil)
	m.showSplash = false
	m.repos = []domain.Repository{
		{Name: "clean-a", Path: "/a"},
		{Name: "clean-b", Path: "/b"},
		{Name: "dirty-c", Path: "/c", IsDirty: true},
		{Name: "clean-d", Path: "/d"},
		{Name: "behind-e", Path: "/e", Behind: 2},
	}
	m.invalidateFilterCache()
	m.handleResize(tea.WindowSizeMsg{Width: 120, Height: 40})

	m.jumpToAttention(1, 1)
	if m.cursor != 2 {
		t.Errorf("cursor = %d, want the dirty repository at 2", m.cursor)
	}

	m.jumpToAttention(1, 1)
	if m.cursor != 4 {
		t.Errorf("cursor = %d, want the behind repository at 4", m.cursor)
	}

	// Wraps around rather than stopping at the end.
	m.jumpToAttention(1, 1)
	if m.cursor != 2 {
		t.Errorf("cursor = %d after wrapping, want 2", m.cursor)
	}
}

func TestJumpListReturnsToPreviousRepository(t *testing.T) {
	m := NewModel("/tmp", 0, nil)
	m.showSplash = false
	for i := range 20 {
		m.repos = append(m.repos, domain.Repository{
			Name: string(rune('a' + i)),
			Path: "/tmp/" + string(rune('a'+i)),
		})
	}
	m.invalidateFilterCache()
	m.handleResize(tea.WindowSizeMsg{Width: 120, Height: 40})

	m.handleCursorMove(10)
	origin := m.cursor
	m.handleCursorMove(-8)

	if m.cursor == origin {
		t.Fatal("the second motion did not move the cursor")
	}

	m.jumpTo(-1)
	if m.cursor != origin {
		t.Errorf("cursor = %d after ctrl+o, want %d", m.cursor, origin)
	}
}

func TestRelativeNumberGutter(t *testing.T) {
	m := NewModel("/tmp", 0, nil)
	m.cfg.RelativeNumber = false
	if got := m.repoGutter(3); got != "" {
		t.Errorf("gutter = %q with relative numbers off, want empty", got)
	}

	m.cfg.RelativeNumber = true
	m.cursor = 5

	if got := stripANSI(m.repoGutter(5)); strings.TrimSpace(got) != "6" {
		t.Errorf("cursor row = %q, want the absolute number 6", got)
	}
	if got := stripANSI(m.repoGutter(9)); strings.TrimSpace(got) != "4" {
		t.Errorf("row 9 = %q, want distance 4", got)
	}
	if got := stripANSI(m.repoGutter(1)); strings.TrimSpace(got) != "4" {
		t.Errorf("row 1 = %q, want distance 4", got)
	}
}

func TestPaletteRunsCommandByPrefix(t *testing.T) {
	m := NewModel("/tmp", 0, nil)
	m.showSplash = false
	m.repos = []domain.Repository{
		{Name: "clean", Path: "/a"},
		{Name: "dirty", Path: "/b", IsDirty: true},
	}
	m.invalidateFilterCache()

	m.runPaletteCommand(":filter dirty")

	if m.statusFilter != FilterDirty {
		t.Errorf("statusFilter = %v, want FilterDirty", m.statusFilter)
	}
	if got := len(m.filteredRepos()); got != 1 {
		t.Errorf("filtered repositories = %d, want 1", got)
	}
}

func TestPaletteRejectsUnknownCommand(t *testing.T) {
	m := NewModel("/tmp", 0, nil)

	m.runPaletteCommand(":nope")

	if !strings.Contains(m.statusMsg, "Unknown command") {
		t.Errorf("status = %q, want an unknown-command message", m.statusMsg)
	}
}

func TestPaletteReportsAmbiguousPrefix(t *testing.T) {
	m := NewModel("/tmp", 0, nil)

	// "p" matches pull-all, push-all.
	m.runPaletteCommand(":p")

	if !strings.Contains(m.statusMsg, "Ambiguous") {
		t.Errorf("status = %q, want an ambiguity message", m.statusMsg)
	}
}

func TestPaletteTabCompletes(t *testing.T) {
	m := NewModel("/tmp", 0, nil)
	m.openPalette()
	m.paletteInput.SetValue("relative")

	m.handlePaletteKeys(tea.KeyMsg{Type: tea.KeyTab})

	if got := m.paletteInput.Value(); !strings.HasPrefix(got, "relativenumber") {
		t.Errorf("completed to %q, want relativenumber", got)
	}
}

func TestPaletteEscapeClosesWithoutRunning(t *testing.T) {
	m := NewModel("/tmp", 0, nil)
	m.openPalette()
	m.paletteInput.SetValue("quit")

	m.handlePaletteKeys(tea.KeyMsg{Type: tea.KeyEsc})

	if m.paletteOpen() {
		t.Error("escape left the palette open")
	}
	if m.quitting {
		t.Error("escape ran the highlighted command")
	}
}

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

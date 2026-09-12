package tui

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/JoaoOliveira889/monogit/internal/domain"
)

func TestRebaseKeyHandlers(t *testing.T) {
	m := &Model{
		detailView:  DetailRebase,
		activePanel: RebasePanel,
		rebaseItems: []domain.RebaseItem{
			{Hash: "111", Action: "pick", Message: "Commit 1"},
			{Hash: "222", Action: "pick", Message: "Commit 2"},
			{Hash: "333", Action: "pick", Message: "Commit 3"},
		},
		rebaseCursor: 0,
	}

	// Move cursor down
	m.handleRebaseKeys(tea.KeyPressMsg{Code: 'j', Text: "j"})
	if m.rebaseCursor != 1 {
		t.Fatalf("expected rebaseCursor 1, got %d", m.rebaseCursor)
	}

	// Change action to squash
	m.handleRebaseKeys(tea.KeyPressMsg{Code: 's', Text: "s"})
	if m.rebaseItems[1].Action != "squash" {
		t.Fatalf("expected Action squash, got %s", m.rebaseItems[1].Action)
	}

	// Change action to fixup
	m.handleRebaseKeys(tea.KeyPressMsg{Code: 'f', Text: "f"})
	if m.rebaseItems[1].Action != "fixup" {
		t.Fatalf("expected Action fixup, got %s", m.rebaseItems[1].Action)
	}

	// Move item down (reorder)
	m.handleRebaseKeys(tea.KeyPressMsg{Code: 'J', Text: "J"})
	if m.rebaseItems[2].Hash != "222" {
		t.Fatalf("expected hash 222 at index 2 after moving down, got %s", m.rebaseItems[2].Hash)
	}

	// Escape cancels rebase
	m.handleRebaseKeys(tea.KeyPressMsg{Code: tea.KeyEscape})
	if m.showRebase() {
		t.Fatalf("expected showRebase false after Esc")
	}
}

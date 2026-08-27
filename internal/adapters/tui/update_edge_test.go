package tui

import (
	"testing"
	"time"

	"github.com/JoaoOliveira889/monogit/internal/domain"
	tea "github.com/charmbracelet/bubbletea"
)

// --- Sprint 1.1 / 1.2: branchCursor and fileCursor bounds ---

func TestDeleteBranchLocalWithStaleCursor(t *testing.T) {
	m := mkModel()
	m.repos = []domain.Repository{{Name: "r1", Path: "/p1"}}
	m.branches = []domain.BranchInfo{{Name: "main"}}
	m.branchCursor = 5 // stale cursor beyond slice bounds

	// Should not panic.
	res, _ := m.executeConfirmedAction("delete_branch_local")
	if res == nil {
		t.Fatal("expected non-nil model")
	}
}

func TestDeleteBranchWorktreeWithStaleCursor(t *testing.T) {
	m := mkModel()
	m.repos = []domain.Repository{{Name: "r1", Path: "/p1"}}
	m.branches = []domain.BranchInfo{{Name: "main"}}
	m.branchCursor = 5

	res, _ := m.executeConfirmedAction("delete_branch_worktree")
	if res == nil {
		t.Fatal("expected non-nil model")
	}
}

func TestDeleteBranchRemoteWithStaleCursor(t *testing.T) {
	m := mkModel()
	m.repos = []domain.Repository{{Name: "r1", Path: "/p1"}}
	m.branches = []domain.BranchInfo{{Name: "main"}}
	m.branchCursor = 5

	res, _ := m.executeConfirmedAction("delete_branch_remote")
	if res == nil {
		t.Fatal("expected non-nil model")
	}
}

func TestDeleteBranchKeyWithStaleCursor(t *testing.T) {
	m := mkModel()
	m.repos = []domain.Repository{{Name: "r1", Path: "/p1"}}
	m.cursor = 0
	m.activePanel = LogPanel
	m.showBranches = true
	m.branches = []domain.BranchInfo{{Name: "feature"}}
	m.branchCursor = 10

	// Should not panic — the case guard should prevent access.
	res, _ := m.handleNormalKeys(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})
	if res == nil {
		t.Fatal("expected non-nil model")
	}
}

func TestMergeKeyWithStaleBranchCursor(t *testing.T) {
	m := mkModel()
	m.repos = []domain.Repository{{Name: "r1", Path: "/p1"}}
	m.cursor = 0
	m.showBranches = true
	m.branches = []domain.BranchInfo{{Name: "dev"}}
	m.branchCursor = 10

	// Should not panic.
	res, _ := m.handleNormalKeys(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("M")})
	if res == nil {
		t.Fatal("expected non-nil model")
	}
}

func TestDiscardKeyWithStaleFileCursor(t *testing.T) {
	m := mkModel()
	m.repos = []domain.Repository{{Name: "r1", Path: "/p1"}}
	m.showFiles = true
	m.files = []domain.FileStatus{{Name: "a.go"}}
	m.fileCursor = 10

	// Should not panic.
	res, _ := m.handleNormalKeys(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	if res == nil {
		t.Fatal("expected non-nil model")
	}
}

func TestDeleteBranchWithEmptyBranches(t *testing.T) {
	m := mkModel()
	m.repos = []domain.Repository{{Name: "r1", Path: "/p1"}}
	m.cursor = 0
	m.activePanel = LogPanel
	m.showBranches = true
	m.branches = nil
	m.branchCursor = 0

	res, _ := m.handleNormalKeys(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})
	updated := res.(*Model)
	if updated.showConfirmModal {
		t.Fatal("expected no confirmation modal when branches list is empty")
	}
}

// --- Sprint 1.3: m.repos[m.cursor] bounds ---

func TestCursorMoveOnEmptyRepos(t *testing.T) {
	m := mkModel()
	m.repos = nil
	m.cursor = 0
	m.activePanel = RepoPanel

	// Should not panic.
	res, _ := m.handleCursorMove(1)
	if res == nil {
		t.Fatal("expected non-nil model")
	}
	res, _ = m.handleCursorMove(-1)
	if res == nil {
		t.Fatal("expected non-nil model")
	}
}

func TestCursorMoveWithCursorBeyondReposLength(t *testing.T) {
	m := mkModel()
	m.repos = []domain.Repository{{Name: "r1", Path: "/p1"}}
	m.cursor = 50 // way beyond len(repos)
	m.activePanel = RepoPanel

	// Should not panic.
	res, _ := m.handleCursorMove(1)
	if res == nil {
		t.Fatal("expected non-nil model")
	}
}

// --- Sprint 1.4: handleEnterKey with branches ---

func TestEnterKeyOnBranchesWithStaleCursor(t *testing.T) {
	m := mkModel()
	m.repos = []domain.Repository{{Name: "r1", Path: "/p1"}}
	m.cursor = 0
	m.activePanel = LogPanel
	m.showBranches = true
	m.branches = []domain.BranchInfo{{Name: "main"}}
	m.branchCursor = 5

	// Should not panic.
	res, _ := m.handleEnterKey()
	if res == nil {
		t.Fatal("expected non-nil model")
	}
}

func TestEnterKeyOnStashesWithStaleCursor(t *testing.T) {
	m := mkModel()
	m.repos = []domain.Repository{{Name: "r1", Path: "/p1"}}
	m.cursor = 0
	m.activePanel = LogPanel
	m.showStashes = true
	m.stashes = []domain.StashInfo{{Index: 0}}
	m.stashCursor = 5

	// Should not panic.
	res, _ := m.handleEnterKey()
	if res == nil {
		t.Fatal("expected non-nil model")
	}
}

// --- Sprint 3.1: Splash skip on keypress ---

func TestSplashSkipOnKeypress(t *testing.T) {
	m := mkModel()
	m.showSplash = true
	m.splashReady = true
	m.splashStartedAt = time.Now().Add(-2 * time.Second)

	res, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	updated := res.(*Model)

	if updated.showSplash {
		t.Fatal("expected splash to be hidden after keypress when splashReady is true")
	}
}

func TestSplashNotSkippedBeforeReady(t *testing.T) {
	m := mkModel()
	m.showSplash = true
	m.splashReady = false

	res, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	updated := res.(*Model)

	if !updated.showSplash {
		t.Fatal("expected splash to remain visible when splashReady is false")
	}
}

// --- Sprint 3.2: Config save success feedback ---

func TestConfigSaveSuccessFeedback(t *testing.T) {
	m := mkModel()

	res, _ := m.Update(configSavedMsg{err: nil})
	updated := res.(*Model)

	if updated.statusMsg == "" {
		t.Fatal("expected non-empty status message on config save success")
	}
	if updated.statusMsg != "✓ Config saved" {
		t.Fatalf("expected '✓ Config saved', got %q", updated.statusMsg)
	}
}

// --- view_panels.go: syncScrollPositions bounds ---

func TestSyncScrollPositionsWithEmptyRepos(t *testing.T) {
	m := mkModel()
	m.repos = nil
	m.cursor = 0
	m.repoViewport.Height = 10

	// Should not panic.
	m.syncScrollPositions()
}

func TestSyncScrollPositionsWithCursorOutOfBounds(t *testing.T) {
	m := mkModel()
	m.repos = []domain.Repository{{Name: "r1", Path: "/p1"}}
	m.cursor = 50
	m.repoViewport.Height = 10

	// Should not panic.
	m.syncScrollPositions()
}

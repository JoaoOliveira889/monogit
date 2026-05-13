package tui

import (
	"monogit/internal/domain"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestHandleResize(t *testing.T) {
	m := mkModel()
	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	result, cmd := m.handleResize(msg)
	if result == nil {
		t.Fatal("expected non-nil model")
	}
	if cmd != nil {
		t.Fatal("expected nil cmd from resize")
	}
	if m.width != 100 || m.height != 50 {
		t.Errorf("expected 100x50, got %dx%d", m.width, m.height)
	}
}

func TestHandleRepoScanned(t *testing.T) {
	m := mkModel()
	repos := []domain.Repository{
		{Name: "r1", Path: "/p1"},
		{Name: "r2", Path: "/p2"},
	}
	msg := repoScannedMsg{repos: repos}
	_, cmd := m.handleRepoScanned(msg)

	if len(m.repos) != 2 {
		t.Errorf("expected 2 repos, got %d", len(m.repos))
	}
	if m.scanning {
		t.Error("expected scanning to be false after scan")
	}
	if cmd == nil {
		t.Fatal("expected non-nil command from handleRepoScanned")
	}
}

func TestHandleRepoStatus(t *testing.T) {
	m := mkModel()
	m.repos = []domain.Repository{{Name: "r1", Path: "/p1"}}

	msg := repoStatusMsg{
		index:  0,
		branch: "main",
		ahead:  1,
		behind: 2,
		dirty:  true,
	}

	_, cmd := m.handleRepoStatus(msg)
	if cmd != nil {
		t.Fatal("expected nil cmd from status update")
	}
	if m.repos[0].Branch != "main" {
		t.Errorf("expected branch main, got %s", m.repos[0].Branch)
	}
	if m.repos[0].Ahead != 1 || m.repos[0].Behind != 2 {
		t.Errorf("expected ahead=1, behind=2, got %d, %d", m.repos[0].Ahead, m.repos[0].Behind)
	}
	if !m.repos[0].IsDirty {
		t.Error("expected dirty")
	}
}

func TestHandleRepoStatusError(t *testing.T) {
	m := mkModel()
	m.repos = []domain.Repository{{Name: "r1", Path: "/p1"}}

	msg := repoStatusMsg{
		index: 0,
		err:   &testError{"git error"},
	}

	m.handleRepoStatus(msg)
	if m.repos[0].Error == "" {
		t.Error("expected error message on repo")
	}
}

type testError struct {
	msg string
}

func (e *testError) Error() string { return e.msg }

func TestHandleRepoStatusRefresh(t *testing.T) {
	m := mkModel()
	m.repos = []domain.Repository{{Name: "r1", Path: "/p1"}}

	msg := repoStatusMsg{
		index:   0,
		refresh: true,
	}

	_, cmd := m.handleRepoStatus(msg)
	if cmd == nil {
		t.Fatal("expected non-nil cmd for refresh status")
	}
}

func TestHandleFetchDone(t *testing.T) {
	m := mkModel()
	m.repos = []domain.Repository{{Name: "r1", Path: "/p1", Fetching: true}}

	msg := fetchDoneMsg{index: 0}
	_, cmd := m.handleFetchDone(msg)
	if cmd == nil {
		t.Fatal("expected non-nil cmd after fetch done")
	}
	if m.repos[0].Fetching {
		t.Error("expected Fetching to be false after fetch done")
	}
}

func TestHandleFetchDoneAll(t *testing.T) {
	m := mkModel()
	m.repos = []domain.Repository{
		{Name: "r1", Path: "/p1", Fetching: true},
		{Name: "r2", Path: "/p2", Fetching: true},
	}

	msg := fetchDoneMsg{all: true}
	_, cmd := m.handleFetchDone(msg)
	if cmd == nil {
		t.Fatal("expected non-nil cmd after all fetch done")
	}
	for _, r := range m.repos {
		if r.Fetching {
			t.Error("expected all Fetching to be false after all fetch")
		}
	}
}

func TestHandleGitBranches(t *testing.T) {
	m := mkModel()
	branches := []domain.BranchInfo{
		{Name: "main", IsLocal: true, IsCurrent: true},
		{Name: "develop", IsLocal: true},
	}
	msg := gitBranchesMsg{branches: branches}

	_, cmd := m.handleGitBranches(msg)
	if cmd != nil {
		t.Fatal("expected nil cmd from git branches")
	}
	if !m.showBranches {
		t.Error("expected showBranches to be true")
	}
	if len(m.branches) != 2 {
		t.Errorf("expected 2 branches, got %d", len(m.branches))
	}
}

func TestHandleNextStepMsg(t *testing.T) {
	m := mkModel()
	m.commitStep = StepMessage
	m.repos = []domain.Repository{{Name: "r1", Path: "/p1"}}

	_, cmd := m.handleNextStepMsg()
	if cmd == nil {
		t.Fatal("expected non-nil cmd for next step")
	}
	if m.inputAction != "commit" {
		t.Errorf("expected inputAction 'commit', got %s", m.inputAction)
	}
}

func TestHandleErrMsg(t *testing.T) {
	m := mkModel()
	em := errMsg{Err: &testError{"something broke"}}

	m.statusMsg = ""
	m.Update(em)
	if m.statusMsg == "" {
		t.Error("expected status msg to be set after error")
	}
}

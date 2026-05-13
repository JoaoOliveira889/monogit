package tui

import (
	"monogit/internal/domain"
	"monogit/internal/usecase"
	"testing"
	"time"
)

type mockGitProvider struct {
	getFilesFunc     func(string) ([]domain.FileStatus, error)
	getSimpleLogFunc func(string, int) (string, error)
}

func (m *mockGitProvider) GetBranch(repoPath string) (string, error)          { return "", nil }
func (m *mockGitProvider) IsDirty(repoPath string) (bool, error)              { return false, nil }
func (m *mockGitProvider) GetAheadBehind(repoPath string) (int, int, error)   { return 0, 0, nil }
func (m *mockGitProvider) FetchAll(repoPath string) error                     { return nil }
func (m *mockGitProvider) Pull(repoPath string) (string, error)              { return "", nil }
func (m *mockGitProvider) Push(repoPath string) (string, error)              { return "", nil }
func (m *mockGitProvider) AddAndCommit(repoPath, message string) (string, error) { return "", nil }
func (m *mockGitProvider) DiscardChanges(repoPath string, f domain.FileStatus) error { return nil }
func (m *mockGitProvider) GetBranches(repoPath string) ([]domain.BranchInfo, error) { return nil, nil }
func (m *mockGitProvider) CheckoutBranch(repoPath, name string) error        { return nil }
func (m *mockGitProvider) CreateBranch(repoPath, name string) error          { return nil }
func (m *mockGitProvider) DeleteBranch(repoPath, name string) (string, error) { return "", nil }
func (m *mockGitProvider) DeleteRemoteBranch(repoPath, remote, name string) (string, error) { return "", nil }
func (m *mockGitProvider) Stash(repoPath, message string) (string, error)    { return "", nil }
func (m *mockGitProvider) StashPop(repoPath string) (string, error)          { return "", nil }
func (m *mockGitProvider) UnstageAll(repoPath string) error                  { return nil }
func (m *mockGitProvider) UnstageFile(repoPath, fileName string) error       { return nil }
func (m *mockGitProvider) UndoCommit(repoPath string) error                  { return nil }
func (m *mockGitProvider) StageByPattern(repoPath, pattern string) error     { return nil }
func (m *mockGitProvider) GetGraphLog(repoPath string, n int) (string, error) { return "", nil }
func (m *mockGitProvider) GetStatusFiles(p string) ([]domain.FileStatus, error) {
	if m.getFilesFunc != nil {
		return m.getFilesFunc(p)
	}
	return nil, nil
}
func (m *mockGitProvider) GetSimpleLog(p string, n int) (string, error) {
	if m.getSimpleLogFunc != nil {
		return m.getSimpleLogFunc(p, n)
	}
	return "", nil
}
func (m *mockGitProvider) GetDiff(p string, f domain.FileStatus) (string, error) { return "", nil }

func mkModel() Model {
	uc := usecase.NewGitUseCase(&mockGitProvider{})
	return NewModel("/root", 30*time.Second, uc)
}

func TestNewModel(t *testing.T) {
	uc := usecase.NewGitUseCase(&mockGitProvider{})
	m := NewModel("/root", 30*time.Second, uc)
	if m.rootPath != "/root" {
		t.Errorf("expected /root, got %s", m.rootPath)
	}
	if m.fetchInterval != 30*time.Second {
		t.Errorf("expected 30s interval, got %v", m.fetchInterval)
	}
	if !m.scanning {
		t.Error("expected scanning to be true on init")
	}
	if m.activePanel != RepoPanel {
		t.Errorf("expected RepoPanel, got %v", m.activePanel)
	}
}

func TestSelectedRepo_Empty(t *testing.T) {
	m := mkModel()
	if r := m.selectedRepo(); r != nil {
		t.Error("expected nil for empty repos")
	}
}

func TestSelectedRepo_WithRepos(t *testing.T) {
	m := mkModel()
	m.repos = []domain.Repository{{Name: "repo1", Path: "/p1"}, {Name: "repo2", Path: "/p2"}}
	m.cursor = 0

	r := m.selectedRepo()
	if r == nil {
		t.Fatal("expected non-nil")
	}
	if r.Name != "repo1" {
		t.Errorf("expected repo1, got %s", r.Name)
	}

	m.cursor = 1
	r = m.selectedRepo()
	if r.Name != "repo2" {
		t.Errorf("expected repo2, got %s", r.Name)
	}

	m.cursor = 5
	if r := m.selectedRepo(); r != nil {
		t.Error("expected nil for out-of-range cursor")
	}
}

func TestLeftPanelWidth(t *testing.T) {
	m := mkModel()
	m.width = 100
	w := m.leftPanelWidth()
	if w != 30 {
		t.Errorf("expected 30 (30%% of 100), got %d", w)
	}

	m.width = 40
	w = m.leftPanelWidth()
	if w < 24 {
		t.Errorf("expected minimum 24, got %d", w)
	}
}

func TestRightPanelWidth(t *testing.T) {
	m := mkModel()
	m.width = 100
	w := m.rightPanelWidth()
	expected := 100 - m.leftPanelWidth() - 4
	if w != expected {
		t.Errorf("expected %d, got %d", expected, w)
	}
}

func TestPanelHeight(t *testing.T) {
	m := mkModel()
	m.height = 50
	h := m.panelHeight()
	expected := 50 - 4
	if h != expected {
		t.Errorf("expected %d, got %d", expected, h)
	}

	m.height = 6
	h = m.panelHeight()
	if h < 5 {
		t.Errorf("expected minimum 5, got %d", h)
	}
}

func TestGetStagedFiles(t *testing.T) {
	m := mkModel()
	m.files = []domain.FileStatus{
		{Name: "a.go", Staged: true},
		{Name: "b.go"},
		{Name: "c.go", Staged: true},
	}
	staged := m.getStagedFiles()
	if len(staged) != 2 {
		t.Errorf("expected 2 staged, got %d", len(staged))
	}
	if staged[0] != "a.go" || staged[1] != "c.go" {
		t.Errorf("unexpected staged files: %v", staged)
	}
}

func TestIsBusy(t *testing.T) {
	m := mkModel()
	m.scanning = false
	if m.isBusy() {
		t.Error("expected not busy when not scanning")
	}

	m.scanning = true
	if !m.isBusy() {
		t.Error("expected busy when scanning")
	}

	m.scanning = false
	m.diffFetching = true
	if !m.isBusy() {
		t.Error("expected busy when diffFetching")
	}

	m.diffFetching = false
	m.repos = []domain.Repository{{Name: "r", Fetching: true}}
	if !m.isBusy() {
		t.Error("expected busy when repo fetching")
	}

	m.repos[0].Fetching = false
	m.repos[0].Pulling = true
	if !m.isBusy() {
		t.Error("expected busy when repo pulling")
	}
}

func TestCancelSpecialModes(t *testing.T) {
	m := mkModel()
	m.showFiles = true
	m.showBranches = true
	m.inputMode = true
	m.showHelp = true
	m.currentDiff = "diff"
	m.fileSelections[0] = true
	m.statusMsg = "msg"

	m.cancelSpecialModes()

	if m.showFiles || m.showBranches || m.inputMode || m.showHelp {
		t.Error("all modes should be cancelled")
	}
	if m.currentDiff != "" {
		t.Error("diff should be cleared")
	}
	if len(m.fileSelections) != 0 {
		t.Error("selections should be cleared")
	}
	if m.statusMsg != "" {
		t.Error("status message should be cleared")
	}
}

func TestRefreshCachedRepoDetail(t *testing.T) {
	m := mkModel()
	m.gitUC = usecase.NewGitUseCase(&mockGitProvider{})

	m.cachedModifiedCount = 10
	m.cachedUntrackedCount = 10
	m.cachedLastCommit = "old"

	m.refreshCachedRepoDetail()
	if m.cachedModifiedCount != 0 || m.cachedUntrackedCount != 0 || m.cachedLastCommit != "" {
		t.Error("expected zero cached detail when no repo selected")
	}

	m.repos = []domain.Repository{{Name: "r", Path: "/p"}}
	m.cursor = 0
	m.gitUC = nil
	m.refreshCachedRepoDetail()
}

func TestGetVisiblePanels(t *testing.T) {
	m := mkModel()

	panels := m.GetVisiblePanels()
	if len(panels) != 2 || panels[0] != RepoPanel || panels[1] != LogPanel {
		t.Errorf("expected [RepoPanel LogPanel], got %v", panels)
	}

	m.activePanel = CommandLogPanel
	panels = m.GetVisiblePanels()
	if len(panels) != 2 || panels[1] != CommandLogPanel {
		t.Errorf("expected [RepoPanel CommandLogPanel], got %v", panels)
	}

	m.showFiles = true
	m.activePanel = LogPanel
	panels = m.GetVisiblePanels()
	if len(panels) != 3 || panels[1] != LogPanel || panels[2] != DiffPanel {
		t.Errorf("expected [RepoPanel LogPanel DiffPanel], got %v", panels)
	}

	m.showFiles = false
	m.showBranches = true
	panels = m.GetVisiblePanels()
	if len(panels) != 2 || panels[1] != LogPanel {
		t.Errorf("expected [RepoPanel LogPanel], got %v", panels)
	}
}

func TestSpinnerView(t *testing.T) {
	m := mkModel()
	m.spinnerFrame = 0
	s := m.spinnerView()
	if s == "" {
		t.Error("expected non-empty spinner")
	}

	m.spinnerFrame = 10
	s2 := m.spinnerView()
	if s == s2 {
		t.Error("expected different spinner at different frames")
	}
}

package tui

import (
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/JoaoOliveira889/monogit/internal/adapters/git"
	"github.com/JoaoOliveira889/monogit/internal/domain"
	"github.com/JoaoOliveira889/monogit/internal/pkg/logging"
	"github.com/JoaoOliveira889/monogit/internal/pkg/ui"
	"github.com/JoaoOliveira889/monogit/internal/testutil"
	"github.com/JoaoOliveira889/monogit/internal/usecase"
)

func mkModel() Model {
	uc := usecase.NewGitUseCase(&testutil.MockGitProvider{})
	return NewModel("/root", 30*time.Second, uc)
}

func TestNewModel(t *testing.T) {
	uc := usecase.NewGitUseCase(&testutil.MockGitProvider{})
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
	m.leftPanelRatio = 0.3
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
	expected := 100 - m.leftPanelWidth()
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

func TestSelectedFiles(t *testing.T) {
	m := mkModel()
	m.files = []domain.FileStatus{
		{Name: "a.go"},
		{Name: "b.go"},
		{Name: "c.go"},
	}
	m.fileSelections[0] = true
	m.fileSelections[2] = true
	selected := m.selectedFiles()
	if len(selected) != 2 {
		t.Errorf("expected 2 selected, got %d", len(selected))
	}
	if selected[0] != "a.go" || selected[1] != "c.go" {
		t.Errorf("unexpected selected files: %v", selected)
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

func TestRenderTitledPanelTruncation(t *testing.T) {
	m := mkModel()
	title := "This is a super extremely incredibly long panel title"
	result := m.renderTitledPanel(30, 10, title, "content", false, ui.ColorMono)

	expectedTitlePart := "This is a super extre..."
	if !strings.Contains(result, expectedTitlePart) {
		t.Errorf("expected truncated title part %q in output, got:\n%s", expectedTitlePart, result)
	}
}

func TestRenderTitledPanelTruncationRightAccent(t *testing.T) {
	m := mkModel()
	title := "Short title"
	result := m.renderTitledPanel(30, 10, title, "content", true, ui.ColorGit)

	if !strings.Contains(result, "Short title") {
		t.Fatalf("expected title to be rendered, got:\n%s", result)
	}
}

func TestRenderRepoLineBranchName(t *testing.T) {
	m := mkModel()
	r := domain.Repository{Name: "test-repo", Path: "/p1", Branch: "feature/super-cool-stuff"}

	result := m.renderRepoLine(0, r, 60)

	if !strings.Contains(result, "test-repo") {
		t.Error("expected repo name 'test-repo' in repo line output")
	}
	if !strings.Contains(result, "feature/super-cool-stuff") {
		t.Errorf("expected branch name 'feature/super-cool-stuff' in repo line output, got:\n%s", result)
	}
}

func TestRenderRepoLineBranchNameTruncation(t *testing.T) {
	m := mkModel()
	r := domain.Repository{Name: "my-repo", Path: "/p1", Branch: "feature-branch", HasUpstream: true}

	result := m.renderRepoLine(0, r, 25)

	if !strings.Contains(result, "branch)") {
		t.Errorf("expected branch suffix to be preserved at the end, got:\n%s", result)
	}
}

func TestRenderRepoLineHealthBadges(t *testing.T) {
	m := mkModel()
	r := domain.Repository{
		Name:           "repo",
		Path:           "/p1",
		Branch:         "feature/x",
		HasUpstream:    false,
		HasConflicts:   true,
		HasUnpushedTag: true,
	}

	result := m.renderRepoLine(0, r, 80)
	for _, want := range []string{"UP", "CF", "TG"} {
		if !strings.Contains(result, want) {
			t.Fatalf("expected %q badge in repo line, got:\n%s", want, result)
		}
	}
}

func TestRenderBeautifiedLogDebug(t *testing.T) {
	m := mkModel()
	logInput := `* 4b732e0||HEAD -> main, tag: v0.4.6, origin/main, origin/HEAD||feat: a method created to perform P2P transfers using a document and account number. (#51)||33 hours ago||Matheus Medeiros Oselame
* 5b86e07||tag: v0.4.5||feat: add hire_date on tenant proto (#50)||4 days ago||Joao Oliveira
* 7cb48ab||tag: v0.4.4||feat: add new values to return on Holder Response (#49)||6 days ago||Joao Oliveira`

	res := m.renderBeautifiedLog(logInput, 80)
	t.Logf("Beautified log result:\n%s", res)
	if res == "" {
		t.Fatal("expected non-empty result")
	}
}

func TestRealLogSnapshotUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	logging.Init()
	defer logging.Close()

	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not found")
	}

	tmpDir := t.TempDir()
	cmd := exec.Command("git", "-C", tmpDir, "init")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Skipf("git init failed: %s", string(out))
	}
	cmd = exec.Command("git", "-C", tmpDir, "config", "user.email", "test@test.com")
	cmd.Run()
	cmd = exec.Command("git", "-C", tmpDir, "config", "user.name", "Test")
	cmd.Run()
	os.WriteFile(tmpDir+"/readme.md", []byte("# test"), 0644)
	cmd = exec.Command("git", "-C", tmpDir, "add", ".")
	cmd.Run()
	commitCmd := exec.Command("git", "-C", tmpDir, "commit", "-m", "initial")
	if out, err := commitCmd.CombinedOutput(); err != nil {
		t.Skipf("git commit failed: %s", string(out))
	}

	realGit := git.NewGitCLIAdapter()
	uc := usecase.NewGitUseCase(realGit)
	m := NewModel(tmpDir, 30*time.Second, uc)

	m.repos = []domain.Repository{
		{Name: "test-repo", Path: tmpDir},
	}
	m.cursor = 0

	cmdQuick := m.refreshQuickSnapshotCmd(0, tmpDir)
	msgQuick := cmdQuick()
	resModel, _ := m.Update(msgQuick)
	m2 := resModel.(*Model)
	t.Logf("Quick snapshot: loading=%v", m2.detailLoading)

	cmdLog := m2.refreshLogSnapshotCmd(0, tmpDir, true)
	msgLog := cmdLog()
	resModel2, _ := m2.Update(msgLog)
	m3 := resModel2.(*Model)
	t.Logf("Log snapshot: logFor=%q, logLen=%d, loading=%v",
		m3.cachedLogFor, len(m3.cachedLog), m3.detailLoading)

	detailView := m3.renderDetailPanel(80, 24)
	if strings.Contains(detailView, "Loading commit history...") && m3.cachedLog != "" {
		t.Error("log is cached but view still shows loading")
	}
}

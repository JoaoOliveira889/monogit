package tui

import (
	"github.com/JoaoOliveira889/monogit/internal/domain"
	"testing"
	"time"

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
	if m.repoViewport.Height <= 0 {
		t.Errorf("expected repo viewport height to be initialized, got %d", m.repoViewport.Height)
	}
	if m.viewport.Height != m.repoViewport.Height {
		t.Errorf("expected left and right panes to start with the same height, got left=%d right=%d", m.repoViewport.Height, m.viewport.Height)
	}
	normalHeight := m.repoViewport.Height

	m.searchMode = true
	_, _ = m.handleResize(msg)
	if m.repoViewport.Height >= normalHeight {
		t.Errorf("expected search mode to reduce repo viewport height, got %d want less than %d", m.repoViewport.Height, normalHeight)
	}
}

func TestHandleSearchEnterPersistsFilter(t *testing.T) {
	m := mkModel()
	m.repos = []domain.Repository{
		{Name: "lib-authorizer", Path: "/p1"},
		{Name: "webapi-holder", Path: "/p2"},
	}
	m.cursor = 0
	m.searchMode = true
	m.searchInput.SetValue("holder")

	_, _ = m.handleSearchKeys(tea.KeyMsg{Type: tea.KeyEnter})

	if m.searchMode {
		t.Fatal("expected search input to close after enter")
	}
	if m.searchQuery != "holder" {
		t.Fatalf("expected search query to persist, got %q", m.searchQuery)
	}
	filtered := m.filteredRepos()
	if len(filtered) != 1 || filtered[0].Name != "webapi-holder" {
		t.Fatalf("expected filtered repos to persist after enter, got %#v", filtered)
	}
}

func TestHandleSearchTypingFiltersLive(t *testing.T) {
	m := mkModel()
	m.repos = []domain.Repository{
		{Name: "lib-authorizer", Path: "/p1"},
		{Name: "webapi-holder", Path: "/p2"},
	}
	m.searchMode = true
	m.searchInput.Focus()

	for _, r := range []rune("hol") {
		_, _ = m.handleSearchKeys(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}

	if m.searchInput.Value() != "hol" {
		t.Fatalf("expected live search input to update, got %q", m.searchInput.Value())
	}
	filtered := m.filteredRepos()
	if len(filtered) != 1 || filtered[0].Name != "webapi-holder" {
		t.Fatalf("expected live filtering while typing, got %#v", filtered)
	}
	if m.searchQuery != "" {
		t.Fatalf("expected applied search query to remain unchanged while typing, got %q", m.searchQuery)
	}
}

func TestHandleSearchEscClearsAppliedFilterInRepoPanel(t *testing.T) {
	m := mkModel()
	m.repos = []domain.Repository{
		{Name: "lib-authorizer", Path: "/p1"},
		{Name: "webapi-holder", Path: "/p2"},
	}
	m.cursor = 0
	m.searchQuery = "holder"
	m.activePanel = RepoPanel

	_, _ = m.handleNormalKeys(tea.KeyMsg{Type: tea.KeyEsc})

	if m.searchQuery != "" {
		t.Fatalf("expected esc in repo panel to clear search, got %q", m.searchQuery)
	}
	if len(m.filteredRepos()) != 2 {
		t.Fatalf("expected full repo list after clearing search, got %d", len(m.filteredRepos()))
	}
}

func TestHandleSearchEscRestoresAppliedFilter(t *testing.T) {
	m := mkModel()
	m.repos = []domain.Repository{
		{Name: "lib-authorizer", Path: "/p1"},
		{Name: "webapi-holder", Path: "/p2"},
	}
	m.searchQuery = "holder"
	m.searchMode = true
	m.searchInput.SetValue("web")

	_, _ = m.handleSearchKeys(tea.KeyMsg{Type: tea.KeyEsc})

	if m.searchMode {
		t.Fatal("expected search mode to close on esc")
	}
	if got := m.searchInput.Value(); got != "holder" {
		t.Fatalf("expected applied search query to be restored, got %q", got)
	}
	filtered := m.filteredRepos()
	if len(filtered) != 1 || filtered[0].Name != "webapi-holder" {
		t.Fatalf("expected applied filter to remain active, got %#v", filtered)
	}
}

func TestHandleNewTagEscReturnsToTagEditor(t *testing.T) {
	m := mkModel()
	m.repos = []domain.Repository{{
		Name: "repo",
		Path: "/r",
		Tags: []string{"alpha"},
	}}
	m.cursor = 0
	m.activePanel = LogPanel
	m.previousPanel = RepoPanel
	m.tagAssignModal = true
	m.inputMode = true
	m.inputAction = "new_tag"
	m.commitInput.SetValue("beta")

	_, _ = m.handleInputKeys(tea.KeyMsg{Type: tea.KeyEsc})

	if m.inputMode {
		t.Fatal("expected new tag input to close on esc")
	}
	if !m.tagAssignModal {
		t.Fatal("expected tag editor to remain open after cancelling new tag")
	}
	if m.activePanel != LogPanel {
		t.Fatalf("expected to return to tag editor panel, got %v", m.activePanel)
	}
}

func TestHandleNewTagTypingWorksInsideTagModal(t *testing.T) {
	m := mkModel()
	m.repos = []domain.Repository{{
		Name: "repo",
		Path: "/r",
		Tags: []string{"alpha"},
	}}
	m.cursor = 0
	m.activePanel = LogPanel
	m.previousPanel = RepoPanel
	m.tagAssignModal = true
	m.inputMode = true
	m.inputAction = "new_tag"
	m.commitInput.Reset()
	m.commitInput.Placeholder = "New tag name..."
	m.commitInput.Focus()

	res, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")})
	m2 := res.(*Model)

	if got := m2.commitInput.Value(); got != "b" {
		t.Fatalf("expected new tag input to accept typing, got %q", got)
	}
	if !m2.inputMode || !m2.tagAssignModal {
		t.Fatal("expected new tag input to keep the tag modal open")
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
	if cmd == nil {
		t.Fatal("expected repo detail refresh command from status update")
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

func TestHandleStartupReposKeepsSplashUntilMinimumDuration(t *testing.T) {
	m := mkModel()
	m.scanning = true
	m.showSplash = true
	m.splashStartedAt = time.Now()

	_, cmd := m.handleStartupRepos(startupReposMsg{repos: []domain.Repository{{Name: "r1", Path: "/p1"}}})
	if cmd != nil {
		t.Fatal("expected nil cmd when loading startup repos")
	}
	if !m.showSplash {
		t.Error("expected splash to remain visible until the minimum duration elapses")
	}
}

func TestHandleRepoScannedHidesSplashAfterMinimumDuration(t *testing.T) {
	m := mkModel()
	m.scanning = true
	m.showSplash = true
	m.splashStartedAt = time.Now().Add(-3 * time.Second)

	_, cmd := m.handleRepoScanned(repoScannedMsg{repos: []domain.Repository{{Name: "r1", Path: "/p1"}}})
	if cmd == nil {
		t.Fatal("expected follow-up commands after repo scan")
	}
	if m.showSplash {
		t.Error("expected splash to hide after the minimum duration elapsed")
	}
}

func TestHandleFetchDone(t *testing.T) {
	m := mkModel()
	m.repos = []domain.Repository{{Name: "r1", Path: "/p1", Fetching: true}}

	msg := fetchDoneMsg{index: 0, output: "Fetched remote updates"}
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

	msg := fetchAllDoneMsg{results: []fetchAllResult{{Index: 0, Name: "r1"}, {Index: 1, Name: "r2"}}}
	_, cmd := m.handleFetchAllDone(msg)
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

func TestAddTagToRepoRespectsLimit(t *testing.T) {
	m := mkModel()
	m.cfg.RepoTags = map[string][]string{
		"/p1": {"a", "b", "c", "d"},
	}
	m.repos = []domain.Repository{{Name: "r1", Path: "/p1", Tags: []string{"a", "b", "c", "d"}}}

	m.addTagToRepo("/p1", "i")

	if len(m.cfg.RepoTags["/p1"]) != 4 {
		t.Fatalf("expected tag limit to remain at 4, got %d", len(m.cfg.RepoTags["/p1"]))
	}
	if m.statusMsg == "" {
		t.Fatal("expected a user-facing status message when tag limit is reached")
	}
}

func TestRemoveTagFromRepo(t *testing.T) {
	m := mkModel()
	m.cfg.RepoTags = map[string][]string{
		"/p1": {"a", "b", "c"},
	}
	m.repos = []domain.Repository{{Name: "r1", Path: "/p1", Tags: []string{"a", "b", "c"}}}

	m.removeTagFromRepo("/p1", "b")

	if len(m.cfg.RepoTags["/p1"]) != 2 {
		t.Fatalf("expected tag to be removed, got %v", m.cfg.RepoTags["/p1"])
	}
	if m.repos[0].Tags[0] != "a" || m.repos[0].Tags[1] != "c" {
		t.Fatalf("expected repo tags to be refreshed, got %v", m.repos[0].Tags)
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

func TestHandleEnterKeyPriority(t *testing.T) {
	m := mkModel()
	m.activePanel = LogPanel
	m.repos = []domain.Repository{{Name: "r1", Path: "/p1"}}
	m.files = []domain.FileStatus{{Name: "a.go", Staged: true}}
	m.fileSelections[0] = true

	m.showFiles = true
	m.commitStep = StepSelectFiles
	m.showBranches = true
	m.branches = []domain.BranchInfo{{Name: "branch-a"}}

	_, cmd := m.handleEnterKey()
	if cmd == nil {
		t.Fatal("expected non-nil cmd from enter key (file commit transition)")
	}
	if m.commitStep != StepMessage {
		t.Errorf("expected commitStep to transition to StepMessage, got %v", m.commitStep)
	}
	if m.showFiles {
		t.Error("expected showFiles to be false after enter key transition")
	}
}

func TestHandleNormalKeysPushAndPushAll(t *testing.T) {
	m := mkModel()
	m.repos = []domain.Repository{{Name: "r1", Path: "/p1", Ahead: 1}}

	msgPush := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("u")}
	res, _ := m.handleNormalKeys(msgPush)
	m2 := res.(*Model)
	if !m2.showConfirmModal {
		t.Error("expected showConfirmModal to be true for push")
	}
	if m2.confirmModalAction != "push" {
		t.Errorf("expected action 'push', got %s", m2.confirmModalAction)
	}

	m2.showConfirmModal = false
	m2.confirmModalAction = ""

	msgPushAll := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("U")}
	resAll, _ := m2.handleNormalKeys(msgPushAll)
	m3 := resAll.(*Model)
	if !m3.showConfirmModal {
		t.Error("expected showConfirmModal to be true for push_all")
	}
	if m3.confirmModalAction != "push_all" {
		t.Errorf("expected action 'push_all', got %s", m3.confirmModalAction)
	}
}

func TestHandleNormalKeysFetchRunsDirectly(t *testing.T) {
	m := mkModel()
	m.repos = []domain.Repository{{Name: "r1", Path: "/p1"}}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("f")}
	res, _ := m.handleNormalKeys(msg)
	m2 := res.(*Model)
	if m2.showConfirmModal {
		t.Fatal("expected fetch to run directly without confirmation modal")
	}
	if !m2.repos[0].Fetching {
		t.Fatal("expected fetch to mark repo as fetching")
	}
}

func TestHandleSelectAllMarksEveryFileWithoutConfirm(t *testing.T) {
	m := mkModel()
	m.showFiles = true
	m.commitStep = StepSelectFiles
	m.files = []domain.FileStatus{{Name: "a.go"}, {Name: "b.go"}}
	m.repos = []domain.Repository{{Name: "r1", Path: "/p1"}}

	res, _ := m.handleSelectAll()
	m2 := res.(*Model)
	if m2.showConfirmModal {
		t.Fatal("expected select all to stay local without confirmation")
	}
	if !m2.fileSelections[0] || !m2.fileSelections[1] {
		t.Fatalf("expected all files to be selected, got %#v", m2.fileSelections)
	}
}

func TestCommitWizardUsesVForManualSelection(t *testing.T) {
	m := mkModel()
	m.repos = []domain.Repository{{Name: "r1", Path: "/p1"}}
	m.activePanel = CommitWizardPanel
	m.commitStep = StepAddOption

	res, cmd := m.handleNormalKeys(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("v")})
	m2 := res.(*Model)

	if m2.showConfirmModal {
		t.Fatal("expected manual file selection to open without confirmation")
	}
	if m2.commitMode != CommitModeSelected {
		t.Fatalf("expected selected commit mode, got %v", m2.commitMode)
	}
	if m2.commitStep != StepSelectFiles {
		t.Fatalf("expected StepSelectFiles, got %v", m2.commitStep)
	}
	if !m2.showFiles {
		t.Fatal("expected file panel to open")
	}
	if cmd == nil {
		t.Fatal("expected file fetch command")
	}
}

func TestCommitWizardSpaceTogglesSelectionWithoutConfirm(t *testing.T) {
	m := mkModel()
	m.showFiles = true
	m.commitStep = StepSelectFiles
	m.files = []domain.FileStatus{{Name: "a.go"}}

	res, _ := m.handleNormalKeys(tea.KeyMsg{Type: tea.KeySpace})
	m2 := res.(*Model)

	if m2.showConfirmModal {
		t.Fatal("expected file toggle to stay local without confirmation")
	}
	if !m2.fileSelections[0] {
		t.Fatal("expected file to be selected")
	}
}

func TestCommitWizardKeyASelectsAllFiles(t *testing.T) {
	m := mkModel()
	m.showFiles = true
	m.commitStep = StepSelectFiles
	m.files = []domain.FileStatus{{Name: "a.go"}, {Name: "b.go"}}

	res, _ := m.handleNormalKeys(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	m2 := res.(*Model)

	if !m2.fileSelections[0] || !m2.fileSelections[1] {
		t.Fatalf("expected key a to select all files, got %#v", m2.fileSelections)
	}
}

func TestCommitWizardKeyNClearsAllFiles(t *testing.T) {
	m := mkModel()
	m.showFiles = true
	m.commitStep = StepSelectFiles
	m.files = []domain.FileStatus{{Name: "a.go"}, {Name: "b.go"}}
	m.fileSelections[0] = true
	m.fileSelections[1] = true

	res, _ := m.handleNormalKeys(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	m2 := res.(*Model)

	if len(m2.fileSelections) != 0 {
		t.Fatalf("expected key n to clear all selections, got %#v", m2.fileSelections)
	}
}

func TestBranchPanelKeyNOpensCreateBranchInput(t *testing.T) {
	m := mkModel()
	m.repos = []domain.Repository{{Name: "r1", Path: "/p1"}}
	m.cursor = 0
	m.activePanel = LogPanel
	m.showBranches = true
	m.branches = []domain.BranchInfo{{Name: "main"}}
	m.fileSelections[0] = true

	res, _ := m.handleNormalKeys(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	m2 := res.(*Model)

	if !m2.inputMode {
		t.Fatal("expected branch creation input to open")
	}
	if m2.inputAction != "create_branch" {
		t.Fatalf("expected inputAction create_branch, got %q", m2.inputAction)
	}
	if got := m2.commitInput.Placeholder; got != "New branch name..." {
		t.Fatalf("expected branch name placeholder, got %q", got)
	}
	if !m2.fileSelections[0] {
		t.Fatal("expected branch shortcut not to clear file selections")
	}
}

func TestHandleNormalKeysPInStashPanelOpensPopConfirmation(t *testing.T) {
	m := mkModel()
	m.repos = []domain.Repository{{Name: "r1", Path: "/p1"}}
	m.activePanel = LogPanel
	m.showStashes = true
	m.stashes = []domain.StashInfo{{Index: 0}}

	res, _ := m.handleNormalKeys(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("p")})
	m2 := res.(*Model)

	if !m2.showConfirmModal {
		t.Fatal("expected stash pop confirmation to open")
	}
	if m2.confirmModalAction != "pop_stash" {
		t.Fatalf("expected pop_stash action, got %q", m2.confirmModalAction)
	}
}

func TestHandleNormalKeysDInBranchPanelOpensDeleteConfirmation(t *testing.T) {
	m := mkModel()
	m.repos = []domain.Repository{{Name: "r1", Path: "/p1"}}
	m.cursor = 0
	m.activePanel = LogPanel
	m.showBranches = true
	m.branches = []domain.BranchInfo{{Name: "feature/test"}}

	res, _ := m.handleNormalKeys(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})
	m2 := res.(*Model)

	if !m2.showConfirmModal {
		t.Fatal("expected branch delete confirmation to open")
	}
	if m2.confirmModalAction != "delete_branch_options" {
		t.Fatalf("expected delete_branch_options action, got %q", m2.confirmModalAction)
	}
}

func TestPrepareSelectFilesClearsStashMode(t *testing.T) {
	m := mkModel()
	m.repos = []domain.Repository{{Name: "r1", Path: "/p1"}}
	m.showStashes = true
	m.showBranches = true
	m.showConflicts = true

	res, _ := m.executeConfirmedAction("prepare_select_files")
	m2 := res.(*Model)

	if m2.showStashes || m2.showBranches || m2.showConflicts {
		t.Fatalf("expected prepare_select_files to clear other modes, got stashes=%v branches=%v conflicts=%v", m2.showStashes, m2.showBranches, m2.showConflicts)
	}
}

func TestCommandLogOpensAndClearsPreviousLogs(t *testing.T) {
	m := mkModel()
	m.commandLogs = []CommandLogEntry{{RepoName: "r1", Command: "push"}}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("o")}
	res, _ := m.handleNormalKeys(msg)
	m2 := res.(*Model)
	if m2.activePanel != CommandLogPanel {
		t.Fatal("expected command log panel to open")
	}
	if len(m2.commandLogs) != 0 {
		t.Fatal("expected command logs to be cleared when opening the panel")
	}

	m2.commandLogs = []CommandLogEntry{{RepoName: "r2", Command: "pull"}}
	res2, _ := m2.handleNormalKeys(msg)
	m3 := res2.(*Model)
	if m3.activePanel == CommandLogPanel {
		t.Fatal("expected command log panel to close on second toggle")
	}
	if len(m3.commandLogs) != 0 {
		t.Fatal("expected command logs to be cleared when closing the panel")
	}
}

func TestHandleConfirmModalKeysPushAll(t *testing.T) {
	m := mkModel()
	m.repos = []domain.Repository{
		{Name: "r1", Path: "/p1", Ahead: 1},
		{Name: "r2", Path: "/p2", Ahead: 0},
	}
	m.confirmModalAction = "push_all"
	m.showConfirmModal = true

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")}
	res, cmd := m.handleConfirmModalKeys(msg)
	m2 := res.(*Model)

	if m2.showConfirmModal {
		t.Error("expected showConfirmModal to be false after confirmation")
	}
	if !m2.repos[0].Pushing {
		t.Error("expected repo 1 (ahead > 0) to have Pushing = true")
	}
	if m2.repos[1].Pushing {
		t.Error("expected repo 2 (ahead == 0) to have Pushing = false")
	}
	if cmd == nil {
		t.Fatal("expected non-nil cmd from push_all confirmation")
	}
}

func TestHandleNormalKeysStashConfirmation(t *testing.T) {
	m := mkModel()
	m.repos = []domain.Repository{{Name: "r1", Path: "/p1"}}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")}
	res, _ := m.handleNormalKeys(msg)
	m2 := res.(*Model)

	if !m2.showConfirmModal {
		t.Error("expected showConfirmModal to be true for stash")
	}
	if m2.confirmModalAction != "stash" {
		t.Errorf("expected action 'stash', got %s", m2.confirmModalAction)
	}
}

func TestHandleConfirmModalKeysStash(t *testing.T) {
	m := mkModel()
	m.repos = []domain.Repository{{Name: "r1", Path: "/p1"}}
	m.confirmModalAction = "stash"
	m.showConfirmModal = true

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")}
	res, cmd := m.handleConfirmModalKeys(msg)
	m2 := res.(*Model)

	if m2.showConfirmModal {
		t.Error("expected showConfirmModal to be false after confirmation")
	}
	if !m2.repos[0].Stashing {
		t.Error("expected repo to have Stashing = true")
	}
	if cmd == nil {
		t.Fatal("expected non-nil cmd from stash confirmation")
	}
}

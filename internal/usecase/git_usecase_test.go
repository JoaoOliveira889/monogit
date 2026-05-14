package usecase

import (
	"monogit/internal/domain"
	"testing"
)

type mockGitProvider struct {
	domain.GitProvider
	getBranchFunc      func(string) (string, error)
	getAheadBehindFunc func(string) (int, int, error)
	isDirtyFunc        func(string) (bool, error)
	fetchAllFunc       func(string) error
	pullFunc           func(string) (string, error)
	pushFunc           func(string) (string, error)
	getRemoteURLFunc   func(string) (string, error)
	addAndCommitFunc   func(string, string) (string, error)
	getStatusFilesFunc func(string) ([]domain.FileStatus, error)
	getDiffFunc        func(string, domain.FileStatus) (string, error)
	discardChangesFunc func(string, domain.FileStatus) error
	getBranchesFunc    func(string) ([]domain.BranchInfo, error)
	checkoutBranchFunc func(string, string) error
	createBranchFunc   func(string, string) error
	stashFunc          func(string, string) (string, error)
	stashPopFunc       func(string) (string, error)
	unstageAllFunc     func(string) error
	unstageFileFunc    func(string, string) error
	undoCommitFunc     func(string) error
	stageByPatternFunc func(string, string) error
	getGraphLogFunc    func(string, int) (string, error)
	getSimpleLogFunc   func(string, int) (string, error)
}

func (m *mockGitProvider) GetBranch(repoPath string) (string, error)      { return m.getBranchFunc(repoPath) }
func (m *mockGitProvider) GetAheadBehind(repoPath string) (int, int, error) { return m.getAheadBehindFunc(repoPath) }
func (m *mockGitProvider) IsDirty(repoPath string) (bool, error)         { return m.isDirtyFunc(repoPath) }
func (m *mockGitProvider) FetchAll(repoPath string) error                { return m.fetchAllFunc(repoPath) }
func (m *mockGitProvider) Pull(repoPath string) (string, error)          { return m.pullFunc(repoPath) }
func (m *mockGitProvider) Push(repoPath string) (string, error)          { return m.pushFunc(repoPath) }
func (m *mockGitProvider) GetRemoteURL(repoPath string) (string, error)   { return m.getRemoteURLFunc(repoPath) }
func (m *mockGitProvider) AddAndCommit(p, msg string) (string, error)    { return m.addAndCommitFunc(p, msg) }
func (m *mockGitProvider) GetStatusFiles(p string) ([]domain.FileStatus, error) { return m.getStatusFilesFunc(p) }
func (m *mockGitProvider) GetDiff(p string, f domain.FileStatus) (string, error) { return m.getDiffFunc(p, f) }
func (m *mockGitProvider) DiscardChanges(p string, f domain.FileStatus) error { return m.discardChangesFunc(p, f) }
func (m *mockGitProvider) GetBranches(p string) ([]domain.BranchInfo, error)        { return m.getBranchesFunc(p) }
func (m *mockGitProvider) CheckoutBranch(p, b string) error              { return m.checkoutBranchFunc(p, b) }
func (m *mockGitProvider) CreateBranch(p, b string) error                { return m.createBranchFunc(p, b) }
func (m *mockGitProvider) Stash(p, msg string) (string, error)           { return m.stashFunc(p, msg) }
func (m *mockGitProvider) StashPop(p string) (string, error)             { return m.stashPopFunc(p) }
func (m *mockGitProvider) UnstageAll(p string) error                    { return m.unstageAllFunc(p) }
func (m *mockGitProvider) UnstageFile(p, f string) error                { return m.unstageFileFunc(p, f) }
func (m *mockGitProvider) UndoCommit(p string) error                    { return m.undoCommitFunc(p) }
func (m *mockGitProvider) StageByPattern(p, pat string) error            { return m.stageByPatternFunc(p, pat) }
func (m *mockGitProvider) GetGraphLog(p string, n int) (string, error)   { return m.getGraphLogFunc(p, n) }
func (m *mockGitProvider) GetSimpleLog(p string, n int) (string, error)  { return m.getSimpleLogFunc(p, n) }

func TestGetRepositoryStatus(t *testing.T) {
	mock := &mockGitProvider{
		getBranchFunc: func(p string) (string, error) { return "main", nil },
		getAheadBehindFunc: func(p string) (int, int, error) { return 1, 2, nil },
		isDirtyFunc: func(p string) (bool, error) { return true, nil },
	}

	uc := NewGitUseCase(mock)
	repo, err := uc.GetRepositoryStatus("/path")
	if err != nil { t.Fatalf("error: %v", err) }
	if repo.Branch != "main" || repo.Ahead != 1 || repo.Behind != 2 || !repo.IsDirty {
		t.Errorf("unexpected repo state: %+v", repo)
	}
}

func TestGitUseCaseMethods(t *testing.T) {
	called := false
	mock := &mockGitProvider{
		fetchAllFunc: func(p string) error { called = true; return nil },
		pullFunc: func(p string) (string, error) { called = true; return "pulled", nil },
		pushFunc: func(p string) (string, error) { called = true; return "pushed", nil },
		addAndCommitFunc: func(p, m string) (string, error) { called = true; return "committed", nil },
		getBranchesFunc: func(p string) ([]domain.BranchInfo, error) { called = true; return []domain.BranchInfo{{Name: "b1"}}, nil },
		stashFunc: func(p, m string) (string, error) { called = true; return "stashed", nil },
		stashPopFunc: func(p string) (string, error) { called = true; return "popped", nil },
		unstageAllFunc: func(p string) error { called = true; return nil },
		unstageFileFunc: func(p, f string) error { called = true; return nil },
		undoCommitFunc: func(p string) error { called = true; return nil },
		stageByPatternFunc: func(p, pat string) error { called = true; return nil },
		getStatusFilesFunc: func(p string) ([]domain.FileStatus, error) { called = true; return []domain.FileStatus{{Name: "f1"}}, nil },
		getDiffFunc: func(p string, f domain.FileStatus) (string, error) { called = true; return "diff", nil },
		discardChangesFunc: func(p string, f domain.FileStatus) error { called = true; return nil },
		getSimpleLogFunc: func(p string, n int) (string, error) { called = true; return "log", nil },
		getGraphLogFunc: func(p string, n int) (string, error) { called = true; return "graph", nil },
		checkoutBranchFunc: func(p, b string) error { called = true; return nil },
		createBranchFunc: func(p, b string) error { called = true; return nil },
		getRemoteURLFunc: func(p string) (string, error) { called = true; return "url", nil },
	}

	uc := NewGitUseCase(mock)
	
	tests := []struct {
		name string
		fn   func() error
	}{
		{"Fetch", func() error { return uc.Fetch("/p") }},
		{"Pull", func() error { _, err := uc.Pull("/p"); return err }},
		{"Push", func() error { _, err := uc.Push("/p"); return err }},
		{"Commit", func() error { _, err := uc.Commit("/p", "msg"); return err }},
		{"GetBranches", func() error { _, err := uc.GetBranches("/p"); return err }},
		{"Stash", func() error { _, err := uc.Stash("/p", "msg"); return err }},
		{"StashPop", func() error { _, err := uc.StashPop("/p"); return err }},
		{"UnstageAll", func() error { return uc.UnstageAll("/p") }},
		{"UndoCommit", func() error { return uc.UndoCommit("/p") }},
		{"StageByPattern", func() error { return uc.StageByPattern("/p", "*") }},
		{"AddAll", func() error { return uc.AddAll("/p") }},
		{"GetFiles", func() error { _, err := uc.GetFiles("/p"); return err }},
		{"GetDiff", func() error { _, err := uc.GetDiff("/p", domain.FileStatus{Name: "f1"}); return err }},
		{"DiscardFile", func() error { return uc.DiscardFile("/p", domain.FileStatus{Name: "f1"}) }},
		{"GetSimpleLog", func() error { _, err := uc.GetSimpleLog("/p", 1); return err }},
		{"GetGraphLog", func() error { _, err := uc.GetGraphLog("/p", 1); return err }},
		{"CheckoutBranch", func() error { return uc.CheckoutBranch("/p", "b") }},
		{"CreateBranch", func() error { return uc.CreateBranch("/p", "b") }},
		{"GetRemoteURL", func() error { _, err := uc.GetRemoteURL("/p"); return err }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called = false
			if err := tt.fn(); err != nil {
				t.Errorf("%s failed: %v", tt.name, err)
			}
			if !called {
				t.Errorf("%s did not call provider", tt.name)
			}
		})
	}
}

func TestToggleFile(t *testing.T) {
	var lastMethod string
	mock := &mockGitProvider{
		unstageFileFunc: func(p, f string) error { lastMethod = "unstage"; return nil },
		stageByPatternFunc: func(p, pat string) error { lastMethod = "stage"; return nil },
	}
	uc := NewGitUseCase(mock)

	uc.ToggleFile("/p", domain.FileStatus{Name: "f1", Staged: false})
	if lastMethod != "stage" { t.Errorf("expected stage, got %s", lastMethod) }

	uc.ToggleFile("/p", domain.FileStatus{Name: "f1", Staged: true})
	if lastMethod != "unstage" { t.Errorf("expected unstage, got %s", lastMethod) }
}

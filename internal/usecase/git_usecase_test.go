package usecase

import (
	"testing"

	"github.com/JoaoOliveira889/monogit/internal/domain"
	"github.com/JoaoOliveira889/monogit/internal/testutil"
)

func TestGetRepositoryStatus(t *testing.T) {
	mock := &testutil.MockGitProvider{
		GetBranchFunc:      func(p string) (string, error) { return "main", nil },
		GetAheadBehindFunc: func(p string) (int, int, error) { return 1, 2, nil },
		IsDirtyFunc:        func(p string) (bool, error) { return true, nil },
		HasUpstreamFunc:    func(p string) (bool, error) { return true, nil },
	}

	uc := NewGitUseCase(mock)
	repo, err := uc.GetRepositoryStatus("/path")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if repo.Branch != "main" || repo.Ahead != 1 || repo.Behind != 2 || !repo.IsDirty || repo.IsDetached || !repo.HasUpstream {
		t.Errorf("unexpected repo state: %+v", repo)
	}
}

func TestGitUseCaseMethods(t *testing.T) {
	called := false
	mock := &testutil.MockGitProvider{
		FetchAllFunc:     func(p string) error { called = true; return nil },
		PullFunc:         func(p string) (string, error) { called = true; return "pulled", nil },
		PushFunc:         func(p string) (string, error) { called = true; return "pushed", nil },
		AddAndCommitFunc: func(p, m string) (string, error) { called = true; return "committed", nil },
		CommitFunc:       func(p, m string) (string, error) { called = true; return "committed", nil },
		GetBranchesFunc: func(p string) ([]domain.BranchInfo, error) {
			called = true
			return []domain.BranchInfo{{Name: "b1"}}, nil
		},
		StashFunc:               func(p, m string) (string, error) { called = true; return "stashed", nil },
		StashPopFunc:            func(p string) (string, error) { called = true; return "popped", nil },
		UnstageAllFunc:          func(p string) error { called = true; return nil },
		UnstageFileFunc:         func(p, f string) error { called = true; return nil },
		UndoCommitFunc:          func(p string) error { called = true; return nil },
		StageByPatternFunc:      func(p, pat string) error { called = true; return nil },
		StageFilesFunc:          func(p string, files []string) error { called = true; return nil },
		GetStatusFilesFunc: func(p string) ([]domain.FileStatus, error) {
			called = true
			return []domain.FileStatus{{Name: "f1"}}, nil
		},
		GetDiffFunc:          func(p string, f domain.FileStatus) (string, error) { called = true; return "diff", nil },
		DiscardChangesFunc:   func(p string, f domain.FileStatus) error { called = true; return nil },
		GetSimpleLogFunc:     func(p string, n int) (string, error) { called = true; return "log", nil },
		GetGraphLogFunc:      func(p string, n int) (string, error) { called = true; return "graph", nil },
		GetRepositorySnapshotFunc: func(p string, viewGraph bool, n int) (domain.RepositorySnapshot, error) {
			called = true
			return domain.RepositorySnapshot{Branch: "main"}, nil
		},
		CheckoutBranchFunc:     func(p, b string) error { called = true; return nil },
		CreateBranchFunc:       func(p, b string) error { called = true; return nil },
		GetRemoteURLFunc:       func(p string) (string, error) { called = true; return "url", nil },
		CreateTagFunc:          func(p, n, m string) (string, error) { called = true; return "tagged", nil },
		PushTagFunc:            func(p, n string) (string, error) { called = true; return "pushed", nil },
		DeleteBranchFunc:       func(p, n string) (string, error) { called = true; return "deleted", nil },
		DeleteRemoteBranchFunc: func(p, r, n string) (string, error) { called = true; return "deleted remote", nil },
		GetStashesFunc:         func(p string) ([]domain.StashInfo, error) { called = true; return []domain.StashInfo{{Index: 0}}, nil },
		ApplyStashFunc:         func(p string, idx int) (string, error) { called = true; return "applied", nil },
		DropStashFunc:          func(p string, idx int) (string, error) { called = true; return "dropped", nil },
		PopStashFunc:           func(p string, idx int) (string, error) { called = true; return "popped", nil },
		MergeFunc:              func(p, b string) (string, error) { called = true; return "merged", nil },
		OpenMergetoolFunc: func(p, tool, file string) (domain.CommandSpec, error) {
			called = true
			return domain.CommandSpec{Name: "git"}, nil
		},
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
		{"CommitAll", func() error { _, err := uc.CommitAll("/p", "msg"); return err }},
		{"GetBranches", func() error { _, err := uc.GetBranches("/p"); return err }},
		{"Stash", func() error { _, err := uc.Stash("/p", "msg"); return err }},
		{"StashPop", func() error { _, err := uc.StashPop("/p"); return err }},
		{"UnstageAll", func() error { return uc.UnstageAll("/p") }},
		{"UndoCommit", func() error { return uc.UndoCommit("/p") }},
		{"StageByPattern", func() error { return uc.StageByPattern("/p", "*") }},
		{"StageFiles", func() error { return uc.git.StageFiles("/p", []string{"file.txt"}) }},
		{"AddAll", func() error { return uc.AddAll("/p") }},
		{"GetFiles", func() error { _, err := uc.GetFiles("/p"); return err }},
		{"GetDiff", func() error { _, err := uc.GetDiff("/p", domain.FileStatus{Name: "f1"}); return err }},
		{"DiscardFile", func() error { return uc.DiscardFile("/p", domain.FileStatus{Name: "f1"}) }},
		{"GetSimpleLog", func() error { _, err := uc.GetSimpleLog("/p", 1); return err }},
		{"GetGraphLog", func() error { _, err := uc.GetGraphLog("/p", 1); return err }},
		{"GetRepositorySnapshot", func() error { _, err := uc.GetRepositorySnapshot("/p", true, 30); return err }},
		{"CheckoutBranch", func() error { return uc.CheckoutBranch("/p", "b") }},
		{"CreateBranch", func() error { return uc.CreateBranch("/p", "b") }},
		{"GetRemoteURL", func() error { _, err := uc.GetRemoteURL("/p"); return err }},
		{"CreateAndPushTag", func() error { _, err := uc.CreateAndPushTag("/p", "v1", "msg"); return err }},
		{"DeleteBranch", func() error { _, err := uc.DeleteBranch("/p", "b"); return err }},
		{"DeleteRemoteBranch", func() error { _, err := uc.DeleteRemoteBranch("/p", "origin", "b"); return err }},
		{"GetStashes", func() error { _, err := uc.GetStashes("/p"); return err }},
		{"ApplyStash", func() error { _, err := uc.ApplyStash("/p", 0); return err }},
		{"DropStash", func() error { _, err := uc.DropStash("/p", 0); return err }},
		{"Merge", func() error { _, err := uc.Merge("/p", "feature"); return err }},
		{"PopStash", func() error { _, err := uc.PopStash("/p", 0); return err }},
		{"OpenMergetool", func() error { _, err := uc.OpenMergetool("/p", "meld", "conflict.txt"); return err }},
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

func TestCommitSelected(t *testing.T) {
	var calls []string
	mock := &testutil.MockGitProvider{
		UnstageAllFunc: func(p string) error {
			calls = append(calls, "unstage")
			return nil
		},
		StageFilesFunc: func(p string, files []string) error {
			calls = append(calls, "stage:"+files[0]+","+files[1])
			return nil
		},
		CommitFunc: func(p, msg string) (string, error) {
			calls = append(calls, "commit:"+msg)
			return "ok", nil
		},
	}

	uc := NewGitUseCase(mock)
	if _, err := uc.CommitSelected("/p", []string{"a.go", "b.go"}, "msg"); err != nil {
		t.Fatalf("CommitSelected failed: %v", err)
	}

	want := []string{"unstage", "stage:a.go,b.go", "commit:msg"}
	if len(calls) != len(want) {
		t.Fatalf("unexpected call count: got %v want %v", calls, want)
	}
	for i := range want {
		if calls[i] != want[i] {
			t.Fatalf("call %d = %q, want %q", i, calls[i], want[i])
		}
	}
}

func TestToggleFile(t *testing.T) {
	var lastMethod string
	mock := &testutil.MockGitProvider{
		UnstageFileFunc:    func(p, f string) error { lastMethod = "unstage"; return nil },
		StageByPatternFunc: func(p, pat string) error { lastMethod = "stage"; return nil },
	}
	uc := NewGitUseCase(mock)

	uc.ToggleFile("/p", domain.FileStatus{Name: "f1", Staged: false})
	if lastMethod != "stage" {
		t.Errorf("expected stage, got %s", lastMethod)
	}

	uc.ToggleFile("/p", domain.FileStatus{Name: "f1", Staged: true})
	if lastMethod != "unstage" {
		t.Errorf("expected unstage, got %s", lastMethod)
	}
}

func TestHasUnpushedHeadTag(t *testing.T) {
	called := false
	mock := &testutil.MockGitProvider{
		HasUnpushedHeadTagFunc: func(p string) (bool, error) {
			called = true
			if p != "/p" {
				t.Errorf("expected path /p, got %s", p)
			}
			return true, nil
		},
	}
	uc := NewGitUseCase(mock)
	got, err := uc.HasUnpushedHeadTag("/p")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected HasUnpushedHeadTag to be called")
	}
	if !got {
		t.Error("expected true, got false")
	}
}


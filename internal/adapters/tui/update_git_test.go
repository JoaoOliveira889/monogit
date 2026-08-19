package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoaoOliveira889/monogit/internal/domain"
	"github.com/JoaoOliveira889/monogit/internal/testutil"
	"github.com/JoaoOliveira889/monogit/internal/usecase"
)

func TestHandleGitOperationDone_MergetoolRefreshesConflicts(t *testing.T) {
	m := mkModel()
	m.repos = []domain.Repository{{Name: "repo1", Path: "/repo1"}}
	m.cursor = 0
	m.conflictFiles = []domain.ConflictFile{{Name: "conflict.txt", Status: "UU"}}
	m.showConflicts = true

	m.gitUC = usecase.NewGitUseCase(&testutil.MockGitProvider{
		HasConflictsFunc: func(path string) (bool, error) { return true, nil },
		ListConflictingFilesFunc: func(path string) ([]domain.ConflictFile, error) {
			return []domain.ConflictFile{{Name: "conflict.txt", Status: "UU"}}, nil
		},
	})

	next, cmd := m.handleGitOperationDone(mergetoolDoneMsg{
		index: 0,
		path:  "/repo1",
		file:  "conflict.txt",
	})

	updated, ok := next.(*Model)
	if !ok {
		t.Fatalf("expected *Model, got %T", next)
	}
	if updated.statusMsg != "Merge resolution complete" {
		t.Fatalf("unexpected status: %q", updated.statusMsg)
	}
	if cmd == nil {
		t.Fatal("expected refresh command")
	}

	msg := cmd()
	batch, ok := msg.(tea.BatchMsg)
	if !ok {
		batch = tea.BatchMsg{func() tea.Msg { return msg }}
	}
	foundConflictRefresh := false
	for _, inner := range batch {
		if inner == nil {
			continue
		}
		innerMsg := inner()
		if conflicts, ok := innerMsg.(conflictFilesMsg); ok {
			foundConflictRefresh = true
			if len(conflicts.files) != 1 || conflicts.files[0].Name != "conflict.txt" {
				t.Fatalf("unexpected conflicts: %+v", conflicts.files)
			}
		}
	}
	if !foundConflictRefresh {
		t.Fatal("expected conflict refresh command in batch")
	}
}

func TestHandleGitOperationDone_StashRefreshesOnlySelectedRepo(t *testing.T) {
	statusCalls := 0
	m := mkModel()
	m.repos = []domain.Repository{{Name: "repo1", Path: "/repo1"}, {Name: "repo2", Path: "/repo2"}}
	m.gitUC = usecase.NewGitUseCase(&testutil.MockGitProvider{
		GetQuickSnapshotFunc: func(path string) (domain.RepositorySnapshot, error) {
			statusCalls++
			return domain.RepositorySnapshot{Branch: "main"}, nil
		},
		GetRepositorySnapshotFunc: func(path string, graph bool, lines int) (domain.RepositorySnapshot, error) {
			return domain.RepositorySnapshot{Branch: "main"}, nil
		},
	})

	_, cmd := m.handleGitOperationDone(stashDoneMsg{index: 0})
	if cmd == nil {
		t.Fatal("expected selected repo refresh")
	}
	batch, ok := cmd().(tea.BatchMsg)
	if !ok {
		t.Fatalf("expected batch refresh, got %T", cmd())
	}
	for _, inner := range batch {
		if inner != nil {
			_ = inner()
		}
	}
	if statusCalls != 1 {
		t.Fatalf("expected 1 selected status refresh, got %d", statusCalls)
	}
}

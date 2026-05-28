package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"monogit/internal/domain"
	"monogit/internal/usecase"
)

func TestHandleGitOperationDone_MergetoolRefreshesConflicts(t *testing.T) {
	m := mkModel()
	m.repos = []domain.Repository{{Name: "repo1", Path: "/repo1"}}
	m.cursor = 0
	m.conflictFiles = []domain.ConflictFile{{Name: "conflict.txt", Status: "UU"}}
	m.showConflicts = true

	m.gitUC = usecase.NewGitUseCase(&mockGitProvider{
		hasConflictsFunc: func(path string) (bool, error) { return true, nil },
		listConflictsFunc: func(path string) ([]domain.ConflictFile, error) {
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
		t.Fatalf("expected tea.BatchMsg, got %T", msg)
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

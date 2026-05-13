package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanForRepos(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "monogit-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	repo1 := filepath.Join(tempDir, "repo1")
	repo2 := filepath.Join(tempDir, "repo2")
	notRepo := filepath.Join(tempDir, "not-a-repo")

	for _, d := range []string{repo1, repo2, notRepo} {
		if err := os.MkdirAll(d, 0755); err != nil {
			t.Fatalf("failed to create dir %s: %v", d, err)
		}
	}

	if err := os.MkdirAll(filepath.Join(repo1, ".git"), 0755); err != nil {
		t.Fatalf("failed to create .git dir in repo1: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(repo2, ".git"), 0755); err != nil {
		t.Fatalf("failed to create .git dir in repo2: %v", err)
	}

	repos, err := ScanForRepos(tempDir)
	if err != nil {
		t.Fatalf("ScanForRepos failed: %v", err)
	}

	if len(repos) != 2 {
		t.Errorf("expected 2 repos, got %d", len(repos))
	}

	foundRepo1 := false
	foundRepo2 := false
	for _, r := range repos {
		if r.Name == "repo1" {
			foundRepo1 = true
		}
		if r.Name == "repo2" {
			foundRepo2 = true
		}
	}

	if !foundRepo1 || !foundRepo2 {
		t.Errorf("expected to find repo1 and repo2")
	}
}

func TestScanForReposEmpty(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "monogit-test-empty-*")
	if err != nil { t.Fatal(err) }
	defer os.RemoveAll(tempDir)

	repos, err := ScanForRepos(tempDir)
	if err != nil { t.Fatal(err) }
	if len(repos) != 0 {
		t.Errorf("expected 0 repos, got %d", len(repos))
	}
}

func TestScanForReposSubdir(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "monogit-test-subdir-*")
	if err != nil { t.Fatal(err) }
	defer os.RemoveAll(tempDir)

	repoPath := filepath.Join(tempDir, "repo")
	if err := os.MkdirAll(filepath.Join(repoPath, ".git"), 0755); err != nil { t.Fatal(err) }

	repos, err := ScanForRepos(tempDir)
	if err != nil { t.Fatal(err) }
	if len(repos) != 1 {
		t.Fatalf("expected 1 repo, got %d", len(repos))
	}
	if repos[0].Name != "repo" {
		t.Errorf("expected repo name 'repo', got %s", repos[0].Name)
	}
}

func TestScanForReposNested(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "monogit-test-nested-*")
	if err != nil { t.Fatal(err) }
	defer os.RemoveAll(tempDir)

	repoPath := filepath.Join(tempDir, "org", "repo", "wt", "main")
	if err := os.MkdirAll(filepath.Join(repoPath, ".git"), 0755); err != nil { t.Fatal(err) }

	repos, err := ScanForRepos(tempDir)
	if err != nil { t.Fatal(err) }
	if len(repos) != 1 {
		t.Fatalf("expected 1 repo, got %d", len(repos))
	}

	expectedName := filepath.Join("org", "repo", "wt", "main")
	if repos[0].Name != expectedName {
		t.Errorf("expected repo name %s, got %s", expectedName, repos[0].Name)
	}
}

func TestScanForReposWorktree(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "monogit-test-worktree-*")
	if err != nil { t.Fatal(err) }
	defer os.RemoveAll(tempDir)

	repoPath := filepath.Join(tempDir, "worktree")
	if err := os.MkdirAll(repoPath, 0755); err != nil { t.Fatal(err) }
	
	if err := os.WriteFile(filepath.Join(repoPath, ".git"), []byte("gitdir: /path/to/real/git"), 0644); err != nil {
		t.Fatal(err)
	}

	repos, err := ScanForRepos(tempDir)
	if err != nil { t.Fatal(err) }
	if len(repos) != 1 {
		t.Fatalf("expected 1 repo, got %d", len(repos))
	}
	if repos[0].Name != "worktree" {
		t.Errorf("expected repo name 'worktree', got %s", repos[0].Name)
	}
}

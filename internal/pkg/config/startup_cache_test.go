package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/JoaoOliveira889/monogit/internal/domain"
)

func TestSaveAndLoadStartupRepos(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("HOME", tempDir)
	t.Setenv("XDG_CONFIG_HOME", tempDir)

	rootDir := filepath.Join(tempDir, "workspace")
	repos := []domain.Repository{
		{Name: "repo-a", Path: filepath.Join(rootDir, "repo-a")},
		{Name: "repo-b", Path: filepath.Join(rootDir, "repo-b")},
	}

	if err := SaveStartupRepos(rootDir, repos); err != nil {
		t.Fatalf("SaveStartupRepos failed: %v", err)
	}

	info, err := os.Stat(GetStartupCachePath())
	if err != nil {
		t.Fatalf("stat startup cache: %v", err)
	}
	if got := info.Mode().Perm(); got != 0600 {
		t.Fatalf("expected 0600 permissions, got %o", got)
	}

	loaded, err := LoadStartupRepos(rootDir, map[string][]string{
		repos[0].Path: []string{"tag-a"},
	})
	if err != nil {
		t.Fatalf("LoadStartupRepos failed: %v", err)
	}

	if len(loaded) != 2 {
		t.Fatalf("expected 2 repos, got %d", len(loaded))
	}
	if loaded[0].Name != "repo-a" || loaded[0].Path != repos[0].Path {
		t.Fatalf("unexpected first repo: %#v", loaded[0])
	}
	if len(loaded[0].Tags) != 1 || loaded[0].Tags[0] != "tag-a" {
		t.Fatalf("expected tags to be restored from config, got %#v", loaded[0].Tags)
	}

	dirInfo, err := os.Stat(filepath.Dir(GetStartupCachePath()))
	if err != nil {
		t.Fatalf("stat startup cache dir: %v", err)
	}
	if got := dirInfo.Mode().Perm(); got != 0700 {
		t.Fatalf("expected 0700 directory permissions, got %o", got)
	}
}

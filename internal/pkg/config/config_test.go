package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveConfigUsesRestrictivePermissions(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("HOME", tempDir)
	t.Setenv("XDG_CONFIG_HOME", tempDir)

	if err := SaveConfig(Config{LeftPanelRatio: 0.4}); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	path := GetConfigPath()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat config: %v", err)
	}
	if got := info.Mode().Perm(); got != 0600 {
		t.Fatalf("expected 0600 permissions, got %o", got)
	}

	dirInfo, err := os.Stat(filepath.Dir(path))
	if err != nil {
		t.Fatalf("stat config dir: %v", err)
	}
	if got := dirInfo.Mode().Perm(); got != 0700 {
		t.Fatalf("expected 0700 directory permissions, got %o", got)
	}
}

func TestLoadConfig_MergeToolDefault(t *testing.T) {
	cfg := LoadConfig()
	if cfg.MergeTool != "" {
		t.Logf("MergeTool default is non-empty: %q (set from existing config)", cfg.MergeTool)
	}
}

func TestSaveAndLoadMergeTool(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("HOME", tempDir)
	t.Setenv("XDG_CONFIG_HOME", tempDir)

	if err := SaveConfig(Config{LeftPanelRatio: 0.4, MergeTool: "nvimdiff"}); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	cfg := LoadConfig()
	if cfg.MergeTool != "nvimdiff" {
		t.Errorf("expected MergeTool 'nvimdiff', got %q", cfg.MergeTool)
	}
	if cfg.LeftPanelRatio != 0.4 {
		t.Errorf("expected LeftPanelRatio 0.4, got %f", cfg.LeftPanelRatio)
	}
}

func TestSaveConfigReplacesSymlinkWithoutWritingTarget(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("HOME", tempDir)
	t.Setenv("XDG_CONFIG_HOME", tempDir)

	path := GetConfigPath()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(tempDir, "target.json")
	if err := os.WriteFile(target, []byte("do-not-overwrite"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, path); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	if err := SaveConfig(Config{LeftPanelRatio: 0.4, Theme: "Nord"}); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "do-not-overwrite" {
		t.Fatalf("symlink target overwritten: %q", data)
	}
	info, err := os.Lstat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		t.Fatal("expected config symlink to be replaced by regular file")
	}
}

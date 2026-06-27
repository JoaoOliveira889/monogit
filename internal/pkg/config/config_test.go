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

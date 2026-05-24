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

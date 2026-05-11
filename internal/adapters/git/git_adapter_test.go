package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestGitCLIAdapter_GetStatusFiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "monogit-git-test-*")
	if err != nil { t.Fatal(err) }
	defer os.RemoveAll(tmpDir)

	adapter := NewGitCLIAdapter()
	_, err = exec.Command("git", "-C", tmpDir, "init").Output()
	if err != nil { t.Skip("git not available or failed to init") }

	os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("content"), 0644)
	
	files, err := adapter.GetStatusFiles(tmpDir)
	if err != nil { t.Fatalf("GetStatusFiles failed: %v", err) }
	
	found := false
	for _, f := range files {
		if f.Name == "file1.txt" && f.Untracked {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected to find untracked file1.txt, got %+v", files)
	}
}

func TestValidatePattern(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		wantErr bool
	}{
		{"valid", "file.txt", false},
		{"dot", ".", false},
		{"empty", "", true},
		{"traversal", "../file", true},
		{"absolute", "/file", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validatePattern(tt.pattern); (err != nil) != tt.wantErr {
				t.Errorf("validatePattern() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateRepoPath(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "monogit-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"valid dir", tmpDir, false},
		{"empty path", "", true},
		{"non-existent", filepath.Join(tmpDir, "nope"), true},
		{"file not dir", filepath.Join(tmpDir, "file"), true},
	}

	if err := os.WriteFile(filepath.Join(tmpDir, "file"), []byte("hi"), 0644); err != nil {
		t.Fatal(err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateRepoPath(tt.path); (err != nil) != tt.wantErr {
				t.Errorf("validateRepoPath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidatePathContainment(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "monogit-test-contain-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	realTmpDir, _ := filepath.EvalSymlinks(tmpDir)

	tests := []struct {
		name    string
		base    string
		target  string
		wantErr bool
	}{
		{"inside", realTmpDir, filepath.Join(realTmpDir, "file.txt"), false},
		{"same", realTmpDir, realTmpDir, false},
		{"outside", realTmpDir, filepath.Dir(realTmpDir), true},
		{"traversal", realTmpDir, filepath.Join(realTmpDir, "..", "other"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validatePathContainment(tt.base, tt.target); (err != nil) != tt.wantErr {
				t.Errorf("validatePathContainment() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

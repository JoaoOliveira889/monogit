package git

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestGitCLIAdapter_GetStatusFiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "monogit-git-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	adapter := NewGitCLIAdapter()
	_, err = exec.Command("git", "-C", tmpDir, "init").Output()
	if err != nil {
		t.Skip("git not available or failed to init")
	}

	if err := os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	files, err := adapter.GetStatusFiles(tmpDir)
	if err != nil {
		t.Fatalf("GetStatusFiles failed: %v", err)
	}

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

	if err := os.WriteFile(filepath.Join(tmpDir, "file"), []byte("hi"), 0644); err != nil {
		t.Fatal(err)
	}

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

	tests := []struct {
		name    string
		base    string
		target  string
		wantErr bool
	}{
		{"inside", tmpDir, filepath.Join(tmpDir, "file.txt"), false},
		{"same", tmpDir, tmpDir, false},
		{"outside", tmpDir, filepath.Dir(tmpDir), true},
		{"traversal", tmpDir, filepath.Join(tmpDir, "..", "other"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validatePathContainment(tt.base, tt.target); (err != nil) != tt.wantErr {
				t.Errorf("validatePathContainment() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateBranchName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		valid bool
	}{
		{"simple", "main", true},
		{"with-slash", "feature/login", true},
		{"with-dot", "v1.0", true},
		{"with-hyphen", "my-branch", true},
		{"with-underscore", "my_branch", true},
		{"empty", "", false},
		{"starts-with-hyphen", "-branch", false},
		{"contains-dotdot", "..", false},
		{"contains-at-brace", "a@{b", false},
		{"contains-space", "bad branch", false},
		{"too-long", string(make([]byte, 256)), false},
		{"starts-with-slash", "/branch", false},
		{"ends-with-slash", "branch/", false},
		{"double-slash", "a//b", false},
		{"ends-with-dot", "branch.", false},
		{"just-dot", ".", false},
		{"non-ascii", "é", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateBranchName(tt.input)
			if tt.valid && err != nil {
				t.Errorf("unexpected error for valid name %q: %v", tt.input, err)
			}
			if !tt.valid && err == nil {
				t.Errorf("expected error for invalid name %q", tt.input)
			}
		})
	}
}

func TestValidateCommitMessage(t *testing.T) {
	tests := []struct {
		name  string
		input string
		valid bool
	}{
		{"simple", "fix: resolve login bug", true},
		{"empty", "", false},
		{"whitespace", "   ", false},
		{"starts-with-hyphen", "-m flag injection", false},
		{"contains-backtick", "msg with `", false},
		{"contains-dollar", "msg with $", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCommitMessage(tt.input)
			if tt.valid && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.valid && err == nil {
				t.Errorf("expected error for input %q", tt.input)
			}
		})
	}
}

func TestLimitedWriter(t *testing.T) {
	t.Run("writes within limit", func(t *testing.T) {
		var buf bytes.Buffer
		lw := &limitedWriter{buf: &buf, max: 100}
		n, err := lw.Write([]byte("hello"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if n != 5 {
			t.Errorf("expected 5, got %d", n)
		}
		if buf.String() != "hello" {
			t.Errorf("expected 'hello', got %q", buf.String())
		}
	})

	t.Run("truncates at limit", func(t *testing.T) {
		var buf bytes.Buffer
		lw := &limitedWriter{buf: &buf, max: 5}
		n, err := lw.Write([]byte("hello world"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if n != 5 {
			t.Errorf("expected 5, got %d", n)
		}
		if buf.String() != "hello" {
			t.Errorf("expected 'hello', got %q", buf.String())
		}
	})

	t.Run("exact limit", func(t *testing.T) {
		var buf bytes.Buffer
		lw := &limitedWriter{buf: &buf, max: 3}
		n, err := lw.Write([]byte("abc"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if n != 3 {
			t.Errorf("expected 3, got %d", n)
		}
	})

	t.Run("overflow returns 0", func(t *testing.T) {
		var buf bytes.Buffer
		lw := &limitedWriter{buf: &buf, max: 3}
		_, _ = lw.Write([]byte("abc"))
		n, err := lw.Write([]byte("extra"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if n != 0 {
			t.Errorf("expected 0, got %d", n)
		}
	})
}

func TestSynthesizeNewFileDiff(t *testing.T) {
	adapter := NewGitCLIAdapter()

	t.Run("non-empty content", func(t *testing.T) {
		diff := adapter.synthesizeNewFileDiff("test.txt", "hello\nworld\n")
		if diff == "" {
			t.Fatal("expected non-empty diff")
		}
		if !contains(diff, "+hello") {
			t.Error("expected +hello in diff")
		}
		if !contains(diff, "+world") {
			t.Error("expected +world in diff")
		}
	})

	t.Run("empty content", func(t *testing.T) {
		diff := adapter.synthesizeNewFileDiff("empty.txt", "")
		if diff == "" {
			t.Error("non-empty diff expected even for empty file")
		}
	})

	t.Run("single line", func(t *testing.T) {
		diff := adapter.synthesizeNewFileDiff("single.txt", "only line")
		if !contains(diff, "+only line") {
			t.Error("expected single line content in diff")
		}
	})
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && containsStr(s, substr)
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

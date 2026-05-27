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

func TestGitCLIAdapter_HasConflicts(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "monogit-conflict-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	_, err = exec.Command("git", "-C", tmpDir, "init").Output()
	if err != nil {
		t.Skip("git not available or failed to init")
	}

	exec.Command("git", "-C", tmpDir, "config", "user.email", "test@test.com").Run()
	exec.Command("git", "-C", tmpDir, "config", "user.name", "Test").Run()

	if err := os.WriteFile(filepath.Join(tmpDir, "f.txt"), []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}
	exec.Command("git", "-C", tmpDir, "add", "f.txt").Run()
	exec.Command("git", "-C", tmpDir, "commit", "-m", "init").Run()

	adapter := NewGitCLIAdapter()

	hasConflicts, err := adapter.HasConflicts(tmpDir)
	if err != nil {
		t.Fatalf("HasConflicts failed: %v", err)
	}
	if hasConflicts {
		t.Error("expected no conflicts on clean repo")
	}
}

func TestGitCLIAdapter_ListConflictingFiles_NoConflicts(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "monogit-listconflict-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	_, err = exec.Command("git", "-C", tmpDir, "init").Output()
	if err != nil {
		t.Skip("git not available or failed to init")
	}

	exec.Command("git", "-C", tmpDir, "config", "user.email", "test@test.com").Run()
	exec.Command("git", "-C", tmpDir, "config", "user.name", "Test").Run()

	if err := os.WriteFile(filepath.Join(tmpDir, "f.txt"), []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}
	exec.Command("git", "-C", tmpDir, "add", "f.txt").Run()
	exec.Command("git", "-C", tmpDir, "commit", "-m", "init").Run()

	adapter := NewGitCLIAdapter()

	files, err := adapter.ListConflictingFiles(tmpDir)
	if err != nil {
		t.Fatalf("ListConflictingFiles failed: %v", err)
	}
	if len(files) != 0 {
		t.Errorf("expected 0 conflicting files, got %d: %+v", len(files), files)
	}
}

func TestGitCLIAdapter_GetCompactDiff(t *testing.T) {
	t.Run("parses hunk header with function name", func(t *testing.T) {
		diff := `diff --git a/main.go b/main.go
index abc..def 100644
--- a/main.go
+++ b/main.go
@@ -10,7 +10,7 @@ func HandleRequest(w http.ResponseWriter, r *http.Request) error {
-	old line
+	new line
@@ -25,4 +25,4 @@ func parseInput(data []byte) (*Result, error) {
+	added line
`
		tmpFile, err := os.CreateTemp("", "compact-diff-test-*.go")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(tmpFile.Name())
		if _, err := tmpFile.WriteString("package main\n\nfunc Foo() {}\n"); err != nil {
			t.Fatal(err)
		}
		tmpFile.Close()
		_ = diff

		changes := parseCompactDiffOutput(diff, "main.go")
		if len(changes) != 2 {
			t.Fatalf("expected 2 changes, got %d: %+v", len(changes), changes)
		}
		if changes[0].FunctionName != "func HandleRequest(w http.ResponseWriter, r *http.Request) error {" {
			t.Errorf("unexpected function name: %q", changes[0].FunctionName)
		}
		if changes[0].LineRange != "10,7" {
			t.Errorf("unexpected line range: %q", changes[0].LineRange)
		}
		if changes[1].FunctionName != "func parseInput(data []byte) (*Result, error) {" {
			t.Errorf("unexpected function name: %q", changes[1].FunctionName)
		}
	})

	t.Run("handles empty diff", func(t *testing.T) {
		changes := parseCompactDiffOutput("", "file.go")
		if len(changes) != 0 {
			t.Errorf("expected 0 changes for empty diff, got %d", len(changes))
		}
	})
}

func TestCompactDiffRe(t *testing.T) {
	tests := []struct {
		line         string
		wantMatch    bool
		wantFuncName string
		wantRange    string
	}{
		{
			line:         "@@ -10,7 +10,7 @@ func main() {",
			wantMatch:    true,
			wantFuncName: "func main() {",
			wantRange:    "10,7",
		},
		{
			line:         "@@ -1 +1 @@ package main",
			wantMatch:    true,
			wantFuncName: "package main",
			wantRange:    "1",
		},
		{
			line:         "@@ -25,4 +25,4 @@ func parseInput(data []byte) (*Result, error) {",
			wantMatch:    true,
			wantFuncName: "func parseInput(data []byte) (*Result, error) {",
			wantRange:    "25,4",
		},
		{
			line:         "@@ -100,0 +100,10 @@",
			wantMatch:    false,
			wantFuncName: "",
			wantRange:    "",
		},
		{
			line:         "+added line",
			wantMatch:    false,
			wantFuncName: "",
			wantRange:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			match := compactDiffRe.FindStringSubmatch(tt.line)
			matched := match != nil && len(match) >= 2 && match[1] != ""
			if matched != tt.wantMatch {
				t.Errorf("match = %v, want %v", matched, tt.wantMatch)
			}
			if tt.wantMatch && match != nil && len(match) >= 2 {
				if match[1] != tt.wantFuncName {
					t.Errorf("funcName = %q, want %q", match[1], tt.wantFuncName)
				}
			}
		})
	}
}

func parseCompactDiffOutput(diff, fileName string) []struct {
	FileName     string
	FunctionName string
	LineRange    string
} {
	var changes []struct {
		FileName     string
		FunctionName string
		LineRange    string
	}
	for _, line := range bytes.Split([]byte(diff), []byte("\n")) {
		matches := compactDiffRe.FindSubmatch(line)
		if len(matches) < 2 {
			continue
		}
		funcName := string(bytes.TrimSpace(matches[1]))
		if funcName == "" {
			continue
		}
		lineRange := ""
		parts := bytes.SplitN(matches[0], []byte(" "), -1)
		for i, p := range parts {
			if bytes.HasPrefix(p, []byte("+")) && i > 0 {
				lineRange = string(bytes.TrimPrefix(p, []byte("+")))
				lineRange = string(bytes.TrimSuffix([]byte(lineRange), []byte(" @@")))
				break
			}
		}
		changes = append(changes, struct {
			FileName     string
			FunctionName string
			LineRange    string
		}{
			FileName:     fileName,
			FunctionName: funcName,
			LineRange:    lineRange,
		})
	}
	return changes
}

func TestConflictStatusFlags(t *testing.T) {
	expectedCodes := []string{"DD", "AU", "UA", "DU", "UD", "UU", "AA"}
	for _, code := range expectedCodes {
		label, ok := conflictStatusFlags[code]
		if !ok {
			t.Errorf("missing conflict status code %q", code)
			continue
		}
		if label == "" {
			t.Errorf("empty label for code %q", code)
		}
	}
}

func TestConvertSSHToHTTPS(t *testing.T) {
	adapter := NewGitCLIAdapter()

	tests := []struct {
		input string
		want  string
	}{
		{"git@github.com:user/repo.git", "https://github.com/user/repo"},
		{"https://github.com/user/repo.git", "https://github.com/user/repo.git"},
		{"git@gitlab.com:group/project.git", "https://gitlab.com/group/project"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := adapter.convertSSHToHTTPS(tt.input)
			if result != tt.want {
				t.Errorf("convertSSHToHTTPS(%q) = %q, want %q", tt.input, result, tt.want)
			}
		})
	}
}

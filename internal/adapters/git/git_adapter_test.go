package git

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JoaoOliveira889/monogit/internal/domain"
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

func TestOpenMergetool(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "monogit-mergetool-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	adapter := NewGitCLIAdapter()
	_, err = exec.Command("git", "-C", tmpDir, "init").Output()
	if err != nil {
		t.Skip("git not available or failed to init")
	}

	spec, err := adapter.OpenMergetool(tmpDir, "meld", "src/conflict.txt")
	if err != nil {
		t.Fatalf("OpenMergetool failed: %v", err)
	}

	if spec.Name != "git" {
		t.Fatalf("expected git command, got %q", spec.Name)
	}
	if spec.Dir != tmpDir {
		t.Fatalf("expected dir %q, got %q", tmpDir, spec.Dir)
	}
	if len(spec.Args) != 4 {
		t.Fatalf("expected 4 args, got %d: %+v", len(spec.Args), spec.Args)
	}
	if spec.Args[0] != "mergetool" || spec.Args[1] != "--tool=meld" || spec.Args[2] != "--" || spec.Args[3] != "src/conflict.txt" {
		t.Fatalf("unexpected mergetool args: %+v", spec.Args)
	}

	if _, err := adapter.OpenMergetool(tmpDir, "meld", "/tmp/outside.txt"); err == nil {
		t.Fatal("expected absolute conflict file to be rejected")
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

func TestGitCLIAdapter_Merge_ValidateBranch(t *testing.T) {
	adapter := NewGitCLIAdapter()
	_, err := adapter.Merge("/tmp", "")
	if err == nil {
		t.Fatal("expected error for empty branch name")
	}
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

		var changes []domain.CompactChange
		for _, line := range strings.Split(diff, "\n") {
			change := parseCompactDiffLine(line, "main.go")
			if change != nil {
				changes = append(changes, *change)
			}
		}
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
		changes := parseCompactDiffLine("", "file.go")
		if changes != nil {
			t.Errorf("expected nil for empty diff, got %+v", changes)
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

func TestValidateRemoteURL(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		wantErr bool
	}{
		{name: "https", raw: "https://github.com/user/repo", wantErr: false},
		{name: "http", raw: "http://git.example.com/repo", wantErr: false},
		{name: "empty", raw: "", wantErr: true},
		{name: "control chars", raw: "https://github.com/user/repo\n", wantErr: true},
		{name: "unsupported scheme", raw: "file:///tmp/repo", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRemoteURL(tt.raw)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateRemoteURL(%q) error = %v, wantErr %v", tt.raw, err, tt.wantErr)
			}
		})
	}
}

func TestParseRepositorySnapshotStatus(t *testing.T) {
	out := strings.Join([]string{
		"# branch.oid abc123",
		"# branch.head feature/x",
		"# branch.upstream origin/feature/x",
		"# branch.ab +2 -1",
		"u UU N... 100644 100644 100644 100644 abc abc abc conflict.txt",
		"? new.txt",
	}, "\x00")

	snapshot, err := parseRepositorySnapshotStatus(out)
	if err != nil {
		t.Fatalf("parseRepositorySnapshotStatus failed: %v", err)
	}
	if snapshot.Branch != "feature/x" || !snapshot.HasUpstream || snapshot.Ahead != 2 || snapshot.Behind != 1 {
		t.Fatalf("unexpected branch snapshot: %+v", snapshot)
	}
	if !snapshot.HasConflicts || !snapshot.IsDirty {
		t.Fatalf("expected dirty conflict snapshot, got %+v", snapshot)
	}
	if snapshot.UntrackedCount != 1 {
		t.Fatalf("expected 1 untracked file, got %+v", snapshot)
	}
}

func TestIsStaleBranch(t *testing.T) {
	snapshot := domain.RepositorySnapshot{
		Branch:         "feature/old",
		HasUpstream:    true,
		LastCommitUnix: time.Now().Add(-(staleBranchThreshold + 24*time.Hour)).Unix(),
	}
	if !isStaleBranch(snapshot) {
		t.Fatal("expected stale branch")
	}

	snapshot.Branch = "main"
	if isStaleBranch(snapshot) {
		t.Fatal("did not expect default branch to be stale")
	}
}

func TestGitCLIAdapter_GetBranches_Worktree(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "monogit-git-wt-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	adapter := NewGitCLIAdapter()
	
	// Initialize git repo
	if _, err = exec.Command("git", "-C", tmpDir, "init", "-b", "main").CombinedOutput(); err != nil {
		// Fallback for older git versions where -b is not supported
		if _, err = exec.Command("git", "-C", tmpDir, "init").CombinedOutput(); err != nil {
			t.Skip("git init failed")
		}
	}

	// Git config for test environment
	exec.Command("git", "-C", tmpDir, "config", "user.name", "Test User").Run()
	exec.Command("git", "-C", tmpDir, "config", "user.email", "test@example.com").Run()

	// Make a commit to establish HEAD/main
	if err := os.WriteFile(filepath.Join(tmpDir, "file.txt"), []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}
	exec.Command("git", "-C", tmpDir, "add", ".").Run()
	exec.Command("git", "-C", tmpDir, "commit", "-m", "initial commit").Run()

	// Create and switch to feature branch in a worktree
	exec.Command("git", "-C", tmpDir, "branch", "feature-wt").Run()
	wtPath := filepath.Join(tmpDir, "wt-dir")
	if _, err = exec.Command("git", "-C", tmpDir, "worktree", "add", wtPath, "feature-wt").CombinedOutput(); err != nil {
		t.Skipf("git worktree not supported or failed: %v", err)
	}

	// Fetch branches
	branches, err := adapter.GetBranches(tmpDir)
	if err != nil {
		t.Fatalf("GetBranches failed: %v", err)
	}

	var foundMain, foundFeature bool
	for _, b := range branches {
		if b.Name == "main" {
			foundMain = true
			if b.IsWorktree {
				t.Error("expected main NOT to be marked as worktree")
			}
		}
		if b.Name == "feature-wt" {
			foundFeature = true
			if !b.IsWorktree {
				t.Error("expected feature-wt to be marked as worktree")
			}
		}
		if strings.Contains(b.Name, "+") {
			t.Errorf("branch name %q should not contain + symbol", b.Name)
		}
	}

	if !foundMain {
		t.Error("expected to find main branch")
	}
	if !foundFeature {
		t.Error("expected to find feature-wt branch")
	}

	// Try deleting the worktree branch
	out, deleteErr := adapter.DeleteBranch(tmpDir, "feature-wt")
	if deleteErr == nil {
		t.Error("expected DeleteBranch on worktree branch to fail, but it succeeded")
	} else {
		t.Logf("DeleteBranch failed as expected: %v, output: %s", deleteErr, out)
	}
}


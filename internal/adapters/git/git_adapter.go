package git

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"monogit/internal/domain"
)

const (
	gitTimeout = 30 * time.Second

	maxOutputBytes = 10 * 1024 * 1024
)

type GitCLIAdapter struct{}

func NewGitCLIAdapter() *GitCLIAdapter {
	return &GitCLIAdapter{}
}

func validatePathContainment(base, target string) error {
	absBase, err := filepath.Abs(base)
	if err != nil {
		return fmt.Errorf("resolve base path: %w", err)
	}

	absTarget, err := filepath.Abs(target)
	if err != nil {
		return fmt.Errorf("resolve target path: %w", err)
	}

	realBase, err := filepath.EvalSymlinks(absBase)
	if err != nil {
		return fmt.Errorf("eval symlinks base: %w", err)
	}

	realTarget, err := filepath.EvalSymlinks(absTarget)
	if err != nil {

		realTarget = absTarget
	}

	if !strings.HasPrefix(realTarget, realBase+string(filepath.Separator)) && realTarget != realBase {
		return fmt.Errorf("path %q escapes repository root %q", target, base)
	}

	return nil
}

func validatePattern(pattern string) error {
	if pattern == "" {
		return fmt.Errorf("empty pattern")
	}
	if strings.Contains(pattern, "..") {
		return fmt.Errorf("pattern contains path traversal: %q", pattern)
	}
	if filepath.IsAbs(pattern) {
		return fmt.Errorf("pattern must be relative: %q", pattern)
	}
	return nil
}

func validateRepoPath(path string) error {
	if path == "" {
		return fmt.Errorf("empty repository path")
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("invalid repository path: %w", err)
	}
	info, err := os.Stat(absPath)
	if err != nil {
		return fmt.Errorf("cannot access repository path: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("repository path is not a directory: %q", absPath)
	}
	return nil
}

func (a *GitCLIAdapter) GetBranch(repoPath string) (string, error) {
	out, err := a.runGit(repoPath, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", fmt.Errorf("get branch: %w", err)
	}
	return strings.TrimSpace(out), nil
}

func (a *GitCLIAdapter) IsDirty(repoPath string) (bool, error) {
	out, err := a.runGit(repoPath, "status", "--porcelain")
	if err != nil {
		return false, fmt.Errorf("get status: %w", err)
	}
	return strings.TrimSpace(out) != "", nil
}

func (a *GitCLIAdapter) GetAheadBehind(repoPath string) (ahead int, behind int, err error) {
	out, err := a.runGit(repoPath, "rev-list", "--left-right", "--count", "HEAD...@{upstream}")
	if err != nil {
		if strings.Contains(err.Error(), "no upstream") || strings.Contains(err.Error(), "unknown revision") {
			return 0, 0, nil
		}
		return 0, 0, fmt.Errorf("get ahead/behind: %w", err)
	}

	parts := strings.Fields(strings.TrimSpace(out))
	if len(parts) != 2 {
		return 0, 0, nil
	}

	ahead, err = strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("parse ahead count: %w", err)
	}

	behind, err = strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("parse behind count: %w", err)
	}

	return ahead, behind, nil
}

func (a *GitCLIAdapter) FetchAll(repoPath string) error {
	_, err := a.runGit(repoPath, "fetch", "--all")
	return err
}

func (a *GitCLIAdapter) Pull(repoPath string) (string, error) {
	return a.runGit(repoPath, "pull")
}

func (a *GitCLIAdapter) AddAndCommit(repoPath string, message string) (string, error) {
	if strings.TrimSpace(message) == "" {
		return "", fmt.Errorf("commit message cannot be empty")
	}

	_, err := a.runGit(repoPath, "add", ".")
	if err != nil {
		return "", fmt.Errorf("git add: %w", err)
	}

	return a.runGit(repoPath, "commit", "-m", message)
}

func (a *GitCLIAdapter) GetStatusFiles(repoPath string) ([]domain.FileStatus, error) {
	out, err := a.runGit(repoPath, "status", "--porcelain", "-z")
	if err != nil {
		return nil, fmt.Errorf("get status files: %w", err)
	}

	parts := strings.Split(out, "\x00")
	files := make([]domain.FileStatus, 0, len(parts))

	for i := 0; i < len(parts); i++ {
		line := parts[i]
		if len(line) < 4 {
			continue
		}
		xy := line[:2]
		name := line[3:]

		if xy[0] == 'R' || xy[0] == 'C' {
			if i+1 < len(parts) {
				i++
				name = parts[i]
			}
		}

		f := domain.FileStatus{Name: name}
		switch xy[0] {
		case 'M', 'A', 'D', 'R', 'C':
			f.Staged = true
		case '?':
			f.Untracked = true
		}
		if xy[1] == 'M' || xy[1] == 'D' {
			f.Modified = true
		}
		files = append(files, f)
	}
	return files, nil
}

func (a *GitCLIAdapter) GetDiff(repoPath string, f domain.FileStatus) (string, error) {
	if out, err := a.runGit(repoPath, "diff", "--cached", "--color=never", "--", f.Name); err == nil && strings.TrimSpace(out) != "" {
		return strings.TrimSpace(out), nil
	}

	if out, err := a.runGit(repoPath, "diff", "--color=never", "--", f.Name); err == nil && strings.TrimSpace(out) != "" {
		return strings.TrimSpace(out), nil
	}

	if f.Untracked {
		targetPath := filepath.Join(repoPath, f.Name)
		if err := validatePathContainment(repoPath, targetPath); err != nil {
			return "", fmt.Errorf("security: %w", err)
		}

		content, err := os.ReadFile(targetPath)
		if err == nil {
			return a.synthesizeNewFileDiff(f.Name, string(content)), nil
		}
	}

	return "No changes detected in " + f.Name, nil
}

func (a *GitCLIAdapter) synthesizeNewFileDiff(name, content string) string {
	var sb strings.Builder
	sb.Grow(len(content) + 128)

	fmt.Fprintf(&sb, "diff --git a/%s b/%s\n", name, name)
	sb.WriteString("new file mode 100644\n")
	sb.WriteString("--- /dev/null\n")
	fmt.Fprintf(&sb, "+++ b/%s\n", name)
	sb.WriteString("@@ -0,0 +1 @@\n")
	for _, l := range strings.Split(strings.TrimRight(content, "\n"), "\n") {
		sb.WriteString("+")
		sb.WriteString(l)
		sb.WriteString("\n")
	}
	return strings.TrimSpace(sb.String())
}

func (a *GitCLIAdapter) DiscardChanges(repoPath string, f domain.FileStatus) error {
	if f.Untracked {
		targetPath := filepath.Join(repoPath, f.Name)
		if err := validatePathContainment(repoPath, targetPath); err != nil {
			return fmt.Errorf("security: %w", err)
		}
		return os.RemoveAll(targetPath)
	}
	_, err := a.runGit(repoPath, "restore", "--", f.Name)
	return err
}

func (a *GitCLIAdapter) GetBranches(repoPath string) ([]string, error) {
	out, err := a.runGit(repoPath, "branch", "-a", "--format=%(refname:short)")
	if err != nil {
		return nil, err
	}

	seen := make(map[string]bool)
	var branches []string
	for _, b := range strings.Split(strings.TrimSpace(out), "\n") {
		b = strings.TrimPrefix(b, "origin/")
		if b != "" && !seen[b] {
			branches = append(branches, b)
			seen[b] = true
		}
	}
	return branches, nil
}

func (a *GitCLIAdapter) Push(repoPath string) (string, error) {
	return a.runGit(repoPath, "push")
}

func (a *GitCLIAdapter) CheckoutBranch(repoPath string, name string) error {
	_, err := a.runGit(repoPath, "checkout", name)
	return err
}

func (a *GitCLIAdapter) CreateBranch(repoPath string, name string) error {
	_, err := a.runGit(repoPath, "checkout", "-b", name)
	return err
}

func (a *GitCLIAdapter) Stash(repoPath string, message string) (string, error) {
	return a.runGit(repoPath, "stash", "push", "-u", "-m", message)
}

func (a *GitCLIAdapter) StashPop(repoPath string) (string, error) {
	return a.runGit(repoPath, "stash", "pop")
}

func (a *GitCLIAdapter) UnstageAll(repoPath string) error {
	_, err := a.runGit(repoPath, "reset", "HEAD", "--", ".")
	return err
}

func (a *GitCLIAdapter) UnstageFile(repoPath string, fileName string) error {
	_, err := a.runGit(repoPath, "reset", "HEAD", "--", fileName)
	return err
}

func (a *GitCLIAdapter) UndoCommit(repoPath string) error {
	_, err := a.runGit(repoPath, "reset", "--soft", "HEAD~1")
	return err
}

func (a *GitCLIAdapter) StageByPattern(repoPath string, pattern string) error {
	if err := validatePattern(pattern); err != nil {
		return fmt.Errorf("invalid staging pattern: %w", err)
	}

	if pattern != "." && !strings.ContainsAny(pattern, "*?[]") {
		pattern = "*" + pattern + "*"
	}
	_, err := a.runGit(repoPath, "add", pattern)
	return err
}

func (a *GitCLIAdapter) GetGraphLog(repoPath string, n int) (string, error) {
	return a.runGit(repoPath, "log", "--graph", "--format=%h||%D||%s||%ar||%an", "--all", "--decorate", "--color=never", fmt.Sprintf("-%d", n))
}

func (a *GitCLIAdapter) GetSimpleLog(repoPath string, n int) (string, error) {
	return a.runGit(repoPath, "log", "--oneline", fmt.Sprintf("-%d", n))
}

func (a *GitCLIAdapter) runGit(dir string, args ...string) (string, error) {
	if err := validateRepoPath(dir); err != nil {
		return "", fmt.Errorf("security: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), gitTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")

	var buf bytes.Buffer
	limitedBuf := &limitedWriter{buf: &buf, max: maxOutputBytes}
	cmd.Stdout = limitedBuf
	cmd.Stderr = limitedBuf

	err := cmd.Run()
	if err != nil {
		return buf.String(), fmt.Errorf("%s: %s", err, strings.TrimSpace(buf.String()))
	}
	return buf.String(), nil
}

type limitedWriter struct {
	buf     *bytes.Buffer
	max     int64
	written int64
}

func (lw *limitedWriter) Write(p []byte) (n int, err error) {
	remaining := lw.max - lw.written
	if remaining <= 0 {
		return len(p), nil
	}
	if int64(len(p)) > remaining {
		p = p[:remaining]
	}
	n, err = lw.buf.Write(p)
	lw.written += int64(n)
	return len(p), err
}

var _ domain.GitProvider = (*GitCLIAdapter)(nil)

package git

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/JoaoOliveira889/monogit/internal/domain"
)

const (
	gitTimeout      = 30 * time.Second
	tagProbeTimeout = 5 * time.Second

	maxOutputBytes       = 10 * 1024 * 1024
	staleBranchThreshold = 30 * 24 * time.Hour
)

var globalGitSemaphore = make(chan struct{}, 10)
var globalNetworkGitSemaphore = make(chan struct{}, 3)

type GitCLIAdapter struct {
	ctx context.Context
}

func NewGitCLIAdapter() *GitCLIAdapter {
	return &GitCLIAdapter{ctx: context.Background()}
}

func NewGitCLIAdapterWithContext(ctx context.Context) *GitCLIAdapter {
	return &GitCLIAdapter{ctx: ctx}
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

	absBase = filepath.Clean(absBase)
	absTarget = filepath.Clean(absTarget)

	baseWithSep := absBase + string(filepath.Separator)
	if !strings.HasPrefix(absTarget, baseWithSep) && absTarget != absBase {
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

func validateRelativePath(path string) error {
	if err := validatePattern(path); err != nil {
		return err
	}
	cleaned := filepath.Clean(path)
	if cleaned == "." {
		return nil
	}
	if strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) || cleaned == ".." {
		return fmt.Errorf("path traversal is not allowed: %q", path)
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

func validateRemoteURL(raw string) error {
	if raw == "" {
		return fmt.Errorf("empty remote url")
	}
	if strings.ContainsAny(raw, "\r\n\t") {
		return fmt.Errorf("remote url contains control characters")
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("parse remote url: %w", err)
	}
	switch parsed.Scheme {
	case "http", "https", "ssh", "git":
		if parsed.Host == "" && parsed.Scheme != "git" {
			return fmt.Errorf("remote url missing host")
		}
		return nil
	case "":
		if strings.HasPrefix(raw, "git@") {
			return nil
		}
		return fmt.Errorf("remote url has no scheme and is not an SSH address")
	default:
		return fmt.Errorf("unsupported remote url scheme: %q", parsed.Scheme)
	}
}

var branchNameRe = regexp.MustCompile(`^[a-zA-Z0-9._\-/]+$`)
var hashRe = regexp.MustCompile(`^[a-fA-F0-9]{7,40}$`)

func validateCommitHash(hash string) error {
	if !hashRe.MatchString(hash) {
		return fmt.Errorf("invalid commit hash: %q", hash)
	}
	return nil
}


func validateBranchName(name string) error {
	if name == "" {
		return fmt.Errorf("branch name cannot be empty")
	}
	if len(name) > 255 {
		return fmt.Errorf("branch name too long (max 255): %q", name)
	}
	if strings.HasPrefix(name, "-") {
		return fmt.Errorf("branch name cannot start with hyphen: %q", name)
	}
	if strings.Contains(name, "..") {
		return fmt.Errorf("branch name cannot contain '..': %q", name)
	}
	if strings.Contains(name, "@{") {
		return fmt.Errorf("branch name contains forbidden sequence '@{': %q", name)
	}
	if strings.Contains(name, " ") || strings.Contains(name, "\t") {
		return fmt.Errorf("branch name cannot contain whitespace: %q", name)
	}
	if !branchNameRe.MatchString(name) {
		return fmt.Errorf("branch name contains invalid characters: %q", name)
	}
	if strings.HasPrefix(name, "/") || strings.HasSuffix(name, "/") {
		return fmt.Errorf("branch name cannot start or end with '/': %q", name)
	}
	if strings.Contains(name, "//") {
		return fmt.Errorf("branch name cannot contain consecutive slashes: %q", name)
	}
	if strings.HasSuffix(name, ".") {
		return fmt.Errorf("branch name cannot end with '.': %q", name)
	}
	if name == "." {
		return fmt.Errorf("branch name cannot be '.'")
	}
	for _, r := range name {
		if r > unicode.MaxASCII || (!unicode.IsPrint(r) && r != '\t') {
			return fmt.Errorf("branch name contains non-printable or non-ASCII characters: %q", name)
		}
	}
	return nil
}

func validateCommitMessage(msg string) error {
	if strings.TrimSpace(msg) == "" {
		return fmt.Errorf("commit message cannot be empty")
	}
	if strings.HasPrefix(msg, "-") {
		return fmt.Errorf("commit message cannot start with '-'")
	}
	if strings.ContainsAny(msg, "`$") {
		return fmt.Errorf("commit message contains forbidden characters")
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
		return 0, 0, nil
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

func (a *GitCLIAdapter) HasUpstream(repoPath string) (bool, error) {
	_, err := a.runGit(repoPath, "rev-parse", "--symbolic-full-name", "@{u}")
	if err != nil {
		return false, nil
	}
	return true, nil
}

func (a *GitCLIAdapter) GetQuickSnapshot(repoPath string) (domain.RepositorySnapshot, error) {
	ch := make(chan struct {
		name string
		val  string
		err  error
	}, 2)

	go func() {
		out, err := a.runGit(repoPath, "status", "--porcelain=v2", "--branch", "-z")
		ch <- struct {
			name string
			val  string
			err  error
		}{"status", out, err}
	}()

	go func() {
		out, err := a.runGit(repoPath, "log", "--format=%h %s||%ct", "-1")
		ch <- struct {
			name string
			val  string
			err  error
		}{"lastCommit", out, err}
	}()

	var snapshot domain.RepositorySnapshot
	var statusErr error
	var statusDone, commitDone bool

	for !statusDone || !commitDone {
		r := <-ch
		switch r.name {
		case "status":
			statusDone = true
			if r.err != nil {
				statusErr = fmt.Errorf("get quick snapshot: %w", r.err)
				continue
			}
			snapshot, statusErr = parseRepositorySnapshotStatus(r.val)
		case "lastCommit":
			commitDone = true
			if r.err != nil {
				if strings.Contains(r.err.Error(), "does not have any commits yet") || strings.Contains(r.err.Error(), "your current branch") {
					snapshot.LastCommit = "(no commits yet)"
					break
				}
				return domain.RepositorySnapshot{}, fmt.Errorf("get quick snapshot last commit: %w", r.err)
			}
			snapshot.LastCommit, snapshot.LastCommitUnix = parseLastCommitLine(strings.TrimSpace(r.val))
		}
	}

	if statusErr != nil {
		return domain.RepositorySnapshot{}, statusErr
	}
	return snapshot, nil
}

func (a *GitCLIAdapter) GetRepositorySnapshot(repoPath string, viewGraph bool, logLines int) (domain.RepositorySnapshot, error) {
	if logLines < 1 {
		logLines = 1
	}

	type result struct {
		name string
		val  string
		err  error
	}

	ch := make(chan result, 3)

	go func() {
		out, err := a.runGit(repoPath, "status", "--porcelain=v2", "--branch", "-z")
		ch <- result{"status", out, err}
	}()

	go func() {
		out, err := a.runGit(repoPath, "log", "--format=%h %s||%ct", "-1")
		ch <- result{"lastCommit", out, err}
	}()

	go func() {
		if viewGraph {
			out, err := a.GetGraphLog(repoPath, logLines)
			ch <- result{"log", out, err}
		} else {
			out, err := a.GetSimpleLog(repoPath, logLines)
			ch <- result{"log", out, err}
		}
	}()

	var snapshot domain.RepositorySnapshot
	var statusErr, logErr, commitErr error
	var statusDone, logDone, commitDone bool

	for !statusDone || !logDone || !commitDone {
		r := <-ch
		switch r.name {
		case "status":
			statusDone = true
			if r.err != nil {
				statusErr = fmt.Errorf("get repository snapshot: %w", r.err)
				continue
			}
			snapshot, statusErr = parseRepositorySnapshotStatus(r.val)
		case "lastCommit":
			commitDone = true
			commitErr = r.err
			if commitErr != nil {
				if strings.Contains(commitErr.Error(), "does not have any commits yet") || strings.Contains(commitErr.Error(), "your current branch") {
					commitErr = nil
					r.val = "(no commits yet)"
				}
			}
			snapshot.LastCommit, snapshot.LastCommitUnix = parseLastCommitLine(strings.TrimSpace(r.val))
		case "log":
			logDone = true
			logErr = r.err
			if logErr != nil {
				if strings.Contains(logErr.Error(), "does not have any commits yet") || strings.Contains(logErr.Error(), "your current branch") {
					logErr = nil
					r.val = ""
				}
			}
			snapshot.Log = strings.TrimSpace(r.val)
			snapshot.LogGraph = viewGraph
		}
	}

	if statusErr != nil {
		return domain.RepositorySnapshot{}, statusErr
	}

	snapshot.IsStale = isStaleBranch(snapshot)

	return snapshot, nil
}

func (a *GitCLIAdapter) FetchAll(repoPath string) error {
	_, err := a.runGit(repoPath, "fetch", "--all", "--prune")
	return err
}

func (a *GitCLIAdapter) Pull(repoPath string) (string, error) {
	return a.runGit(repoPath, "pull", "--prune")
}

func (a *GitCLIAdapter) AddAndCommit(repoPath string, message string) (string, error) {
	if err := validateCommitMessage(message); err != nil {
		return "", fmt.Errorf("invalid commit message: %w", err)
	}

	_, err := a.runGit(repoPath, "add", ".")
	if err != nil {
		return "", fmt.Errorf("git add: %w", err)
	}

	return a.runGit(repoPath, "commit", "-m", message)
}

func (a *GitCLIAdapter) Commit(repoPath string, message string) (string, error) {
	if err := validateCommitMessage(message); err != nil {
		return "", fmt.Errorf("invalid commit message: %w", err)
	}
	return a.runGit(repoPath, "commit", "-m", message)
}

func (a *GitCLIAdapter) CherryPick(repoPath string, hash string) (string, error) {
	if err := validateCommitHash(hash); err != nil {
		return "", err
	}
	return a.runGit(repoPath, "cherry-pick", hash)
}

func (a *GitCLIAdapter) Revert(repoPath string, hash string) (string, error) {
	if err := validateCommitHash(hash); err != nil {
		return "", err
	}
	return a.runGit(repoPath, "revert", "--no-edit", hash)
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
	trimmed := strings.TrimRight(content, "\n")
	if trimmed != "" {
		for _, l := range strings.Split(trimmed, "\n") {
			sb.WriteString("+")
			sb.WriteString(l)
			sb.WriteString("\n")
		}
	}
	return strings.TrimSpace(sb.String())
}

func (a *GitCLIAdapter) DiscardChanges(repoPath string, f domain.FileStatus) error {
	if f.Untracked {
		if err := validateRelativePath(f.Name); err != nil {
			return fmt.Errorf("security: %w", err)
		}
		targetPath := filepath.Join(repoPath, f.Name)
		if err := validatePathContainment(repoPath, targetPath); err != nil {
			return fmt.Errorf("security: %w", err)
		}
		return os.RemoveAll(targetPath)
	}
	_, err := a.runGit(repoPath, "restore", "--", f.Name)
	return err
}

func (a *GitCLIAdapter) GetBranches(repoPath string) ([]domain.BranchInfo, error) {
	out, err := a.runGit(repoPath, "branch", "-a")
	if err != nil {
		return nil, err
	}

	branchMap := make(map[string]*domain.BranchInfo)
	var branchNames []string

	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		line = strings.TrimSpace(line)
		isCurrent := strings.HasPrefix(line, "*")
		b := strings.TrimSpace(strings.TrimPrefix(line, "*"))
		if b == "" {
			continue
		}

		isRemote := strings.HasPrefix(b, "remotes/") || strings.HasPrefix(b, "origin/")
		name := b
		if isRemote {
			name = strings.TrimPrefix(name, "remotes/")
			name = strings.TrimPrefix(name, "origin/")
		}

		if name == "" || strings.Contains(name, "HEAD ->") {
			continue
		}

		if info, ok := branchMap[name]; ok {
			if isRemote {
				info.IsRemote = true
			} else {
				info.IsLocal = true
			}
			if isCurrent {
				info.IsCurrent = true
			}
		} else {
			info := &domain.BranchInfo{
				Name:      name,
				IsRemote:  isRemote,
				IsLocal:   !isRemote,
				IsCurrent: isCurrent,
			}
			branchMap[name] = info
			branchNames = append(branchNames, name)
		}
	}

	var branches []domain.BranchInfo
	for _, name := range branchNames {
		branches = append(branches, *branchMap[name])
	}

	return branches, nil
}

func (a *GitCLIAdapter) Merge(repoPath string, branch string) (string, error) {
	if err := validateBranchName(branch); err != nil {
		return "", fmt.Errorf("invalid branch name: %w", err)
	}
	return a.runGit(repoPath, "merge", branch)
}

func (a *GitCLIAdapter) Push(repoPath string) (string, error) {
	branch, err := a.GetBranch(repoPath)
	if err != nil {
		return "", err
	}

	_, err = a.runGit(repoPath, "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{u}")
	if err != nil {
		remote := "origin"
		remotesOut, err := a.runGit(repoPath, "remote")
		if err == nil {
			remotes := strings.Fields(remotesOut)
			if len(remotes) > 0 {
				remote = remotes[0]
			}
		}
		return a.runGit(repoPath, "push", "--set-upstream", remote, branch)
	}

	return a.runGit(repoPath, "push")
}

func (a *GitCLIAdapter) CreateTag(repoPath string, name string, message string) (string, error) {
	if err := validateBranchName(name); err != nil {
		return "", fmt.Errorf("invalid tag name: %w", err)
	}
	if err := validateCommitMessage(message); err != nil {
		return "", fmt.Errorf("invalid tag message: %w", err)
	}
	return a.runGit(repoPath, "tag", "-a", name, "-m", message)
}

func (a *GitCLIAdapter) PushTag(repoPath string, name string) (string, error) {
	if err := validateBranchName(name); err != nil {
		return "", fmt.Errorf("invalid tag name: %w", err)
	}
	return a.runGit(repoPath, "push", "origin", name)
}
func (a *GitCLIAdapter) GetRemoteURL(repoPath string) (string, error) {
	out, err := a.runGit(repoPath, "remote", "get-url", "origin")
	if err != nil {
		return "", fmt.Errorf("get remote url: %w", err)
	}
	url := strings.TrimSpace(out)
	url = a.convertSSHToHTTPS(url)
	if err := validateRemoteURL(url); err != nil {
		return "", fmt.Errorf("invalid remote url: %w", err)
	}
	return url, nil
}

func (a *GitCLIAdapter) convertSSHToHTTPS(url string) string {
	if strings.HasPrefix(url, "ssh://") {
		return "https://" + strings.TrimPrefix(url, "ssh://")
	}
	if !strings.HasPrefix(url, "git@") {
		return url
	}

	url = strings.TrimPrefix(url, "git@")
	url = strings.Replace(url, ":", "/", 1)
	url = strings.TrimSuffix(url, ".git")

	return "https://" + url
}

func (a *GitCLIAdapter) CheckoutBranch(repoPath string, name string) error {
	if err := validateBranchName(name); err != nil {
		return fmt.Errorf("invalid branch name: %w", err)
	}
	_, err := a.runGit(repoPath, "checkout", name)
	return err
}

func (a *GitCLIAdapter) CreateBranch(repoPath string, name string) error {
	if err := validateBranchName(name); err != nil {
		return fmt.Errorf("invalid branch name: %w", err)
	}
	_, err := a.runGit(repoPath, "checkout", "-b", name)
	return err
}

func (a *GitCLIAdapter) Stash(repoPath string, message string) (string, error) {
	return a.runGit(repoPath, "stash", "push", "-u", "-m", message)
}

func (a *GitCLIAdapter) StashPop(repoPath string) (string, error) {
	return a.runGit(repoPath, "stash", "pop")
}

func (a *GitCLIAdapter) GetStashes(repoPath string) ([]domain.StashInfo, error) {
	out, err := a.runGit(repoPath, "stash", "list", "--format=%gd|%gs")
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	var stashes []domain.StashInfo
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 2)
		if len(parts) < 2 {
			continue
		}
		ref := parts[0]
		message := parts[1]

		idxStr := strings.TrimPrefix(ref, "stash@{")
		idxStr = strings.TrimSuffix(idxStr, "}")
		idx, err := strconv.Atoi(idxStr)
		if err != nil {
			continue
		}
		stashes = append(stashes, domain.StashInfo{
			Index:   idx,
			Message: message,
		})
	}
	return stashes, nil
}

func (a *GitCLIAdapter) ApplyStash(repoPath string, index int) (string, error) {
	return a.runGit(repoPath, "stash", "apply", fmt.Sprintf("stash@{%d}", index))
}

func (a *GitCLIAdapter) DropStash(repoPath string, index int) (string, error) {
	return a.runGit(repoPath, "stash", "drop", fmt.Sprintf("stash@{%d}", index))
}

func (a *GitCLIAdapter) PopStash(repoPath string, index int) (string, error) {
	return a.runGit(repoPath, "stash", "pop", fmt.Sprintf("stash@{%d}", index))
}

func (a *GitCLIAdapter) GetStashFiles(repoPath string, index int) ([]string, error) {
	out, err := a.runGit(repoPath, "stash", "show", "--name-only", fmt.Sprintf("stash@{%d}", index))
	if err != nil {
		return nil, err
	}
	if out == "" {
		return nil, nil
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	return lines, nil
}

func (a *GitCLIAdapter) GetStashFileDiff(repoPath string, index int, file string) (string, error) {
	if err := validatePattern(file); err != nil {
		return "", fmt.Errorf("security: %w", err)
	}
	out, err := a.runGit(repoPath, "stash", "show", "-p", fmt.Sprintf("stash@{%d}", index), "--", file)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
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

func (a *GitCLIAdapter) StageFiles(repoPath string, files []string) error {
	if len(files) == 0 {
		return fmt.Errorf("no files selected")
	}

	args := make([]string, 0, len(files)+2)
	args = append(args, "add", "--")
	for _, file := range files {
		if err := validatePattern(file); err != nil {
			return fmt.Errorf("invalid file path: %w", err)
		}
		args = append(args, file)
	}

	_, err := a.runGit(repoPath, args...)
	return err
}

func (a *GitCLIAdapter) GetGraphLog(repoPath string, n int) (string, error) {
	return a.runGit(repoPath, "log", "--graph", "--format=%h||%D||%s||%ar||%an", "--decorate", "--color=never", fmt.Sprintf("-%d", n))
}

func (a *GitCLIAdapter) GetSimpleLog(repoPath string, n int) (string, error) {
	return a.runGit(repoPath, "log", "--oneline", fmt.Sprintf("-%d", n))
}

func (a *GitCLIAdapter) DeleteBranch(repoPath string, name string) (string, error) {
	if err := validateBranchName(name); err != nil {
		return "", fmt.Errorf("invalid branch name: %w", err)
	}
	return a.runGit(repoPath, "branch", "-D", name)
}

func (a *GitCLIAdapter) DeleteRemoteBranch(repoPath string, remote string, name string) (string, error) {
	if err := validateBranchName(name); err != nil {
		return "", fmt.Errorf("invalid branch name: %w", err)
	}
	return a.runGit(repoPath, "push", remote, "--delete", name)
}

func isNetworkCommand(args []string) bool {
	if len(args) == 0 {
		return false
	}
	switch args[0] {
	case "fetch", "pull", "push", "ls-remote":
		return true
	}
	return false
}

func (a *GitCLIAdapter) runGit(dir string, args ...string) (string, error) {
	return a.runGitWithTimeout(dir, gitTimeout, args...)
}

func (a *GitCLIAdapter) runGitWithTimeout(dir string, timeout time.Duration, args ...string) (string, error) {
	if err := validateRepoPath(dir); err != nil {
		return "", fmt.Errorf("security: %w", err)
	}

	ctx, cancel := context.WithTimeout(a.ctx, timeout)
	defer cancel()

	sem := globalGitSemaphore
	if isNetworkCommand(args) {
		sem = globalNetworkGitSemaphore
	}

	select {
	case sem <- struct{}{}:
		defer func() { <-sem }()
	case <-ctx.Done():
		return "", ctx.Err()
	}

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

func (lw *limitedWriter) Write(p []byte) (int, error) {
	remaining := lw.max - lw.written
	if remaining <= 0 {
		return 0, nil
	}
	if int64(len(p)) > remaining {
		p = p[:remaining]
	}
	n, err := lw.buf.Write(p)
	lw.written += int64(n)
	return n, err
}

func (a *GitCLIAdapter) HasConflicts(repoPath string) (bool, error) {
	out, err := a.runGit(repoPath, "diff", "--name-only", "--diff-filter=U")
	if err != nil {
		return false, nil
	}
	return strings.TrimSpace(out) != "", nil
}

var conflictStatusFlags = map[string]string{
	"DD": "both deleted",
	"AU": "added by us",
	"UA": "added by them",
	"DU": "deleted by us",
	"UD": "deleted by them",
	"UU": "both modified",
	"AA": "both added",
}

func (a *GitCLIAdapter) ListConflictingFiles(repoPath string) ([]domain.ConflictFile, error) {
	out, err := a.runGit(repoPath, "status", "--porcelain")
	if err != nil {
		return nil, fmt.Errorf("list conflicts: %w", err)
	}

	var results []domain.ConflictFile
	seen := make(map[string]bool)
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if len(line) < 4 {
			continue
		}
		status := line[:2]
		name := strings.TrimSpace(line[3:])

		label, ok := conflictStatusFlags[status]
		if !ok {
			continue
		}
		if seen[name] {
			continue
		}
		seen[name] = true
		results = append(results, domain.ConflictFile{Name: name, Status: label})
	}
	return results, nil
}

var compactDiffRe = regexp.MustCompile(`^@@ -\d+(?:,\d+)? \+\d+(?:,\d+)? @@ (.*)$`)

func (a *GitCLIAdapter) GetCompactDiff(repoPath string, f domain.FileStatus) ([]domain.CompactChange, error) {
	out, err := a.runGit(repoPath, "diff", "--function-context", "--color=never", "--", f.Name)
	if err != nil {
		return nil, nil
	}
	if out == "" {
		return nil, nil
	}

	var changes []domain.CompactChange
	for _, line := range strings.Split(out, "\n") {
		change := parseCompactDiffLine(line, f.Name)
		if change != nil {
			changes = append(changes, *change)
		}
	}
	return changes, nil
}

func parseCompactDiffLine(line, fileName string) *domain.CompactChange {
	matches := compactDiffRe.FindStringSubmatch(line)
	if len(matches) < 2 {
		return nil
	}
	funcName := strings.TrimSpace(matches[1])
	if funcName == "" {
		return nil
	}
	lineRange := ""
	parts := strings.Split(matches[0], " ")
	for i, p := range parts {
		if strings.HasPrefix(p, "+") && i > 0 {
			lineRange = strings.TrimPrefix(p, "+")
			lineRange = strings.TrimSuffix(lineRange, " @@")
			break
		}
	}
	return &domain.CompactChange{
		FileName:     fileName,
		FunctionName: funcName,
		LineRange:    lineRange,
	}
}

func (a *GitCLIAdapter) OpenMergetool(repoPath string, tool string, file string) (domain.CommandSpec, error) {
	if err := validateRepoPath(repoPath); err != nil {
		return domain.CommandSpec{}, fmt.Errorf("security: %w", err)
	}
	if file == "" {
		return domain.CommandSpec{}, fmt.Errorf("empty conflict file")
	}
	if err := validateRelativePath(file); err != nil {
		return domain.CommandSpec{}, fmt.Errorf("security: %w", err)
	}
	targetPath := filepath.Clean(filepath.Join(repoPath, file))
	if err := validatePathContainment(repoPath, targetPath); err != nil {
		return domain.CommandSpec{}, fmt.Errorf("security: %w", err)
	}

	args := []string{"mergetool"}
	if tool != "" {
		args = append(args, "--tool="+tool)
	}
	args = append(args, "--", file)

	return domain.CommandSpec{
		Name: "git",
		Args: args,
		Dir:  repoPath,
	}, nil
}

func parseRepositorySnapshotStatus(out string) (domain.RepositorySnapshot, error) {
	var snapshot domain.RepositorySnapshot

	for _, entry := range strings.Split(out, "\x00") {
		if entry == "" {
			continue
		}
		switch {
		case strings.HasPrefix(entry, "# branch.head "):
			snapshot.Branch = strings.TrimSpace(strings.TrimPrefix(entry, "# branch.head "))
			if snapshot.Branch == "(detached)" {
				snapshot.Branch = "HEAD (detached)"
				snapshot.IsDetached = true
			}
		case strings.HasPrefix(entry, "# branch.upstream "):
			snapshot.HasUpstream = strings.TrimSpace(strings.TrimPrefix(entry, "# branch.upstream ")) != ""
		case strings.HasPrefix(entry, "# branch.ab "):
			fields := strings.Fields(entry)
			if len(fields) >= 4 {
				if ahead, err := strconv.Atoi(strings.TrimPrefix(fields[2], "+")); err == nil {
					snapshot.Ahead = ahead
				}
				if behind, err := strconv.Atoi(strings.TrimPrefix(fields[3], "-")); err == nil {
					snapshot.Behind = behind
				}
			}
		case strings.HasPrefix(entry, "? "):
			snapshot.UntrackedCount++
			snapshot.IsDirty = true
		case strings.HasPrefix(entry, "1 "), strings.HasPrefix(entry, "2 "), strings.HasPrefix(entry, "u "):
			snapshot.IsDirty = true
			if strings.HasPrefix(entry, "u ") {
				snapshot.HasConflicts = true
			}
			fields := strings.Fields(entry)
			if len(fields) < 2 {
				continue
			}
			xy := fields[1]
			if len(xy) >= 2 && (xy[0] != '.' || xy[1] != '.') {
				snapshot.ModifiedCount++
			}
		}
	}

	if snapshot.Branch == "" {
		snapshot.Branch = "HEAD"
	}
	if snapshot.IsDetached {
		snapshot.HasUpstream = false
	}

	return snapshot, nil
}

func parseLastCommitLine(raw string) (string, int64) {
	if raw == "" || raw == "(no commits yet)" {
		return raw, 0
	}
	parts := strings.Split(raw, "||")
	if len(parts) != 2 {
		return raw, 0
	}
	commit := strings.TrimSpace(parts[0])
	unixStr := strings.TrimSpace(parts[1])
	ts, err := strconv.ParseInt(unixStr, 10, 64)
	if err != nil {
		return commit, 0
	}
	return commit, ts
}

func isStaleBranch(snapshot domain.RepositorySnapshot) bool {
	if snapshot.IsDetached || !snapshot.HasUpstream || snapshot.LastCommitUnix == 0 {
		return false
	}
	branch := strings.ToLower(snapshot.Branch)
	switch branch {
	case "main", "master", "develop", "dev", "trunk":
		return false
	}
	age := time.Since(time.Unix(snapshot.LastCommitUnix, 0))
	return age >= staleBranchThreshold
}

func (a *GitCLIAdapter) HasUnpushedHeadTag(repoPath string) (bool, error) {
	out, err := a.runGit(repoPath, "tag", "--points-at", "HEAD")
	if err != nil {
		return false, err
	}
	tags := strings.Fields(strings.TrimSpace(out))
	if len(tags) == 0 {
		return false, nil
	}
	for _, tag := range tags {
		if err := validateBranchName(tag); err != nil {
			continue
		}
		remoteRef := "refs/tags/" + tag
		remoteOut, err := a.runGitWithTimeout(repoPath, tagProbeTimeout, "ls-remote", "--tags", "--refs", "origin", remoteRef)
		if err != nil {
			return false, err
		}
		if strings.TrimSpace(remoteOut) == "" {
			return true, nil
		}
	}
	return false, nil
}

var _ domain.GitProvider = (*GitCLIAdapter)(nil)

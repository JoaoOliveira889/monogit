package tui

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/JoaoOliveira889/monogit/internal/adapters/git"
	"github.com/JoaoOliveira889/monogit/internal/usecase"
)

// buildWorkspace creates three repositories: one clean, one with an unstaged
// change, and one with a staged rename.
func buildWorkspace(t *testing.T) string {
	t.Helper()

	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}

	root := t.TempDir()
	run := func(dir string, args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(cmd.Environ(),
			"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t",
			"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t",
		)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
		}
	}

	for _, name := range []string{"alpha", "beta", "gamma"} {
		dir := filepath.Join(root, name)
		if err := exec.Command("mkdir", dir).Run(); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
		run(dir, "init", "-q", ".")
		writeFile(t, filepath.Join(dir, "f.txt"), "hello "+name+"\n")
		run(dir, "add", ".")
		run(dir, "commit", "-qm", "feat: initial "+name)
	}

	writeFile(t, filepath.Join(root, "alpha", "f.txt"), "hello alpha\nlocal change\n")
	run(filepath.Join(root, "beta"), "mv", "f.txt", "renamed.txt")

	return root
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	cmd := exec.Command("sh", "-c", "cat > \"$0\"", path)
	cmd.Stdin = strings.NewReader(content)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("write %s: %v\n%s", path, err, out)
	}
}

// TestDashboardAgainstRealRepositories drives the model against repositories on
// disk so scanning, status parsing and rendering are exercised together.
func TestDashboardAgainstRealRepositories(t *testing.T) {
	root := buildWorkspace(t)

	uc := usecase.NewGitUseCase(git.NewGitCLIAdapter())
	m := NewModel(root, time.Minute, uc)
	m.showSplash = false
	m.handleResize(tea.WindowSizeMsg{Width: 140, Height: 44})

	m.Update(m.scanReposCmd(root)())

	if len(m.repos) != 3 {
		t.Fatalf("scanned %d repositories, want 3: %+v", len(m.repos), m.repos)
	}
	for i, r := range m.repos {
		m.Update(m.refreshStatusCmd(i, r.Path)())
	}

	frame := stripANSI(m.render())
	for _, name := range []string{"alpha", "beta", "gamma"} {
		if !strings.Contains(frame, name) {
			t.Errorf("frame is missing repository %q:\n%s", name, frame)
		}
	}

	// beta carries a staged rename; the file list must show the new path.
	for i, r := range m.repos {
		if r.Name != "beta" {
			continue
		}
		m.cursor = i
		files := m.fetchFilesCmd(r.Path)().(gitFilesMsg)
		if len(files.files) != 1 {
			t.Fatalf("beta reports %d files, want 1: %+v", len(files.files), files.files)
		}
		if files.files[0].Name != "renamed.txt" {
			t.Errorf("rename resolved to %q, want renamed.txt", files.files[0].Name)
		}
	}
}

func TestAttentionMotionAcrossRealRepositories(t *testing.T) {
	root := buildWorkspace(t)

	uc := usecase.NewGitUseCase(git.NewGitCLIAdapter())
	m := NewModel(root, time.Minute, uc)
	m.showSplash = false
	m.handleResize(tea.WindowSizeMsg{Width: 140, Height: 44})
	m.Update(m.scanReposCmd(root)())
	for i, r := range m.repos {
		m.Update(m.refreshStatusCmd(i, r.Path)())
	}

	// gamma is clean; alpha and beta both have changes.
	for i, r := range m.repos {
		if r.Name == "gamma" {
			m.cursor = i
		}
	}

	m.jumpToAttention(1, 1)
	if got := m.repos[m.cursor].Name; got != "alpha" {
		t.Errorf("} from gamma landed on %q, want alpha", got)
	}

	m.jumpToAttention(1, 1)
	if got := m.repos[m.cursor].Name; got != "beta" {
		t.Errorf("} from alpha landed on %q, want beta", got)
	}
}

// TestDiffViewerAgainstRealRepository opens the full-screen diff on a real
// working-tree change and checks the numbered gutter reflects the file.
func TestDiffViewerAgainstRealRepository(t *testing.T) {
	root := buildWorkspace(t)

	uc := usecase.NewGitUseCase(git.NewGitCLIAdapter())
	m := NewModel(root, time.Minute, uc)
	m.showSplash = false
	m.handleResize(tea.WindowSizeMsg{Width: 140, Height: 36})
	m.Update(m.scanReposCmd(root)())

	for i, r := range m.repos {
		if r.Name == "alpha" {
			m.cursor = i
		}
	}

	// d opens the viewer and asks for the file list.
	_, cmd := m.handleNormalKeys(keyPress("d"))
	if !m.showDiff() {
		t.Fatal("d did not open the diff viewer")
	}
	if cmd == nil {
		t.Fatal("opening the viewer did not request the file list")
	}
	m.Update(cmd())

	if len(m.files) != 1 {
		t.Fatalf("alpha reports %d files, want 1: %+v", len(m.files), m.files)
	}

	// The file list arriving triggers the diff fetch for the first file.
	m.Update(m.fetchDiffCmd(m.repos[m.cursor].Path, m.files[0])())

	if m.parsedDiff.Added == 0 {
		t.Errorf("parsed diff reports no additions: %+v", m.parsedDiff)
	}

	frame := stripANSI(m.render())
	t.Logf("\n%s", frame)

	if !strings.Contains(frame, "f.txt") {
		t.Errorf("viewer does not name the changed file:\n%s", frame)
	}
	if !strings.Contains(frame, "local change") {
		t.Errorf("viewer does not show the added line:\n%s", frame)
	}
	lines := strings.Split(m.render(), "\n")
	if !strings.Contains(stripANSI(lines[len(lines)-1]), "? help") {
		t.Errorf("footer missing from the diff viewer")
	}
}

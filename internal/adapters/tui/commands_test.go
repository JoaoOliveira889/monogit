package tui

import (
	"os"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestExportCommandLogUsesRestrictivePermissions(t *testing.T) {
	m := mkModel()
	m.rootPath = t.TempDir()
	m.appendCommandLog(CommandLogEntry{Output: "safe output"})

	msg := m.exportCommandLogCmd(m.rootPath)()
	result, ok := msg.(exportLogMsg)
	if !ok || result.err != nil {
		t.Fatalf("export failed: %#v", msg)
	}
	info, err := os.Stat(result.path)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0600 {
		t.Fatalf("expected 0600 permissions, got %o", got)
	}
}

func testCmd(t *testing.T, cmd tea.Cmd, expectedMsg string) {
	t.Helper()
	if cmd == nil {
		t.Fatalf("expected non-nil command for %s", expectedMsg)
	}

	msg := cmd()
	if msg == nil {
		t.Fatalf("expected non-nil message from %s command", expectedMsg)
	}

	switch msg.(type) {
	case errMsg:
	default:
	}
}

func TestTickCmd(t *testing.T) {
	cmd := tickCmd(10 * time.Millisecond)
	if cmd == nil {
		t.Fatal("expected non-nil tick command")
	}

	msg := cmd()
	if _, ok := msg.(tickMsg); !ok {
		t.Errorf("expected tickMsg, got %T", msg)
	}
}

func TestSpinnerTickCmd(t *testing.T) {
	cmd := spinnerTickCmd()
	if cmd == nil {
		t.Fatal("expected non-nil spinner command")
	}

	msg := cmd()
	if _, ok := msg.(spinnerTickMsg); !ok {
		t.Errorf("expected spinnerTickMsg, got %T", msg)
	}
}

func TestClearStatusCmd(t *testing.T) {
	cmd := clearStatusCmd(1)
	if cmd == nil {
		t.Fatal("expected non-nil clearStatus command")
	}

	msg := cmd()
	if _, ok := msg.(clearStatusMsg); !ok {
		t.Errorf("expected clearStatusMsg, got %T", msg)
	}
}

func TestFetchRepoCmd(t *testing.T) {
	m := mkModel()
	cmd := m.fetchRepoCmd(0, "/test/path")
	if cmd == nil {
		t.Fatal("expected non-nil fetch command")
	}

	msg := cmd()
	if _, ok := msg.(fetchDoneMsg); !ok {
		t.Errorf("expected fetchDoneMsg, got %T", msg)
	}
}

func TestPullRepoCmd(t *testing.T) {
	m := mkModel()
	cmd := m.pullRepoCmd(0, "/test/path")
	if cmd == nil {
		t.Fatal("expected non-nil pull command")
	}

	msg := cmd()
	if _, ok := msg.(pullDoneMsg); !ok {
		t.Errorf("expected pullDoneMsg, got %T", msg)
	}
}

func TestPushCmd(t *testing.T) {
	m := mkModel()
	cmd := m.pushCmd(0, "/test/path")
	if cmd == nil {
		t.Fatal("expected non-nil push command")
	}

	msg := cmd()
	if _, ok := msg.(pushDoneMsg); !ok {
		t.Errorf("expected pushDoneMsg, got %T", msg)
	}
}

func TestCommitAllCmd(t *testing.T) {
	m := mkModel()
	cmd := m.commitAllCmd(0, "/test/path", "test message")
	if cmd == nil {
		t.Fatal("expected non-nil commit command")
	}

	msg := cmd()
	if _, ok := msg.(commitDoneMsg); !ok {
		t.Errorf("expected commitDoneMsg, got %T", msg)
	}
}

func TestCommitSelectedCmd(t *testing.T) {
	m := mkModel()
	cmd := m.commitSelectedCmd(0, "/test/path", []string{"a.go"}, "test message")
	if cmd == nil {
		t.Fatal("expected non-nil commit command")
	}

	msg := cmd()
	if _, ok := msg.(commitDoneMsg); !ok {
		t.Errorf("expected commitDoneMsg, got %T", msg)
	}
}

func TestStashCmd(t *testing.T) {
	m := mkModel()
	cmd := m.stashCmd(0, "/test/path")
	if cmd == nil {
		t.Fatal("expected non-nil stash command")
	}

	msg := cmd()
	if _, ok := msg.(stashDoneMsg); !ok {
		t.Errorf("expected stashDoneMsg, got %T", msg)
	}
}

func TestStashPopCmd(t *testing.T) {
	m := mkModel()
	cmd := m.stashPopCmd(0, "/test/path")
	if cmd == nil {
		t.Fatal("expected non-nil stash pop command")
	}

	msg := cmd()
	if _, ok := msg.(stashPopDoneMsg); !ok {
		t.Errorf("expected stashPopDoneMsg, got %T", msg)
	}
}

func TestRefreshCachedRepoDetailCmdSkipsFreshCache(t *testing.T) {
	m := mkModel()
	m.detailCache["/test/path"] = repoDetailCacheEntry{updatedAt: time.Now()}

	if cmd := m.refreshCachedRepoDetailCmd(0, "/test/path"); cmd != nil {
		t.Fatal("expected fresh detail cache to skip Git commands")
	}
}

func TestRestoreFreshRepoDetailStopsLoading(t *testing.T) {
	m := mkModel()
	m.detailLoading = true
	m.detailCache["/test/path"] = repoDetailCacheEntry{
		modifiedCount: 2,
		lastCommit:    "abc test",
		log:           "abc test",
		updatedAt:     time.Now(),
	}

	if !m.restoreFreshRepoDetail("/test/path") {
		t.Fatal("expected fresh detail cache to be restored")
	}
	if m.detailLoading || m.cachedLogFor != "/test/path" || m.cachedModifiedCount != 2 {
		t.Fatalf("fresh cache not restored: loading=%v logFor=%q modified=%d", m.detailLoading, m.cachedLogFor, m.cachedModifiedCount)
	}
}

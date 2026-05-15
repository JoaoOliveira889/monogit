package tui

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)


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
		// errMsg is also acceptable
	default:
		// Accept the message type
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

func TestCommitCmd(t *testing.T) {
	m := mkModel()
	cmd := m.commitCmd(0, "/test/path", "test message")
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

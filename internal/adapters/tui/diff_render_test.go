package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/JoaoOliveira889/monogit/internal/domain"
)

func TestDiffViewerRendersFullScreen(t *testing.T) {
	m := NewModel("/tmp", 0, nil)
	m.showSplash = false
	m.repos = []domain.Repository{{Name: "monogit", Path: "/tmp/repo", Branch: "main", IsDirty: true}}
	m.invalidateFilterCache()
	m.handleResize(tea.WindowSizeMsg{Width: 120, Height: 32})

	m.setDetailView(DetailDiff)
	m.activePanel = DiffPanel
	m.files = []domain.FileStatus{
		{Name: "internal/adapters/tui/view.go", Modified: true},
		{Name: "internal/adapters/tui/palette.go", Untracked: true},
		{Name: "docs/releases/v0.4.0.md", Staged: true},
	}
	m.currentDiff = sampleDiff
	m.parsedDiff = parseUnifiedDiff(sampleDiff)
	m.refreshDiffViewport()

	frame := m.render()
	t.Logf("\n%s", stripANSI(frame))

	lines := strings.Split(frame, "\n")
	if len(lines) != 32 {
		t.Errorf("frame has %d lines, want 32", len(lines))
	}
	if !strings.Contains(stripANSI(lines[len(lines)-1]), "? help") {
		t.Errorf("footer missing: %q", stripANSI(lines[len(lines)-1]))
	}
}

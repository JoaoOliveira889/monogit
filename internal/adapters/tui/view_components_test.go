package tui

import (
	"strings"
	"testing"
)

func TestRenderFooterIncludesHelpAndVersion(t *testing.T) {
	m := mkModel()
	m.width = 120
	m.activePanel = RepoPanel

	footer := m.renderFooter()

	if !strings.Contains(footer, "?") || !strings.Contains(footer, "help") {
		t.Fatalf("expected footer to include ? help, got %q", footer)
	}
	if !strings.Contains(footer, "MonoGit 0.0.8") {
		t.Fatalf("expected footer to include version, got %q", footer)
	}
}

func TestRenderFooterPreservesVersionInNarrowWidth(t *testing.T) {
	m := mkModel()
	m.width = 30
	m.activePanel = RepoPanel

	footer := m.renderFooter()

	if !strings.Contains(footer, "?") || !strings.Contains(footer, "help") {
		t.Fatalf("expected footer to preserve ? help in narrow width, got %q", footer)
	}
	if !strings.Contains(footer, "MonoGit 0.0.8") {
		t.Fatalf("expected footer to preserve version in narrow width, got %q", footer)
	}
}

func TestRenderFooterIncludesHelpInModalState(t *testing.T) {
	m := mkModel()
	m.width = 80
	m.showConfirmModal = true

	footer := m.renderFooter()

	if !strings.Contains(footer, "help") || !strings.Contains(footer, "?") {
		t.Fatalf("expected modal footer to keep ? help visible, got %q", footer)
	}
}

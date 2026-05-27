package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"monogit/internal/domain"
	"monogit/internal/pkg/ui"
)

func TestRenderFooterIncludesHelpAndVersion(t *testing.T) {
	m := mkModel()
	m.width = 120
	m.activePanel = RepoPanel

	footer := m.renderFooter()

	if !strings.Contains(footer, "?") || !strings.Contains(footer, "help") {
		t.Fatalf("expected footer to include ? help, got %q", footer)
	}
	if !strings.Contains(footer, "MonoGit 0.0.11") {
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
	if !strings.Contains(footer, "MonoGit 0.0.11") {
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

func TestRenderHelpOverlayUsesBrandTitleAndAltSeparators(t *testing.T) {
	m := mkModel()
	m.width = 120
	m.height = 40

	help := m.renderHelpOverlay()

	if strings.Contains(help, "cmd+") {
		t.Fatalf("expected help menu to omit cmd aliases, got %q", help)
	}
	if !strings.Contains(help, "MonoGit SHORTCUTS") {
		t.Fatalf("expected help overlay title to reuse brand styling, got %q", help)
	}
	if !strings.Contains(help, " | ") {
		t.Fatalf("expected help overlay to use | separators, got %q", help)
	}
	for _, expected := range []string{"ctrl+c", "COMMIT WIZARD", "STASH MODE"} {
		if !strings.Contains(help, expected) {
			t.Fatalf("expected help overlay to include %q, got %q", expected, help)
		}
	}
}

func TestRenderHelpMenuFitsNarrowWidth(t *testing.T) {
	m := mkModel()
	m.width = 72
	m.height = 30

	help := m.renderHelpMenu(72, 30)

	for _, line := range strings.Split(help, "\n") {
		if lipgloss.Width(line) > m.width {
			t.Fatalf("expected help line to fit within width %d, got %d for %q", m.width, lipgloss.Width(line), line)
		}
	}
}

func TestViewHelpUsesMostOfTerminal(t *testing.T) {
	m := mkModel()
	m.width = 120
	m.height = 40
	m.showHelp = true

	view := m.View()
	maxLineWidth := 0
	for _, line := range strings.Split(view, "\n") {
		if w := lipgloss.Width(line); w > maxLineWidth {
			maxLineWidth = w
		}
	}

	if maxLineWidth < 100 {
		t.Fatalf("expected help overlay to use most of the terminal width, got max line width %d", maxLineWidth)
	}
}

func TestRenderTitledPanelActiveUsesBorderNotBackgroundFill(t *testing.T) {
	m := mkModel()
	panel := m.renderTitledPanel(40, 12, "Title", "body", true, lipgloss.Color(ui.ColorGit))

	if strings.Contains(panel, string(ui.ColorSelected)) {
		t.Fatalf("expected active panel not to use selected background fill, got %q", panel)
	}
	if !strings.Contains(panel, "╔") && !strings.Contains(panel, "╭") {
		t.Fatalf("expected panel border to render, got %q", panel)
	}
}

func TestRenderRepoTagsSectionSummarizesTags(t *testing.T) {
	m := mkModel()
	m.repos = []domain.Repository{{
		Name: "repo",
		Path: "/r",
		Tags: []string{"alpha", "beta", "gamma", "delta"},
	}}
	m.cursor = 0
	m.cfg.RepoTags = map[string][]string{
		"/other": {"omega"},
	}

	section := m.renderRepoTagsSection(80)

	if !strings.Contains(section, "Tags (4/4)") {
		t.Fatalf("expected tag count in section, got %q", section)
	}
	if !strings.Contains(section, "alpha") || !strings.Contains(section, "delta") {
		t.Fatalf("expected tag badges in section, got %q", section)
	}
	if strings.Contains(section, "omega") {
		t.Fatalf("did not expect tags from other repos in tag editor, got %q", section)
	}
}

func TestRenderRepoTagsSectionEditorShowsRepoTagsOnly(t *testing.T) {
	m := mkModel()
	m.repos = []domain.Repository{{
		Name: "repo",
		Path: "/r",
		Tags: []string{"alpha", "beta"},
	}}
	m.cursor = 0
	m.tagAssignModal = true
	m.availableTags = []string{"omega", "zeta"}

	section := m.renderRepoTagsSection(80)

	if !strings.Contains(section, "Selected:") {
		t.Fatalf("expected selected tags summary in editor, got %q", section)
	}
	if !strings.Contains(section, "alpha") || !strings.Contains(section, "beta") {
		t.Fatalf("expected current repo tags in editor, got %q", section)
	}
	if strings.Contains(section, "omega") || strings.Contains(section, "zeta") {
		t.Fatalf("expected editor to hide global tags, got %q", section)
	}
	if !strings.Contains(section, "+ New tag...") {
		t.Fatalf("expected new tag action in editor, got %q", section)
	}
}

func TestRenderTagFilterModalUsesSessionReposOnly(t *testing.T) {
	m := mkModel()
	m.repos = []domain.Repository{
		{Name: "repo-a", Path: "/a", Tags: []string{"alpha"}},
		{Name: "repo-b", Path: "/b", Tags: []string{"beta"}},
	}
	m.cfg.RepoTags = map[string][]string{
		"/p1": {"a", "c"},
	}
	m.refreshAvailableTags()

	modal := m.renderTagFilterModal(80, 30)

	if !strings.Contains(modal, "alpha") || !strings.Contains(modal, "beta") {
		t.Fatalf("expected session tags in tag filter modal, got %q", modal)
	}
	if strings.Contains(modal, " a ") || strings.Contains(modal, " c ") {
		t.Fatalf("expected modal to avoid stale config-only tags, got %q", modal)
	}
}

func TestRenderDetailPanelWrapsLongText(t *testing.T) {
	m := mkModel()
	m.width = 120
	m.height = 40
	m.repos = []domain.Repository{{
		Name:   "repo",
		Path:   "/r",
		Tags:   []string{"alpha"},
		Branch: "main",
	}}
	m.cursor = 0
	m.cachedDetailFor = "/r"
	m.cachedLogFor = "/r"
	m.cachedLastCommit = "abc1234 this is a deliberately long commit message that should wrap instead of clipping on the right side"
	m.cachedLog = "abc1234||*||main||feat||this is a deliberately long graph entry that should wrap inside the detail panel"
	m.selectedRepo()

	_, _ = m.handleResize(tea.WindowSizeMsg{Width: m.width, Height: m.height})
	m.refreshViewports()

	panel := m.renderDetailPanel(m.rightPanelWidth(), m.panelHeight())
	for _, line := range strings.Split(panel, "\n") {
		if w := lipgloss.Width(line); w > m.rightPanelWidth() {
			t.Fatalf("expected detail panel line to fit within width %d, got %d for %q", m.rightPanelWidth(), w, line)
		}
	}
}

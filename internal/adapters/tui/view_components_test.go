package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/JoaoOliveira889/monogit/internal/domain"
	"github.com/JoaoOliveira889/monogit/internal/pkg/ui"
)

func TestRenderHeaderIncludesWorkspaceHealth(t *testing.T) {
	m := mkModel()
	m.width = 120
	m.activePanel = RepoPanel

	header := m.renderHeader()

	if !strings.Contains(header, "repos") {
		t.Fatalf("expected header to include workspace health summary, got %q", header)
	}
	if strings.Contains(header, "Press ? for help") {
		t.Fatalf("expected header to omit Press ? for help, got %q", header)
	}
}

func TestRenderFooterIncludesVersion(t *testing.T) {
	m := mkModel()
	m.width = 120
	m.activePanel = RepoPanel

	footer := m.renderFooter()

	expectedVersion := "MonoGit " + Version
	if !strings.Contains(footer, expectedVersion) {
		t.Fatalf("expected footer to include version, got %q", footer)
	}
}

func TestRenderFooterPreservesVersionInNarrowWidth(t *testing.T) {
	m := mkModel()
	m.width = 30
	m.activePanel = RepoPanel

	footer := m.renderFooter()

	expectedVersion := "MonoGit " + Version
	if !strings.Contains(footer, expectedVersion) {
		t.Fatalf("expected footer to preserve version in narrow width, got %q", footer)
	}
}

func TestFooterAlwaysKeepsHelpAndVersionVisible(t *testing.T) {
	contexts := []struct {
		name  string
		setup func(*Model)
	}{
		{name: "repository", setup: func(m *Model) { m.activePanel = RepoPanel }},
		{name: "branches", setup: func(m *Model) { m.showBranches = true; m.activePanel = LogPanel }},
		{name: "files", setup: func(m *Model) { m.showFiles = true; m.activePanel = DiffPanel }},
		{name: "confirmation", setup: func(m *Model) { m.showConfirmModal = true }},
	}

	for _, tt := range contexts {
		t.Run(tt.name, func(t *testing.T) {
			m := mkModel()
			m.width = 120
			tt.setup(&m)

			footer := m.renderFooter()
			if !strings.Contains(footer, "?") || !strings.Contains(footer, "help") {
				t.Fatalf("expected persistent help hint, got %q", footer)
			}
			if !strings.Contains(footer, "MonoGit "+Version) {
				t.Fatalf("expected persistent version, got %q", footer)
			}
		})
	}
}

func TestRenderLogFooterMatchesContextualBindings(t *testing.T) {
	m := mkModel()
	m.width = 140
	m.activePanel = LogPanel

	footer := m.renderFooter()
	if !strings.Contains(footer, "enter") || !strings.Contains(footer, "details") {
		t.Fatalf("expected enter details binding, got %q", footer)
	}
	if !strings.Contains(footer, "d") || !strings.Contains(footer, "diff") {
		t.Fatalf("expected d diff binding, got %q", footer)
	}
	if !strings.Contains(footer, "y") || !strings.Contains(footer, "copy hash") {
		t.Fatalf("expected y copy hash binding, got %q", footer)
	}
}

func TestViewUsesSinglePaneCompactLayout(t *testing.T) {
	m := mkModel()
	m.showSplash = false
	m.width = 48
	m.height = 18
	m.repos = []domain.Repository{{Name: "repo", Path: "/r", Branch: "main"}}
	m.cursor = 0
	_, _ = m.handleResize(tea.WindowSizeMsg{Width: m.width, Height: m.height})

	view := m.View()
	if strings.Contains(view, "Terminal too small") {
		t.Fatalf("expected compact usable layout, got %q", view)
	}
	if !strings.Contains(view, "MonoGit "+Version) {
		t.Fatalf("expected global footer with version in compact layout, got %q", view)
	}
	for _, line := range strings.Split(view, "\n") {
		if width := lipgloss.Width(line); width > m.width {
			t.Fatalf("compact line width %d exceeds terminal width %d: %q", width, m.width, line)
		}
	}
}

func TestCompactConfigurationPanelFitsWidth(t *testing.T) {
	m := mkModel()
	m.showSplash = false
	m.width = 48
	m.height = 20
	m.activePanel = ConfigPanel
	m.repos = []domain.Repository{{Name: "repo", Path: "/r", Branch: "main"}}
	m.cfg.ScanExcludes = []string{"node_modules", "vendor", "directory-with-a-very-long-name"}
	_, _ = m.handleResize(tea.WindowSizeMsg{Width: m.width, Height: m.height})

	for _, line := range strings.Split(m.View(), "\n") {
		if width := lipgloss.Width(line); width > m.width {
			t.Fatalf("config line width %d exceeds terminal width %d: %q", width, m.width, line)
		}
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

func TestSelectedRowsUseAnIndicatorInsteadOfBackgroundFill(t *testing.T) {
	if background := ui.SelectedItemStyle.GetBackground(); !isNoColor(background) {
		t.Fatalf("selected rows must not use a background fill, got %v", background)
	}

	row := renderActiveRow("repo main", 40)
	if width := lipgloss.Width(row); width != lipgloss.Width("repo main") {
		t.Fatalf("selected row should not pad a full-width block, got width %d", width)
	}
}

func TestFooterStylesDoNotPaintBackgroundBlocks(t *testing.T) {
	if background := ui.FooterStyle.GetBackground(); !isNoColor(background) {
		t.Fatalf("footer surface must stay transparent, got %v", background)
	}
	if background := ui.FooterKeyStyle.GetBackground(); !isNoColor(background) {
		t.Fatalf("footer key must not use a background fill, got %v", background)
	}
}

func isNoColor(color lipgloss.TerminalColor) bool {
	_, ok := color.(lipgloss.NoColor)
	return ok
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

func TestRepositoryOverviewOmitsEmptyTagsSection(t *testing.T) {
	m := mkModel()
	m.width = 120
	m.height = 40
	m.repos = []domain.Repository{{Name: "repo", Path: "/r", Branch: "main"}}
	m.cursor = 0

	_, _ = m.handleResize(tea.WindowSizeMsg{Width: m.width, Height: m.height})
	panel := m.renderDetailPanel(m.rightPanelWidth(), m.panelHeight())

	if strings.Contains(panel, "No tags assigned") || strings.Contains(panel, "ctrl+t to manage tags") {
		t.Fatalf("expected empty tags to stay out of the overview, got %q", panel)
	}
	if !strings.Contains(panel, "Recent Activity") {
		t.Fatalf("expected the available space to prioritize activity, got %q", panel)
	}
}

func TestRenderFilterModalShowsStatusCategories(t *testing.T) {
	m := mkModel()
	m.repos = []domain.Repository{
		{Name: "r1", Path: "/r1", IsDirty: true},
		{Name: "r2", Path: "/r2", Ahead: 2},
		{Name: "r3", Path: "/r3", Behind: 1},
		{Name: "r4", Path: "/r4", HasConflicts: true},
		{Name: "r5", Path: "/r5", Tags: []string{"v1"}},
	}
	m.filterModal = true

	modal := m.renderFilterModal(80, 20)
	for _, expected := range []string{"All", "Dirty", "Behind", "Ahead", "Conflicts", "Tagged"} {
		if !strings.Contains(modal, expected) {
			t.Fatalf("expected filter modal to contain category %q, got %q", expected, modal)
		}
	}
}

func TestStatusFilterCategories(t *testing.T) {
	m := mkModel()
	m.repos = []domain.Repository{
		{Name: "r1", Path: "/r1", IsDirty: true},
		{Name: "r2", Path: "/r2", Ahead: 2},
		{Name: "r3", Path: "/r3", Behind: 1},
		{Name: "r4", Path: "/r4", HasConflicts: true},
		{Name: "r5", Path: "/r5", Tags: []string{"v1"}},
		{Name: "r6", Path: "/r6", Branch: "main"},
	}

	m.statusFilter = FilterDirty
	if len(m.filteredRepos()) != 1 || m.filteredRepos()[0].Name != "r1" {
		t.Fatalf("expected 1 dirty repo r1, got %v", m.filteredRepos())
	}

	m.invalidateFilterCache()
	m.statusFilter = FilterAhead
	if len(m.filteredRepos()) != 1 || m.filteredRepos()[0].Name != "r2" {
		t.Fatalf("expected 1 ahead repo r2, got %v", m.filteredRepos())
	}

	m.invalidateFilterCache()
	m.statusFilter = FilterBehind
	if len(m.filteredRepos()) != 1 || m.filteredRepos()[0].Name != "r3" {
		t.Fatalf("expected 1 behind repo r3, got %v", m.filteredRepos())
	}

	m.invalidateFilterCache()
	m.statusFilter = FilterConflicts
	if len(m.filteredRepos()) != 1 || m.filteredRepos()[0].Name != "r4" {
		t.Fatalf("expected 1 conflict repo r4, got %v", m.filteredRepos())
	}

	m.invalidateFilterCache()
	m.statusFilter = FilterTagged
	if len(m.filteredRepos()) != 1 || m.filteredRepos()[0].Name != "r5" {
		t.Fatalf("expected 1 tagged repo r5, got %v", m.filteredRepos())
	}
}

func TestHeaderOmitsZeroMetrics(t *testing.T) {
	m := mkModel()
	m.width = 120
	m.repos = []domain.Repository{
		{Name: "r1", Path: "/r1", Behind: 1},
		{Name: "r2", Path: "/r2"},
	}
	health := m.renderWorkspaceHealth()

	if !strings.Contains(health, "1 behind") {
		t.Fatalf("expected 1 behind in health summary, got %q", health)
	}
	if strings.Contains(health, "ahead") || strings.Contains(health, "dirty") || strings.Contains(health, "conflict") {
		t.Fatalf("expected zero metrics to be omitted from header, got %q", health)
	}
}

func TestRepoLineSemanticStatusLabels(t *testing.T) {
	m := mkModel()
	r := domain.Repository{
		Name:          "webapi-notifications",
		Path:          "/r",
		Branch:        "develop",
		Behind:        5,
		Ahead:         2,
		IsDirty:       true,
		ModifiedCount: 3,
	}

	line := m.renderRepoLine(0, r, 80)

	if !strings.Contains(line, "↓5 behind") {
		t.Fatalf("expected ↓5 behind in repo line, got %q", line)
	}
	if !strings.Contains(line, "↑2 ahead") {
		t.Fatalf("expected ↑2 ahead in repo line, got %q", line)
	}
	if !strings.Contains(line, "✎3 dirty") {
		t.Fatalf("expected ✎3 dirty in repo line, got %q", line)
	}
}

func TestDetailPanelTitleFormatting(t *testing.T) {
	m := mkModel()
	m.width = 120
	m.height = 40
	m.repos = []domain.Repository{{Name: "lib-shared-kernel", Path: "/lib"}}
	m.cursor = 0

	panelOverview := m.renderDetailPanel(60, 20)
	if !strings.Contains(panelOverview, "[2] Repository · lib-shared-kernel") {
		t.Fatalf("expected Overview title format [2] Repository · lib-shared-kernel, got %q", panelOverview)
	}

	m.showBranches = true
	panelBranches := m.renderDetailPanel(60, 20)
	if !strings.Contains(panelBranches, "[2] Branches · lib-shared-kernel") {
		t.Fatalf("expected Branches title format [2] Branches · lib-shared-kernel, got %q", panelBranches)
	}
}

func TestBranchesListScopeFormatting(t *testing.T) {
	m := mkModel()
	m.branches = []domain.BranchInfo{
		{Name: "main", IsLocal: true, IsRemote: true, IsCurrent: true},
		{Name: "feat/http", IsLocal: true, IsRemote: false},
	}
	m.branchCursor = 1

	out := m.renderBranchesList(60)
	if !strings.Contains(out, "local · remote") {
		t.Fatalf("expected local · remote scope alignment in branch list, got %q", out)
	}
	if !strings.Contains(out, "▶ ") {
		t.Fatalf("expected cyan pointer indicator for selected branch, got %q", out)
	}
}

func TestBranchesPanelIncludesSelectedBranchPreview(t *testing.T) {
	m := mkModel()
	m.width = 120
	m.height = 40
	m.repos = []domain.Repository{{Name: "repo", Path: "/r", Branch: "main"}}
	m.cursor = 0
	m.showBranches = true
	m.branches = []domain.BranchInfo{
		{Name: "main", IsLocal: true, IsRemote: true, IsCurrent: true},
		{Name: "feat/layout", IsLocal: true},
	}
	m.branchCursor = 1

	panel := m.renderDetailPanel(m.rightPanelWidth(), m.panelHeight())
	for _, expected := range []string{"Current", "Local", "Selected branch", "feat/layout"} {
		if !strings.Contains(panel, expected) {
			t.Fatalf("expected branch workbench to include %q, got %q", expected, panel)
		}
	}
}

func TestWideFilesWorkspaceUsesSideBySideLayout(t *testing.T) {
	m := mkModel()
	m.showSplash = false
	m.width = 140
	m.height = 40
	m.leftPanelRatio = 0.35
	m.repos = []domain.Repository{{Name: "repo", Path: "/r", Branch: "main"}}
	m.cursor = 0
	m.showFiles = true
	m.activePanel = DiffPanel
	m.files = []domain.FileStatus{{Name: "internal/adapters/tui/view_panels.go", Modified: true}}
	m.currentDiff = "@@ -1 +1 @@\n-old\n+new"

	_, _ = m.handleResize(tea.WindowSizeMsg{Width: m.width, Height: m.height})
	if !m.usesSideBySideDiff() {
		t.Fatal("expected wide terminal to use the side-by-side files and diff workspace")
	}

	workspace := m.renderFilesWorkspace(m.rightPanelWidth())
	for _, expected := range []string{"[2] Files (1)", "[3] Diff"} {
		if !strings.Contains(workspace, expected) {
			t.Fatalf("expected workspace to include %q, got %q", expected, workspace)
		}
	}
	for _, line := range strings.Split(workspace, "\n") {
		if width := lipgloss.Width(line); width > m.rightPanelWidth()-2 {
			t.Fatalf("workspace line width %d exceeds inner panel width: %q", width, line)
		}
	}
}

func TestModalWidthStaysCompactOnWideTerminal(t *testing.T) {
	m := mkModel()
	m.width = 160
	if got := m.modalWidthForContent("Create Branch\n\nEnter a new branch name"); got > 72 {
		t.Fatalf("expected compact modal width, got %d", got)
	}
}

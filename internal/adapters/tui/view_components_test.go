package tui

import (
	"fmt"
	"image/color"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

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
		{name: "branches", setup: func(m *Model) { m.setDetailView(DetailBranches); m.activePanel = LogPanel }},
		{name: "files", setup: func(m *Model) { m.setDetailView(DetailFiles); m.activePanel = DiffPanel }},
		{name: "confirmation", setup: func(m *Model) { m.pushOverlay(OverlayConfirm) }},
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

	view := m.render()
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

	for _, line := range strings.Split(m.render(), "\n") {
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
	if !strings.Contains(stripANSI(help), "MonoGit SHORTCUTS") {
		t.Fatalf("expected help overlay title to reuse brand styling, got %q", help)
	}
	if !strings.Contains(help, " | ") {
		t.Fatalf("expected help overlay to use | separators, got %q", help)
	}
	// The overlay scrolls, so the frame only has to show the first sections;
	// TestShortcutsOverlayContainsAllKeybindings covers the full reference.
	if !strings.Contains(help, "MOTIONS & PANELS") {
		t.Fatalf("expected help overlay to start at the first section, got %q", help)
	}

	menu := m.renderHelpMenu(m.width-8, 200)
	for _, expected := range []string{"ctrl+c", "COMMIT WIZARD", "STASH MODE"} {
		if !strings.Contains(menu, expected) {
			t.Fatalf("expected the shortcut reference to include %q, got %q", expected, menu)
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
	m.pushOverlay(OverlayHelp)

	view := m.render()
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
	panel := m.renderTitledPanel(40, 12, "Title", "body", true, ui.ColorGit)

	if background := ui.ActivePanelStyle.GetBackground(); !isNoColor(background) {
		t.Fatalf("active panels must be marked by their border, not a fill, got %v", background)
	}
	if strings.Contains(panel, ansiBackground(ui.ColorSelected)) {
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

func isNoColor(color color.Color) bool {
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
	m.pushOverlay(OverlayTagAssign)
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
	m.pushOverlay(OverlayStatusFilter)

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

	m.setDetailView(DetailBranches)
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
	m.setDetailView(DetailBranches)
	m.branches = []domain.BranchInfo{
		{Name: "main", IsLocal: true, IsRemote: true, IsCurrent: true},
		{Name: "feat/layout", IsLocal: true},
	}
	m.branchCursor = 1

	panel := m.renderDetailPanel(m.rightPanelWidth(), m.panelHeight())
	for _, expected := range []string{"Selected branch", "feat/layout"} {
		if !strings.Contains(panel, expected) {
			t.Fatalf("expected branch workbench to include %q, got %q", expected, panel)
		}
	}
	for _, unexpected := range []string{"Current (", "Local ("} {
		if strings.Contains(panel, unexpected) {
			t.Fatalf("expected branch workbench not to contain group header %q, got %q", unexpected, panel)
		}
	}
}

func TestBranchesListUnifiedFlatRendering(t *testing.T) {
	m := mkModel()
	m.branches = []domain.BranchInfo{
		{Name: "develop", IsLocal: true, IsRemote: true},
		{Name: "feat/public-reset", IsLocal: true},
		{Name: "feature/selfie", IsLocal: true, IsCurrent: true},
		{Name: "main", IsLocal: true, IsRemote: true},
	}
	m.branchCursor = 2

	out := m.renderBranchesList(80)
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 4 {
		t.Fatalf("expected 4 lines in flat branch list, got %d:\n%s", len(lines), out)
	}
	if !strings.Contains(lines[0], "develop") {
		t.Errorf("expected line 0 to have develop, got %q", lines[0])
	}
	if !strings.Contains(lines[1], "feat/public-reset") {
		t.Errorf("expected line 1 to have feat/public-reset, got %q", lines[1])
	}
	if !strings.Contains(lines[2], "feature/selfie") || !strings.Contains(lines[2], "▶") || !strings.Contains(lines[2], "✓") {
		t.Errorf("expected line 2 to have feature/selfie with pointer and checkmark, got %q", lines[2])
	}
	if !strings.Contains(lines[3], "main") {
		t.Errorf("expected line 3 to have main, got %q", lines[3])
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
	m.setDetailView(DetailFiles)
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

func TestShortcutsOverlayContainsAllKeybindings(t *testing.T) {
	m := mkModel()
	m.width = 140
	m.height = 40
	m.pushOverlay(OverlayHelp)

	help := m.renderHelpMenu(130, 30)

	expectedKeys := []string{
		"jk | ↑↓", "ctrl+d/u", "gg | G", "hl | ←→", "ctrl+w 1/2/3", "tab", "< | >",
		"v | y", "? | ctrl+p", "esc", "q | ctrl+c",
		"enter | l", "f | F", "p | P", "u | U", "/", "ctrl+f", "ctrl+g", "ctrl+t", "t",
		"c", "a", "v", "space", "n", "x", "z",
		"B", "Z", "e", "w", ",",
		"b", "enter", "M", "n | d", "R", "ctrl+y", "ctrl+r",
		"d", "C", "m", "gl",
		"s | S", "p | enter", "a | d", "o | E",
	}

	for _, k := range expectedKeys {
		if !strings.Contains(help, k) {
			t.Errorf("expected shortcuts menu to include key %q", k)
		}
	}
}

func TestHelpOverlayWidthNeverExceedsTerminalWidth(t *testing.T) {
	widths := []int{40, 68, 80, 105, 120, 160}
	for _, w := range widths {
		m := mkModel()
		m.width = w
		m.height = 40
		m.pushOverlay(OverlayHelp)

		overlay := m.renderHelpOverlay()
		lines := strings.Split(overlay, "\n")
		for i, line := range lines {
			lineWidth := lipgloss.Width(line)
			if lineWidth > w {
				t.Fatalf("width %d: line %d exceeded terminal width: got %d for %q", w, i, lineWidth, line)
			}
		}
	}
}

func TestHelpMenuHasNoNestedBoxBordersAndNoDoublePipes(t *testing.T) {
	m := mkModel()
	for _, w := range []int{60, 75, 110, 140} {
		help := m.renderHelpMenu(w, 30)

		// Must not contain box border corners inside viewport
		for _, corner := range []string{"╭", "╮", "╰", "╯"} {
			if strings.Contains(help, corner) {
				t.Fatalf("width %d: expected no box border corner %q in viewport content", w, corner)
			}
		}

		// Must not contain broken double vertical pipes
		if strings.Contains(help, "│ │") || strings.Contains(help, "│  │") {
			t.Fatalf("width %d: expected no double vertical pipes in help menu", w)
		}
	}
}

func TestHelpMenuSearchFilter(t *testing.T) {
	m := mkModel()
	m.width = 120
	m.height = 40
	m.pushOverlay(OverlayHelp)

	// Filter for "rebase"
	m.helpSearchInput.SetValue("rebase")
	help := m.renderHelpMenu(110, 30)
	if !strings.Contains(help, "BRANCHES & REBASE") {
		t.Errorf("expected filtered help to contain 'BRANCHES & REBASE', got: %s", help)
	}
	if strings.Contains(help, "COMMIT WIZARD") {
		t.Errorf("expected filtered help to exclude 'COMMIT WIZARD' when searching 'rebase'")
	}

	// Filter for "cherry-pick"
	m.helpSearchInput.SetValue("cherry-pick")
	help = m.renderHelpMenu(110, 30)
	if !strings.Contains(help, "ctrl+y") {
		t.Errorf("expected filtered help to contain 'ctrl+y' for cherry-pick, got: %s", help)
	}

	// Search non-existent term
	m.helpSearchInput.SetValue("xyznotfound")
	help = m.renderHelpMenu(110, 30)
	if !strings.Contains(help, "No shortcuts matching") {
		t.Errorf("expected empty state message for non-matching query, got: %s", help)
	}
}

func TestHandleHelpKeys(t *testing.T) {
	m := mkModel()
	m.width = 120
	m.height = 40
	m.pushOverlay(OverlayHelp)
	m.helpSearchInput.Focus()

	// Type a query
	newM, _ := m.handleHelpKeys(tea.KeyPressMsg{Code: 'b', Text: "b"})
	m = *newM.(*Model)
	if m.helpSearchInput.Value() != "b" {
		t.Errorf("expected helpSearchInput value 'b', got %q", m.helpSearchInput.Value())
	}
	if !m.showHelp() {
		t.Errorf("expected showHelp to remain true while typing")
	}

	// Typing normal key 'c' should update search input without triggering background commit
	newM, _ = m.handleHelpKeys(tea.KeyPressMsg{Code: 'c', Text: "c"})
	m = *newM.(*Model)
	if m.helpSearchInput.Value() != "bc" {
		t.Errorf("expected helpSearchInput value 'bc', got %q", m.helpSearchInput.Value())
	}

	// First Esc should clear the search input, but keep help open
	newM, _ = m.handleHelpKeys(tea.KeyPressMsg{Code: tea.KeyEscape})
	m = *newM.(*Model)
	if m.helpSearchInput.Value() != "" {
		t.Errorf("expected helpSearchInput to be cleared on first Esc, got %q", m.helpSearchInput.Value())
	}
	if !m.showHelp() {
		t.Errorf("expected help modal to remain open after clearing search")
	}

	// Second Esc should close the help modal
	newM, _ = m.handleHelpKeys(tea.KeyPressMsg{Code: tea.KeyEscape})
	m = *newM.(*Model)
	if m.showHelp() {
		t.Errorf("expected help modal to close on second Esc")
	}

	// Reopen help modal, then close with '?'
	m.pushOverlay(OverlayHelp)
	m.helpSearchInput.Reset()
	newM, _ = m.handleHelpKeys(tea.KeyPressMsg{Code: '?', Text: "?"})
	m = *newM.(*Model)
	if m.showHelp() {
		t.Errorf("expected help modal to close on '?'")
	}
}

func TestHandleHelpKeys_LeakedMouseSequences(t *testing.T) {
	m := mkModel()
	m.width = 120
	m.height = 40
	m.pushOverlay(OverlayHelp)
	m.helpSearchInput.Focus()

	// 1. Simulate SGR mouse wheel down sequence leaked as KeyRunes
	mouseMsg := keyText("[<65;65;10M")
	newM, _ := m.handleHelpKeys(mouseMsg)
	m = *newM.(*Model)
	if m.helpSearchInput.Value() != "" {
		t.Errorf("expected helpSearchInput to remain empty after mouse sequence, got %q", m.helpSearchInput.Value())
	}

	// 2. Another mouse sequence variant
	mouseMsg2 := keyText("[<64;65;10M")
	newM, _ = m.handleHelpKeys(mouseMsg2)
	m = *newM.(*Model)
	if m.helpSearchInput.Value() != "" {
		t.Errorf("expected helpSearchInput to remain empty, got %q", m.helpSearchInput.Value())
	}

	// 3. Single rune '[' (frequent leak when \x1b is split during fast trackpad/mouse scroll)
	bracketMsg := keyText(`[`)
	for i := 0; i < 5; i++ {
		newM, _ = m.handleHelpKeys(bracketMsg)
		m = *newM.(*Model)
	}
	if m.helpSearchInput.Value() != "" {
		t.Errorf("expected helpSearchInput to ignore '[' leaks, got %q", m.helpSearchInput.Value())
	}

	// 4. Single rune ']', '<', '>', ';'
	for _, r := range []rune{']', '<', '>', ';', '~', '\\'} {
		newM, _ = m.handleHelpKeys(keyText(string(r)))
		m = *newM.(*Model)
		if m.helpSearchInput.Value() != "" {
			t.Errorf("expected helpSearchInput to ignore rune %q, got %q", string(r), m.helpSearchInput.Value())
		}
	}

	// 5. Split wheel down sequence: "65;20;10M"
	m.height = 20             // small height so content exceeds viewport
	_ = m.renderHelpOverlay() // populate content so LineDown can advance
	initYOffset := m.helpViewport.YOffset()
	wheelDownSplit := keyText("65;20;10M")
	newM, _ = m.handleHelpKeys(wheelDownSplit)
	m = *newM.(*Model)
	if m.helpSearchInput.Value() != "" {
		t.Errorf("expected helpSearchInput to ignore wheel chunk, got %q", m.helpSearchInput.Value())
	}
	if m.helpViewport.YOffset() <= initYOffset {
		t.Errorf("expected helpViewport YOffset to advance on wheel down chunk")
	}
}

func TestRenderHelpOverlay_PreservesScrollOffset(t *testing.T) {
	m := mkModel()
	m.showSplash = false
	m.width = 100
	m.height = 25 // small height so content exceeds viewport height
	m.pushOverlay(OverlayHelp)

	// Initial render
	_ = m.renderHelpOverlay()

	// Scroll down via mouse wheel
	msg := tea.MouseWheelMsg{
		Button: tea.MouseWheelDown,
	}
	newM, _ := m.handleMouse(msg)
	m = *newM.(*Model)

	if m.helpViewport.YOffset() == 0 {
		t.Fatalf("expected helpViewport.YOffset > 0 after wheel down")
	}
	offsetBefore := m.helpViewport.YOffset()

	// Render again - must NOT reset YOffset back to 0
	_ = m.renderHelpOverlay()

	if m.helpViewport.YOffset() != offsetBefore {
		t.Errorf("expected helpViewport.YOffset to be preserved across renders, had %d, got %d", offsetBefore, m.helpViewport.YOffset())
	}
}

// ansiBackground returns the SGR sequence lipgloss emits for a background
// colour, so a rendered frame can be checked for that fill.
func ansiBackground(c color.Color) string {
	r, g, b, _ := c.RGBA()
	return fmt.Sprintf("48;2;%d;%d;%d", r>>8, g>>8, b>>8)
}

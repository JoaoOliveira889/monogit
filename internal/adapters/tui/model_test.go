package tui

import (
	"fmt"
	"html"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/JoaoOliveira889/monogit/internal/adapters/git"
	"github.com/JoaoOliveira889/monogit/internal/domain"
	"github.com/JoaoOliveira889/monogit/internal/pkg/logging"
	"github.com/JoaoOliveira889/monogit/internal/pkg/ui"
	"github.com/JoaoOliveira889/monogit/internal/testutil"
	"github.com/JoaoOliveira889/monogit/internal/usecase"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func mkModel() Model {
	uc := usecase.NewGitUseCase(&testutil.MockGitProvider{})
	return NewModel("/root", 30*time.Second, uc)
}

func TestNewModel(t *testing.T) {
	uc := usecase.NewGitUseCase(&testutil.MockGitProvider{})
	m := NewModel("/root", 30*time.Second, uc)
	if m.rootPath != "/root" {
		t.Errorf("expected /root, got %s", m.rootPath)
	}
	if m.fetchInterval != 30*time.Second {
		t.Errorf("expected 30s interval, got %v", m.fetchInterval)
	}
	if !m.scanning {
		t.Error("expected scanning to be true on init")
	}
	if m.activePanel != RepoPanel {
		t.Errorf("expected RepoPanel, got %v", m.activePanel)
	}
}

func TestSelectedRepo_Empty(t *testing.T) {
	m := mkModel()
	if r := m.selectedRepo(); r != nil {
		t.Error("expected nil for empty repos")
	}
}

func TestAppendCommandLogRedactsSecrets(t *testing.T) {
	m := mkModel()
	m.appendCommandLog(CommandLogEntry{Output: "https://user:secret@example.com/repo token=abc123"})

	got := m.commandLogs[0].Output
	if strings.Contains(got, "secret") || strings.Contains(got, "abc123") {
		t.Fatalf("expected credentials redacted, got %q", got)
	}
	if !strings.Contains(got, "[REDACTED]") {
		t.Fatalf("expected redaction marker, got %q", got)
	}
}

func TestSelectedRepo_WithRepos(t *testing.T) {
	m := mkModel()
	m.repos = []domain.Repository{{Name: "repo1", Path: "/p1"}, {Name: "repo2", Path: "/p2"}}
	m.cursor = 0

	r := m.selectedRepo()
	if r == nil {
		t.Fatal("expected non-nil")
	}
	if r.Name != "repo1" {
		t.Errorf("expected repo1, got %s", r.Name)
	}

	m.cursor = 1
	r = m.selectedRepo()
	if r.Name != "repo2" {
		t.Errorf("expected repo2, got %s", r.Name)
	}

	m.cursor = 5
	if r := m.selectedRepo(); r != nil {
		t.Error("expected nil for out-of-range cursor")
	}
}

func TestLeftPanelWidth(t *testing.T) {
	m := mkModel()
	m.leftPanelRatio = 0.3
	m.width = 100
	w := m.leftPanelWidth()
	if w != 30 {
		t.Errorf("expected 30 (30%% of 100), got %d", w)
	}

	m.width = 40
	w = m.leftPanelWidth()
	if w < 24 {
		t.Errorf("expected minimum 24, got %d", w)
	}
}

func TestRightPanelWidth(t *testing.T) {
	m := mkModel()
	m.width = 100
	w := m.rightPanelWidth()
	expected := 100 - m.leftPanelWidth()
	if w != expected {
		t.Errorf("expected %d, got %d", expected, w)
	}
}

func TestPanelHeight(t *testing.T) {
	m := mkModel()
	m.height = 50
	h := m.panelHeight()
	expected := 50 - 4
	if h != expected {
		t.Errorf("expected %d, got %d", expected, h)
	}

	m.height = 6
	h = m.panelHeight()
	if h < 5 {
		t.Errorf("expected minimum 5, got %d", h)
	}
}

func TestSelectedFiles(t *testing.T) {
	m := mkModel()
	m.files = []domain.FileStatus{
		{Name: "a.go"},
		{Name: "b.go"},
		{Name: "c.go"},
	}
	m.fileSelections[0] = true
	m.fileSelections[2] = true
	selected := m.selectedFiles()
	if len(selected) != 2 {
		t.Errorf("expected 2 selected, got %d", len(selected))
	}
	if selected[0] != "a.go" || selected[1] != "c.go" {
		t.Errorf("unexpected selected files: %v", selected)
	}
}

func TestIsBusy(t *testing.T) {
	m := mkModel()
	m.scanning = false
	if m.isBusy() {
		t.Error("expected not busy when not scanning")
	}

	m.scanning = true
	if !m.isBusy() {
		t.Error("expected busy when scanning")
	}

	m.scanning = false
	m.diffFetching = true
	if !m.isBusy() {
		t.Error("expected busy when diffFetching")
	}

	m.diffFetching = false
	m.repos = []domain.Repository{{Name: "r", Fetching: true}}
	if !m.isBusy() {
		t.Error("expected busy when repo fetching")
	}

	m.repos[0].Fetching = false
	m.repos[0].Pulling = true
	if !m.isBusy() {
		t.Error("expected busy when repo pulling")
	}
}

func TestCancelSpecialModes(t *testing.T) {
	m := mkModel()
	m.showFiles = true
	m.showBranches = true
	m.inputMode = true
	m.showHelp = true
	m.currentDiff = "diff"
	m.fileSelections[0] = true
	m.statusMsg = "msg"

	m.cancelSpecialModes()

	if m.showFiles || m.showBranches || m.inputMode || m.showHelp {
		t.Error("all modes should be cancelled")
	}
	if m.currentDiff != "" {
		t.Error("diff should be cleared")
	}
	if len(m.fileSelections) != 0 {
		t.Error("selections should be cleared")
	}
	if m.statusMsg != "" {
		t.Error("status message should be cleared")
	}
}

func TestGetVisiblePanels(t *testing.T) {
	m := mkModel()

	panels := m.GetVisiblePanels()
	if len(panels) != 2 || panels[0] != RepoPanel || panels[1] != LogPanel {
		t.Errorf("expected [RepoPanel LogPanel], got %v", panels)
	}

	m.activePanel = CommandLogPanel
	panels = m.GetVisiblePanels()
	if len(panels) != 2 || panels[1] != CommandLogPanel {
		t.Errorf("expected [RepoPanel CommandLogPanel], got %v", panels)
	}

	m.showFiles = true
	m.activePanel = LogPanel
	panels = m.GetVisiblePanels()
	if len(panels) != 3 || panels[1] != LogPanel || panels[2] != DiffPanel {
		t.Errorf("expected [RepoPanel LogPanel DiffPanel], got %v", panels)
	}

	m.showFiles = false
	m.showBranches = true
	panels = m.GetVisiblePanels()
	if len(panels) != 2 || panels[1] != LogPanel {
		t.Errorf("expected [RepoPanel LogPanel], got %v", panels)
	}
}

func TestSpinnerView(t *testing.T) {
	m := mkModel()
	m.spinnerFrame = 0
	s := m.spinnerView()
	if s == "" {
		t.Error("expected non-empty spinner")
	}

	m.spinnerFrame = 10
	s2 := m.spinnerView()
	if s == s2 {
		t.Error("expected different spinner at different frames")
	}
}

func TestRenderTitledPanelTruncation(t *testing.T) {
	m := mkModel()
	title := "This is a super extremely incredibly long panel title"
	result := m.renderTitledPanel(30, 10, title, "content", false, ui.ColorMono)

	expectedTitlePart := "This is a super extre..."
	if !strings.Contains(result, expectedTitlePart) {
		t.Errorf("expected truncated title part %q in output, got:\n%s", expectedTitlePart, result)
	}
}

func TestRenderTitledPanelTruncationRightAccent(t *testing.T) {
	m := mkModel()
	title := "Short title"
	result := m.renderTitledPanel(30, 10, title, "content", true, ui.ColorGit)

	if !strings.Contains(result, "Short title") {
		t.Fatalf("expected title to be rendered, got:\n%s", result)
	}
}

func TestRenderRepoLineBranchName(t *testing.T) {
	m := mkModel()
	r := domain.Repository{Name: "test-repo", Path: "/p1", Branch: "feature/super-cool-stuff"}

	result := m.renderRepoLine(0, r, 60)

	if !strings.Contains(result, "test-repo") {
		t.Error("expected repo name 'test-repo' in repo line output")
	}
	if !strings.Contains(result, "feature/super-cool-stuff") {
		t.Errorf("expected branch name 'feature/super-cool-stuff' in repo line output, got:\n%s", result)
	}
}

func TestRenderRepoLineBranchNameTruncation(t *testing.T) {
	m := mkModel()
	r := domain.Repository{Name: "my-repo", Path: "/p1", Branch: "feature-branch-extremely-long", HasUpstream: true}

	result := m.renderRepoLine(0, r, 25)

	if !strings.Contains(result, "my-repo") {
		t.Errorf("expected repo name 'my-repo' to be preserved, got:\n%s", result)
	}
	if !strings.Contains(result, "…") {
		t.Errorf("expected branch to be truncated with trailing ellipsis, got:\n%s", result)
	}
}

func TestRenderRepoLineHealthBadges(t *testing.T) {
	m := mkModel()
	r := domain.Repository{
		Name:           "repo",
		Path:           "/p1",
		Branch:         "feature/x",
		HasUpstream:    false,
		HasConflicts:   true,
		HasUnpushedTag: true,
	}

	result := m.renderRepoLine(0, r, 80)
	for _, want := range []string{"UP", "TG", "!"} {
		if !strings.Contains(result, want) {
			t.Fatalf("expected %q indicator in repo line, got:\n%s", want, result)
		}
	}
}

func TestRenderBeautifiedLogDebug(t *testing.T) {
	m := mkModel()
	logInput := `* 4b732e0||HEAD -> main, tag: v0.4.6, origin/main, origin/HEAD||feat: a method created to perform P2P transfers using a document and account number. (#51)||33 hours ago||Matheus Medeiros Oselame
* 5b86e07||tag: v0.4.5||feat: add hire_date on tenant proto (#50)||4 days ago||Joao Oliveira
* 7cb48ab||tag: v0.4.4||feat: add new values to return on Holder Response (#49)||6 days ago||Joao Oliveira`

	res := m.renderBeautifiedLog(logInput, 80)
	t.Logf("Beautified log result:\n%s", res)
	if res == "" {
		t.Fatal("expected non-empty result")
	}
}

func TestRealLogSnapshotUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	logging.Init()
	defer logging.Close()

	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not found")
	}

	tmpDir := t.TempDir()
	cmd := exec.Command("git", "-C", tmpDir, "init")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Skipf("git init failed: %s", string(out))
	}
	cmd = exec.Command("git", "-C", tmpDir, "config", "user.email", "test@test.com")
	cmd.Run()
	cmd = exec.Command("git", "-C", tmpDir, "config", "user.name", "Test")
	cmd.Run()
	os.WriteFile(tmpDir+"/readme.md", []byte("# test"), 0644)
	cmd = exec.Command("git", "-C", tmpDir, "add", ".")
	cmd.Run()
	commitCmd := exec.Command("git", "-C", tmpDir, "commit", "-m", "initial")
	if out, err := commitCmd.CombinedOutput(); err != nil {
		t.Skipf("git commit failed: %s", string(out))
	}

	realGit := git.NewGitCLIAdapter()
	uc := usecase.NewGitUseCase(realGit)
	m := NewModel(tmpDir, 30*time.Second, uc)

	m.repos = []domain.Repository{
		{Name: "test-repo", Path: tmpDir},
	}
	m.cursor = 0

	cmdQuick := m.refreshQuickSnapshotCmd(0, tmpDir)
	msgQuick := cmdQuick()
	resModel, _ := m.Update(msgQuick)
	m2 := resModel.(*Model)
	t.Logf("Quick snapshot: loading=%v", m2.detailLoading)

	cmdLog := m2.refreshLogSnapshotCmd(0, tmpDir, true)
	msgLog := cmdLog()
	resModel2, _ := m2.Update(msgLog)
	m3 := resModel2.(*Model)
	t.Logf("Log snapshot: logFor=%q, logLen=%d, loading=%v",
		m3.cachedLogFor, len(m3.cachedLog), m3.detailLoading)

	detailView := m3.renderDetailPanel(80, 24)
	if strings.Contains(detailView, "Loading commit history...") && m3.cachedLog != "" {
		t.Error("log is cached but view still shows loading")
	}
}

func TestScrollbarAlignment(t *testing.T) {
	vp := viewport.New(40, 5)
	vp.SetContent("Line 1\nLine 2\nLine 3\nLine 4\nLine 5\nLine 6\nLine 7\nLine 8")
	res := renderViewportWithScrollbar(vp, true)
	lines := strings.Split(res, "\n")
	if len(lines) != 5 {
		t.Fatalf("expected 5 lines, got %d", len(lines))
	}
	expectedWidth := 42
	for i, l := range lines {
		w := lipgloss.Width(l)
		if w != expectedWidth {
			t.Errorf("line %d width = %d; expected %d", i, w, expectedWidth)
		}
		// Scrollbar glyph must be preceded by a space and at the very end of the line
		if !strings.HasSuffix(l, "┃") && !strings.HasSuffix(l, "│") {
			t.Errorf("line %d does not end with scrollbar glyph: %q", i, l)
		}
	}
}

func TestGitLogLineWidths(t *testing.T) {
	m := mkModel()
	logInput := `* a56aa81||develop||fix: Remove the requirement to provide the token when calling the endpoint (#65)||19h ago||Matheus Medeiros Oselame
|\
| * a00137b||||fix: Remove the requirement to provide the token when calling the endpoint||20h ago||Matheus Medeiros Oselame
|/
* fc71b35||||feat: add aggregated P2P transfer statistics endpoint (#64)||20h ago||Matheus Medeiros Oselame`

	width := 60
	rendered := m.renderBeautifiedLog(logInput, width)
	vp := viewport.New(width, 5)
	vp.SetContent(rendered)
	withScroll := renderViewportWithScrollbar(vp, true)
	lines := strings.Split(withScroll, "\n")
	if len(lines) != 5 {
		t.Fatalf("expected 5 lines, got %d", len(lines))
	}
	expectedWidth := width + 2
	for i, l := range lines {
		w := lipgloss.Width(l)
		if w != expectedWidth {
			t.Errorf("line %d width = %d, expected %d", i, w, expectedWidth)
		}
		if !strings.HasSuffix(l, "┃") && !strings.HasSuffix(l, "│") {
			t.Errorf("line %d does not end with scrollbar glyph: %q", i, l)
		}
	}
}

func TestHelpOverlayScrollbarAlignment(t *testing.T) {
	m := mkModel()
	m.width = 110
	m.height = 25
	m.showHelp = true
	rendered := m.renderHelpOverlay()
	lines := strings.Split(rendered, "\n")
	foundScrollbar := false
	for i, l := range lines {
		if strings.Contains(l, "┃") || strings.Contains(l, "│") {
			foundScrollbar = true
			w := lipgloss.Width(l)
			if w != m.width {
				t.Errorf("rendered help line %d width = %d, expected %d", i, w, m.width)
			}
		}
	}
	if !foundScrollbar {
		t.Error("expected scrollbar in help overlay with height 25 at width 110")
	}

	// At width 130, cards fit cleanly
	m.width = 130
	renderedWide := m.renderHelpOverlay()
	for i, l := range strings.Split(renderedWide, "\n") {
		if w := lipgloss.Width(l); w != m.width {
			t.Errorf("rendered help wide line %d width = %d, expected %d", i, w, m.width)
		}
	}
}

func TestLipglossCardWidth(t *testing.T) {
	cWidth := 36
	contentW := cWidth - 4
	keyWidth := 10
	actW := contentW - keyWidth - 3
	if actW < 10 {
		actW = 10
	}

	title := lipgloss.NewStyle().Foreground(ui.ColorCyan).Bold(true).Render("NAVIGATION")
	header := fmt.Sprintf("%-*s │ %-*s", keyWidth, "KEY", actW, "ACTION")
	divider := strings.Repeat("─", keyWidth+1) + "┼" + strings.Repeat("─", actW+1)

	lines := []string{title, "", header, divider}

	entries := []struct{ k, a string }{
		{"jk | ↑↓", "Move cursor"},
		{"ctrl+d/u", "Half-page scroll"},
		{"R", "Interactive rebase"},
	}
	for _, e := range entries {
		kPadded := fmt.Sprintf("%-*s", keyWidth, e.k)
		row := fmt.Sprintf("%s │ %-*s", kPadded, actW, e.a)
		lines = append(lines, row)
	}

	style := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		Width(cWidth - 2).
		Padding(0, 1)

	rendered := style.Render(strings.Join(lines, "\n"))
	rLines := strings.Split(rendered, "\n")
	t.Logf("contentW=%d, rendered lines count=%d", contentW, len(rLines))
	for idx, rl := range rLines {
		t.Logf("rl[%d] (width=%d): %q", idx, lipgloss.Width(rl), rl)
	}
	for idx, rl := range rLines {
		if w := lipgloss.Width(rl); w != cWidth {
			t.Errorf("line %d width = %d; expected %d", idx, w, cWidth)
		}
	}
}

func ansiToHTML(s string) string {
	var out strings.Builder
	re := regexp.MustCompile(`\x1b\[([0-9;]*)m`)

	parts := re.FindAllStringSubmatchIndex(s, -1)
	lastIdx := 0
	openSpans := 0

	for _, match := range parts {
		textBefore := s[lastIdx:match[0]]
		out.WriteString(html.EscapeString(textBefore))

		codeStr := s[match[2]:match[3]]
		lastIdx = match[1]

		if codeStr == "" || codeStr == "0" {
			for openSpans > 0 {
				out.WriteString("</span>")
				openSpans--
			}
			continue
		}

		codes := strings.Split(codeStr, ";")
		for i := 0; i < len(codes); i++ {
			c, _ := strconv.Atoi(codes[i])
			switch {
			case c == 1:
				out.WriteString(`<span style="font-weight:bold;">`)
				openSpans++
			case c == 2:
				out.WriteString(`<span style="opacity:0.65;">`)
				openSpans++
			case c == 38 && i+4 < len(codes) && codes[i+1] == "2":
				r, _ := strconv.Atoi(codes[i+2])
				g, _ := strconv.Atoi(codes[i+3])
				b, _ := strconv.Atoi(codes[i+4])
				out.WriteString(fmt.Sprintf(`<span style="color:rgb(%d,%d,%d);">`, r, g, b))
				openSpans++
				i += 4
			case c == 48 && i+4 < len(codes) && codes[i+1] == "2":
				r, _ := strconv.Atoi(codes[i+2])
				g, _ := strconv.Atoi(codes[i+3])
				b, _ := strconv.Atoi(codes[i+4])
				out.WriteString(fmt.Sprintf(`<span style="background-color:rgb(%d,%d,%d);">`, r, g, b))
				openSpans++
				i += 4
			case c == 39:
				if openSpans > 0 {
					out.WriteString("</span>")
					openSpans--
				}
			}
		}
	}
	out.WriteString(html.EscapeString(s[lastIdx:]))
	for openSpans > 0 {
		out.WriteString("</span>")
		openSpans--
	}
	return out.String()
}

func wrapInTerminalHTML(title, bodyHTML string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<style>
  body {
    background-color: #13141c;
    margin: 0;
    padding: 24px;
    display: flex;
    justify-content: center;
    align-items: center;
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
  }
  .window {
    background-color: #1a1b26;
    border-radius: 12px;
    box-shadow: 0 20px 50px rgba(0, 0, 0, 0.6), 0 0 0 1px rgba(255, 255, 255, 0.08);
    overflow: hidden;
    display: inline-block;
  }
  .titlebar {
    background: linear-gradient(180deg, #1f2335 0%%, #1a1b26 100%%);
    padding: 12px 16px;
    display: flex;
    align-items: center;
    border-bottom: 1px solid #292e42;
  }
  .dots {
    display: flex;
    gap: 8px;
  }
  .dot {
    width: 12px;
    height: 12px;
    border-radius: 50%%;
  }
  .dot.red { background-color: #ff5f56; }
  .dot.yellow { background-color: #ffbd2e; }
  .dot.green { background-color: #27c93f; }
  .title {
    color: #7aa2f7;
    font-size: 13px;
    font-weight: 600;
    letter-spacing: 0.5px;
    margin: 0 auto;
    padding-right: 56px;
  }
  .content {
    padding: 16px 20px;
    font-family: "JetBrains Mono", "SF Mono", Menlo, Monaco, Consolas, monospace;
    font-size: 13px;
    line-height: 1.35;
    color: #c0caf5;
    white-space: pre;
  }
</style>
</head>
<body>
  <div class="window">
    <div class="titlebar">
      <div class="dots">
        <div class="dot red"></div>
        <div class="dot yellow"></div>
        <div class="dot green"></div>
      </div>
      <div class="title">%s</div>
    </div>
    <div class="content">%s</div>
  </div>
</body>
</html>`, html.EscapeString(title), bodyHTML)
}

func TestGenerateScreenshotHTML(t *testing.T) {
	m := mkModel()
	m.width = 112
	m.height = 27
	_, _ = m.handleResize(tea.WindowSizeMsg{Width: 112, Height: 27})
	m.scanning = false
	m.showSplash = false

	m.repos = []domain.Repository{
		{
			Name: "api-gateway", Path: "/dev/api-gateway",
			Branch: "main", HasUpstream: true,
			LastFetch: time.Now().Add(-4 * time.Minute),
		},
		{
			Name: "auth-service", Path: "/dev/auth-service",
			Branch: "feature/oauth2", Ahead: 1, IsDirty: true, ModifiedCount: 2, UntrackedCount: 1,
			HasUpstream: true,
			LastFetch:   time.Now().Add(-12 * time.Minute),
		},
		{
			Name: "billing-api", Path: "/dev/billing-api",
			Branch: "develop", HasUpstream: true,
			LastFetch: time.Now().Add(-2 * time.Minute),
		},
		{
			Name: "notification-svc", Path: "/dev/notification-svc",
			Branch: "main", Behind: 2, HasUpstream: true,
			LastFetch: time.Now().Add(-1 * time.Minute),
		},
		{
			Name: "web-portal", Path: "/dev/web-portal",
			Branch: "main", HasUpstream: true,
			LastFetch: time.Now().Add(-5 * time.Minute),
		},
	}
	m.cursor = 0

	logOutput := `* a56aa81||main||fix: validate authorization header in proxy router (#124)||2 hours ago||Alex Chen
|\
| * a00137b||feature/auth||feat: add OAuth2 PKCE challenge support||4 hours ago||Sarah Connor
|/
* fc71b35||||perf: optimize connection pool keepalive (#122)||1 day ago||Devin Vance
* 7b49e27||||refactor: decouple rate limiter middleware||2 days ago||Elena Rostova
* 698b3a2||||chore: bump dependencies to Go 1.24||3 days ago||Alex Chen
* dad55a3||||docs: update architecture overview and ADR-04||4 days ago||Alex Chen`

	m.cachedLog = logOutput
	m.cachedLogFor = m.repos[0].Path
	m.refreshViewports()

	// 1. Dashboard HTML
	dashView := m.View()
	dashHTML := wrapInTerminalHTML("MonoGit · Multi-Repository Dashboard", ansiToHTML(dashView))
	if err := os.WriteFile("/tmp/monogit_dashboard.html", []byte(dashHTML), 0644); err != nil {
		t.Fatalf("failed to write dashboard html: %v", err)
	}

	// 2. Shortcuts Modal HTML
	m.showHelp = true
	helpView := m.renderHelpOverlay()
	helpHTML := wrapInTerminalHTML("MonoGit · Shortcuts Reference", ansiToHTML(helpView))
	if err := os.WriteFile("/tmp/monogit_shortcuts.html", []byte(helpHTML), 0644); err != nil {
		t.Fatalf("failed to write shortcuts html: %v", err)
	}

	// 3. Branches View HTML
	m.showHelp = false
	m.showBranches = true
	m.branches = []domain.BranchInfo{
		{Name: "main", IsCurrent: true},
		{Name: "develop", IsCurrent: false},
		{Name: "feature/oauth2", IsCurrent: false},
		{Name: "origin/main", IsRemote: true},
		{Name: "origin/develop", IsRemote: true},
	}
	m.branchCursor = 0
	branchesView := m.View()
	branchesHTML := wrapInTerminalHTML("MonoGit · Branch Manager", ansiToHTML(branchesView))
	if err := os.WriteFile("/tmp/monogit_branches.html", []byte(branchesHTML), 0644); err != nil {
		t.Fatalf("failed to write branches html: %v", err)
	}

	t.Log("Screenshots HTML written to /tmp successfully")
}









package tui

import (
	"slices"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoaoOliveira889/monogit/internal/domain"
	"github.com/JoaoOliveira889/monogit/internal/pkg/config"
	"github.com/JoaoOliveira889/monogit/internal/pkg/ui"
	"github.com/JoaoOliveira889/monogit/internal/usecase"
)

var Version = "0.0.17"

const splashMinDuration = 2 * time.Second
const maxTagsPerRepo = 4
const maxTagLabelWidth = 14
const searchSectionHeight = 2

type Panel int

const (
	RepoPanel Panel = iota
	LogPanel
	FilePanel
	HelpPanel
	CommitWizardPanel
	DiffPanel
	CommandLogPanel
	ConflictPanel
)

type CommitStep int

const (
	StepAddOption CommitStep = iota
	StepSelectFiles
	StepMessage
)

type CommitMode int

const (
	CommitModeAll CommitMode = iota
	CommitModeSelected
)

type CommandLogEntry struct {
	Time     time.Time
	RepoName string
	Command  string
	Output   string
	Error    error
}

type Model struct {
	gitUC *usecase.GitUseCase
	cfg   config.Config

	activePanel     Panel
	previousPanel   Panel
	quitting        bool
	showSplash      bool
	splashStartedAt time.Time
	splashReady     bool
	splashFrame     int
	showHelp        bool
	viewGraph       bool
	statusMsg       string
	statusMsgID     int
	inputMode       bool
	inputAction     string

	repos         []domain.Repository
	rootPath      string
	fetchInterval time.Duration
	cursor        int

	tagFilter          []string
	tagFilterActive    bool
	tagFilterModal     bool
	tagAssignModal     bool
	tagModalCursor     int
	tagModalSelections map[int]bool
	availableTags      []string
	tagEditorRepo      string

	searchMode  bool
	searchQuery string

	cachedModifiedCount  int
	cachedUntrackedCount int
	cachedLastCommit     string
	cachedLog            string
	cachedDetailFor      string
	cachedLogFor         string
	cachedLogGraph       bool

	files          []domain.FileStatus
	fileCursor     int
	fileSelections map[int]bool
	branches       []domain.BranchInfo
	branchCursor   int
	stashes        []domain.StashInfo
	stashCursor    int
	stashFiles     []string
	showFiles      bool
	showBranches   bool
	showStashes    bool

	width        int
	height       int
	viewport     viewport.Model
	repoViewport viewport.Model
	fileViewport viewport.Model
	diffViewport viewport.Model
	spinnerFrame int

	commitStep           CommitStep
	commitMode           CommitMode
	commitInput          textinput.Model
	searchInput          textinput.Model
	showConfirmModal     bool
	confirmModalTitle    string
	confirmModalDetail   string
	confirmModalAction   string
	pendingCommitMessage string
	pendingBranchName    string
	pendingTagVersion    string
	pendingTagMessage    string
	pendingPattern       string
	pendingTagName       string

	showEditorModal  bool
	availableEditors []string
	editorCursor     int

	currentDiff   string
	diffFetching  bool
	detailLoading bool
	scanning      bool

	conflictFiles   []domain.ConflictFile
	conflictCursor  int
	showConflicts   bool
	compactDiff     bool
	compactChanges  []domain.CompactChange
	compactFetching bool

	commandLogs      []CommandLogEntry
	logViewport      viewport.Model
	helpViewport     viewport.Model
	commandLogCursor int
	selectionActive  bool
	selectionPanel   Panel
	selectionStart   int
	selectionEnd     int

	leftPanelRatio float64
}

const commitCharLimit = 200
const commitInputWidth = 50
const searchCharLimit = 100
const searchInputWidth = 30

func NewModel(rootPath string, fetchInterval time.Duration, gitUC *usecase.GitUseCase) Model {
	ti := textinput.New()
	ti.Placeholder = "Commit message..."
	ti.CharLimit = commitCharLimit
	ti.Width = commitInputWidth
	ti.PromptStyle = ui.LabelStyle
	ti.TextStyle = ui.ValueStyle

	si := textinput.New()
	si.Placeholder = "Filter repos..."
	si.CharLimit = searchCharLimit
	si.Width = searchInputWidth
	si.PromptStyle = ui.LabelStyle
	si.TextStyle = ui.ValueStyle

	cfg := config.LoadConfig()

	model := Model{
		gitUC:              gitUC,
		cfg:                cfg,
		rootPath:           rootPath,
		fetchInterval:      fetchInterval,
		commitInput:        ti,
		searchInput:        si,
		spinnerFrame:       0,
		showSplash:         true,
		splashStartedAt:    time.Now(),
		viewGraph:          true,
		fileSelections:     make(map[int]bool),
		tagModalSelections: make(map[int]bool),
		scanning:           true,
		repoViewport:       viewport.New(0, 0),
		viewport:           viewport.New(0, 0),
		fileViewport:       viewport.New(0, 0),
		diffViewport:       viewport.New(0, 0),
		logViewport:        viewport.New(0, 0),
		leftPanelRatio:     cfg.LeftPanelRatio,
	}

	model.refreshAvailableTags()
	return model
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.loadStartupReposCmd(m.rootPath),
		m.scanReposCmd(m.rootPath),
		spinnerTickCmd(),
		splashTickCmd(),
	)
}

func (m *Model) selectedRepo() *domain.Repository {
	if len(m.repos) == 0 || m.cursor < 0 || m.cursor >= len(m.repos) {
		return nil
	}
	return &m.repos[m.cursor]
}

func (m Model) leftPanelWidth() int {
	w := int(float64(m.width) * m.leftPanelRatio)
	if w < 24 {
		w = 24
	}
	if w > m.width-30 && m.width > 54 {
		w = m.width - 30
	}
	return w
}

func (m Model) rightPanelWidth() int {
	return m.width - m.leftPanelWidth() - 4
}

func (m Model) panelHeight() int {
	h := m.height - 4
	if h < 5 {
		h = 5
	}
	return h
}

func (m Model) selectedFiles() []string {
	selected := make([]string, 0, len(m.files))
	for i, f := range m.files {
		if m.fileSelections[i] {
			selected = append(selected, f.Name)
		}
	}
	return selected
}

func (m Model) isBusy() bool {
	if m.diffFetching || m.scanning {
		return true
	}
	for _, r := range m.repos {
		if r.Fetching || r.Pulling || r.Pushing || r.Stashing || r.Committing || r.CheckingOut || r.Tagging {
			return true
		}
	}
	return false
}

func (m *Model) cancelSpecialModes() {
	m.showFiles = false
	m.showBranches = false
	m.showStashes = false
	m.showConflicts = false
	m.inputMode = false
	m.showHelp = false
	m.showConfirmModal = false
	m.confirmModalTitle = ""
	m.confirmModalDetail = ""
	m.confirmModalAction = ""
	m.commitStep = StepAddOption
	m.commitMode = CommitModeAll
	m.currentDiff = ""
	m.compactDiff = false
	m.compactChanges = nil
	m.compactFetching = false
	m.fileSelections = make(map[int]bool)
	m.statusMsg = ""
	m.pendingCommitMessage = ""
	m.pendingBranchName = ""
	m.pendingTagVersion = ""
	m.pendingTagMessage = ""
	m.pendingPattern = ""
	m.pendingTagName = ""
	m.stashFiles = nil
	m.clearSelection()
	m.tagFilterModal = false
	m.tagAssignModal = false
	m.tagEditorRepo = ""
}

func (m *Model) filteredRepos() []domain.Repository {
	repos := m.repos

	if m.tagFilterActive && len(m.tagFilter) > 0 {
		tagSet := make(map[string]bool)
		for _, t := range m.tagFilter {
			tagSet[t] = true
		}
		var filtered []domain.Repository
		for _, r := range repos {
			for _, t := range r.Tags {
				if tagSet[t] {
					filtered = append(filtered, r)
					break
				}
			}
		}
		repos = filtered
	}

	if query := m.searchFilterQuery(); query != "" {
		var filtered []domain.Repository
		for _, r := range repos {
			if fuzzyMatch(query, r.Name) {
				filtered = append(filtered, r)
			}
		}
		repos = filtered
	}

	return repos
}

func (m *Model) searchFilterQuery() string {
	if m.searchMode {
		return strings.TrimSpace(m.searchInput.Value())
	}
	return m.searchQuery
}

func fuzzyMatch(query, name string) bool {
	query = strings.ToLower(strings.TrimSpace(query))
	name = strings.ToLower(name)
	if query == "" {
		return true
	}
	for _, token := range strings.Fields(query) {
		if !strings.Contains(name, token) {
			return false
		}
	}
	return true
}

func (m *Model) refreshAvailableTags() {
	tagMap := make(map[string]bool)
	for _, repo := range m.repos {
		for _, t := range repo.Tags {
			tagMap[t] = true
		}
	}
	m.availableTags = make([]string, 0, len(tagMap))
	for t := range tagMap {
		m.availableTags = append(m.availableTags, t)
	}
	slices.Sort(m.availableTags)
}

func (m *Model) repoTagCount(repoPath string) int {
	tags := m.cfg.RepoTags[repoPath]
	return len(tags)
}

func (m *Model) repoHasTag(repoPath, tag string) bool {
	for _, current := range m.cfg.RepoTags[repoPath] {
		if current == tag {
			return true
		}
	}
	return false
}

func (m *Model) syncCursorToFilter() {
	filtered := m.filteredRepos()
	if len(filtered) == 0 {
		return
	}
	if m.cursor < 0 || m.cursor >= len(m.repos) {
		for i := range m.repos {
			if m.repos[i].Path == filtered[0].Path {
				m.cursor = i
				return
			}
		}
		return
	}
	current := m.repos[m.cursor]
	for _, r := range filtered {
		if r.Path == current.Path {
			return
		}
	}
	for i := range m.repos {
		if m.repos[i].Path == filtered[0].Path {
			m.cursor = i
			return
		}
	}
}

func (m *Model) clearCommandLogs() {
	m.commandLogs = nil
	m.commandLogCursor = 0
	m.refreshLogViewport()
}

func (m *Model) clearCachedRepoDetailState() {
	m.cachedModifiedCount = 0
	m.cachedUntrackedCount = 0
	m.cachedLastCommit = ""
	m.cachedLog = ""
	m.cachedDetailFor = ""
	m.cachedLogFor = ""
	m.cachedLogGraph = false
}

func (m *Model) appendCommandLog(entry CommandLogEntry) {
	m.commandLogs = append(m.commandLogs, entry)
	const maxLogs = 120
	if len(m.commandLogs) > maxLogs {
		m.commandLogs = append([]CommandLogEntry(nil), m.commandLogs[len(m.commandLogs)-maxLogs:]...)
	}
	if m.activePanel == CommandLogPanel {
		m.refreshLogViewport()
	}
}

func (m *Model) clearSelection() {
	m.selectionActive = false
	m.selectionPanel = RepoPanel
	m.selectionStart = 0
	m.selectionEnd = 0
}

func (m *Model) beginSelection(panel Panel, index int) {
	m.selectionActive = true
	m.selectionPanel = panel
	m.selectionStart = index
	m.selectionEnd = index
}

func (m *Model) updateSelection(panel Panel, index int) {
	if !m.selectionActive || m.selectionPanel != panel {
		return
	}
	m.selectionEnd = index
}

func (m *Model) selectionBounds() (Panel, int, int, bool) {
	if !m.selectionActive {
		return RepoPanel, 0, 0, false
	}
	start := m.selectionStart
	end := m.selectionEnd
	if start > end {
		start, end = end, start
	}
	return m.selectionPanel, start, end, true
}

func (m *Model) lineSelected(panel Panel, index int) bool {
	selPanel, start, end, ok := m.selectionBounds()
	if !ok || selPanel != panel {
		return false
	}
	return index >= start && index <= end
}

func (m *Model) panelSelectionLabel() string {
	switch m.selectionPanel {
	case RepoPanel:
		return "repositories"
	case LogPanel:
		if m.showFiles {
			return "files"
		}
		if m.showBranches {
			return "branches"
		}
		if m.showStashes {
			return "stashes"
		}
		if m.showConflicts {
			return "conflicts"
		}
		return "details"
	case DiffPanel:
		return "diff"
	case CommandLogPanel:
		return "command logs"
	case ConflictPanel:
		return "conflicts"
	default:
		return "selection"
	}
}

func (m *Model) GetVisiblePanels() []Panel {
	panels := []Panel{RepoPanel}

	if m.activePanel == CommandLogPanel {
		panels = append(panels, CommandLogPanel)
	} else if m.showFiles {
		panels = append(panels, LogPanel, DiffPanel)
	} else if m.showBranches {
		panels = append(panels, LogPanel)
	} else if m.showStashes {
		panels = append(panels, LogPanel)
	} else if m.showConflicts {
		panels = append(panels, ConflictPanel)
	} else {
		panels = append(panels, LogPanel)
	}

	return panels
}

var spinnerFrames = []string{"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"}

func (m *Model) spinnerView() string {
	return spinnerFrames[m.spinnerFrame%len(spinnerFrames)]
}

var _ tea.Model = &Model{}

func (m *Model) isStatusPersistent() bool {
	if m.statusMsg == "⟳ Auto-fetching..." || m.statusMsg == "Enter commit message..." || m.scanning {
		return true
	}
	return false
}

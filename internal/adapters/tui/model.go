package tui

import (
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"monogit/internal/domain"
	"monogit/internal/pkg/config"
	"monogit/internal/pkg/ui"
	"monogit/internal/usecase"
)

var Version = "0.0.8"

const splashFrameLimit = 20

type Panel int

const (
	RepoPanel Panel = iota
	LogPanel
	FilePanel
	HelpPanel
	CommitWizardPanel
	DiffPanel
	CommandLogPanel
)

type CommitStep int

const (
	StepAddOption CommitStep = iota
	StepSelectFiles
	StepMessage
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

	activePanel   Panel
	previousPanel Panel
	quitting      bool
	showSplash    bool
	splashFrame   int
	showHelp      bool
	viewGraph     bool
	statusMsg     string
	statusMsgID   int
	inputMode     bool
	inputAction   string

	repos         []domain.Repository
	rootPath      string
	fetchInterval time.Duration
	cursor        int

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
	commitInput          textinput.Model
	showConfirmModal     bool
	confirmModalTitle    string
	confirmModalDetail   string
	confirmModalAction   string
	pendingCommitMessage string
	pendingBranchName    string
	pendingTagVersion    string
	pendingTagMessage    string
	pendingPattern       string

	showEditorModal  bool
	availableEditors []string
	editorCursor     int

	currentDiff   string
	diffFetching  bool
	detailLoading bool
	scanning      bool

	commandLogs      []CommandLogEntry
	logViewport      viewport.Model
	commandLogCursor int
	selectionActive  bool
	selectionPanel   Panel
	selectionStart   int
	selectionEnd     int

	leftPanelRatio float64
}

func NewModel(rootPath string, fetchInterval time.Duration, gitUC *usecase.GitUseCase) Model {
	ti := textinput.New()
	ti.Placeholder = "Commit message..."
	ti.CharLimit = 200
	ti.Width = 50
	ti.PromptStyle = ui.LabelStyle
	ti.TextStyle = ui.ValueStyle

	cfg := config.LoadConfig()

	return Model{
		gitUC:          gitUC,
		rootPath:       rootPath,
		fetchInterval:  fetchInterval,
		commitInput:    ti,
		spinnerFrame:   0,
		showSplash:     true,
		viewGraph:      true,
		fileSelections: make(map[int]bool),
		scanning:       true,
		repoViewport:   viewport.New(0, 0),
		viewport:       viewport.New(0, 0),
		fileViewport:   viewport.New(0, 0),
		diffViewport:   viewport.New(0, 0),
		logViewport:    viewport.New(0, 0),
		leftPanelRatio: cfg.LeftPanelRatio,
	}
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(
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

func (m Model) getStagedFiles() []string {
	staged := make([]string, 0, len(m.files))
	for _, f := range m.files {
		if f.Staged {
			staged = append(staged, f.Name)
		}
	}
	return staged
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
	m.inputMode = false
	m.showHelp = false
	m.showConfirmModal = false
	m.confirmModalTitle = ""
	m.confirmModalDetail = ""
	m.confirmModalAction = ""
	m.commitStep = StepAddOption
	m.currentDiff = ""
	m.fileSelections = make(map[int]bool)
	m.statusMsg = ""
	m.pendingCommitMessage = ""
	m.pendingBranchName = ""
	m.pendingTagVersion = ""
	m.pendingTagMessage = ""
	m.pendingPattern = ""
	m.stashFiles = nil
	m.clearSelection()
}

func (m *Model) refreshCachedRepoDetail() {
	r := m.selectedRepo()
	if r == nil {
		m.cachedModifiedCount = 0
		m.cachedUntrackedCount = 0
		m.cachedLastCommit = ""
		m.cachedLog = ""
		m.cachedDetailFor = ""
		m.cachedLogFor = ""
		return
	}
	if m.gitUC == nil {
		return
	}
	files, _ := m.gitUC.GetFiles(r.Path)
	var mod, untracked int
	for _, f := range files {
		if f.Untracked {
			untracked++
		} else if f.Modified {
			mod++
		}
	}
	m.cachedModifiedCount = mod
	m.cachedUntrackedCount = untracked

	m.cachedLastCommit, _ = m.gitUC.GetSimpleLog(r.Path, 1)
	if m.viewGraph {
		m.cachedLog, _ = m.gitUC.GetGraphLog(r.Path, 30)
		m.cachedLogGraph = true
	} else {
		m.cachedLog, _ = m.gitUC.GetSimpleLog(r.Path, 30)
		m.cachedLogGraph = false
	}
	m.cachedLogFor = r.Path
}

func (m *Model) clearCommandLogs() {
	m.commandLogs = nil
	m.commandLogCursor = 0
	m.refreshLogViewport()
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
		return "details"
	case DiffPanel:
		return "diff"
	case CommandLogPanel:
		return "command logs"
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

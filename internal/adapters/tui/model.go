package tui

import (
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"monogit/internal/domain"
	"monogit/internal/pkg/ui"
	"monogit/internal/usecase"
)

var Version = "0.0.6"

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
	showHelp      bool
	viewGraph   bool
	statusMsg   string
	statusMsgID int
	inputMode   bool
	inputAction string

	repos         []domain.Repository
	rootPath      string
	fetchInterval time.Duration
	cursor        int

	cachedModifiedCount  int
	cachedUntrackedCount int
	cachedLastCommit     string

	files          []domain.FileStatus
	fileCursor     int
	fileSelections map[int]bool
	branches       []domain.BranchInfo
	branchCursor   int
	showFiles      bool
	showBranches   bool

	width        int
	height       int
	viewport     viewport.Model
	repoViewport viewport.Model
	fileViewport viewport.Model
	diffViewport viewport.Model
	spinnerFrame int

	commitStep         CommitStep
	commitInput        textinput.Model
	showConfirmModal   bool
	confirmModalTitle  string
	confirmModalAction string

	showEditorModal  bool
	availableEditors []string
	editorCursor     int

	currentDiff  string
	tagVersion   string
	diffFetching bool
	scanning     bool

	commandLogs []CommandLogEntry
	logViewport viewport.Model
}

func NewModel(rootPath string, fetchInterval time.Duration, gitUC *usecase.GitUseCase) Model {
	ti := textinput.New()
	ti.Placeholder = "Commit message..."
	ti.CharLimit = 200
	ti.Width = 50
	ti.PromptStyle = ui.LabelStyle
	ti.TextStyle = ui.ValueStyle


	return Model{
		gitUC:          gitUC,
		rootPath:       rootPath,
		fetchInterval:  fetchInterval,
		commitInput:    ti,
		spinnerFrame:   0,
		viewGraph:      true,
		fileSelections: make(map[int]bool),
		scanning:       true,
		repoViewport:   viewport.New(0, 0),
		viewport:       viewport.New(0, 0),
		fileViewport:   viewport.New(0, 0),
		diffViewport:   viewport.New(0, 0),
		logViewport:    viewport.New(0, 0),
	}
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.scanReposCmd(m.rootPath),
		spinnerTickCmd(),
	)
}

func (m *Model) selectedRepo() *domain.Repository {
	if len(m.repos) == 0 || m.cursor < 0 || m.cursor >= len(m.repos) {
		return nil
	}
	return &m.repos[m.cursor]
}

func (m Model) leftPanelWidth() int {
	w := m.width * 30 / 100
	if w < 24 {
		w = 24
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
	m.inputMode = false
	m.showHelp = false
	m.commitStep = StepAddOption
	m.currentDiff = ""
	m.fileSelections = make(map[int]bool)
	m.statusMsg = ""
	m.tagVersion = ""
}

func (m *Model) refreshCachedRepoDetail() {
	r := m.selectedRepo()
	if r == nil {
		m.cachedModifiedCount = 0
		m.cachedUntrackedCount = 0
		m.cachedLastCommit = ""
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
}

func (m *Model) GetVisiblePanels() []Panel {
	panels := []Panel{RepoPanel}

	if m.activePanel == CommandLogPanel {
		panels = append(panels, CommandLogPanel)
	} else if m.showFiles {
		panels = append(panels, LogPanel, DiffPanel)
	} else if m.showBranches {
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


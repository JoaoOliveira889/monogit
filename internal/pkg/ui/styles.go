package ui

import "github.com/charmbracelet/lipgloss"

var (
	ColorBg        = lipgloss.Color("#1a1b26")
	ColorFg        = lipgloss.Color("#a9b1d6")
	ColorHighlight = lipgloss.Color("#7aa2f7")
	ColorSelected  = lipgloss.Color("#364a82")
	ColorAccent    = lipgloss.Color("#bb9af7")
	ColorSuccess   = lipgloss.Color("#9ece6a")
	ColorError     = lipgloss.Color("#f7768e")
	ColorWarning   = lipgloss.Color("#e0af68")
	ColorSubtle    = lipgloss.Color("#565f89")
	ColorBorder    = lipgloss.Color("#414868")
	ColorCyan      = lipgloss.Color("#7dcfff")
	ColorOrange    = lipgloss.Color("#ff9e64")
	ColorIndigo    = lipgloss.Color("#6366f1")
	ColorRose      = lipgloss.Color("#fb7185")
	ColorEmerald   = lipgloss.Color("#34d399")
	ColorAmber     = lipgloss.Color("#fbbf24")
)

const (
	IconAhead   = "↑"
	IconBehind  = "↓"
	IconDirty   = "✗"
	IconClean   = "✓"
	IconPointer = ""
	IconSpace   = " "
)

var (
	LeftPanelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			Padding(0, 0)

	ActivePanelStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorHighlight).
				Padding(0, 0)

	RightPanelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			Padding(0, 0)

	HeaderStyle = lipgloss.NewStyle().
			Background(ColorHighlight).
			Foreground(ColorBg).
			Bold(true).
			Padding(0, 1)

	FooterStyle = lipgloss.NewStyle().
			Background(ColorBg).
			Foreground(ColorSubtle).
			Padding(0, 1)

	FooterKeyStyle = lipgloss.NewStyle().
			Foreground(ColorAccent).
			Bold(true)

	FooterActionStyle = lipgloss.NewStyle().
				Foreground(ColorFg)

	PanelTitleStyle = lipgloss.NewStyle().
			Foreground(ColorHighlight).
			Bold(true).
			Padding(0, 1)

	LabelStyle = lipgloss.NewStyle().
			Foreground(ColorAccent).
			Bold(true)
)

var (
	SelectedItemStyle = lipgloss.NewStyle().
				Background(ColorHighlight).
				Foreground(ColorBg).
				Bold(true)

	NormalItemStyle = lipgloss.NewStyle().
			Foreground(ColorFg)

	PointerStyle = lipgloss.NewStyle().
			Foreground(ColorAccent).
			Bold(true)
)

var (
	BranchStyle = lipgloss.NewStyle().
			Foreground(ColorHighlight).
			Bold(true)

	AheadStyle = lipgloss.NewStyle().
			Foreground(ColorSuccess).
			Bold(true)

	BehindStyle = lipgloss.NewStyle().
			Foreground(ColorWarning).
			Bold(true)

	DirtyStyle = lipgloss.NewStyle().
			Foreground(ColorError).
			Bold(true)

	CleanStyle = lipgloss.NewStyle().
			Foreground(ColorSuccess)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorError).
			Bold(true)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(ColorSuccess).
			Bold(true)

	SubtleStyle = lipgloss.NewStyle().
			Foreground(ColorSubtle)

	SpinnerStyle = lipgloss.NewStyle().
			Foreground(ColorCyan)

	ValueStyle = lipgloss.NewStyle().
			Foreground(ColorFg)

	InputStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(ColorHighlight).
			Padding(0, 1)

	GraphColors = []lipgloss.Color{
		lipgloss.Color("#7aa2f7"),
		lipgloss.Color("#bb9af7"),
		lipgloss.Color("#7dcfff"),
		lipgloss.Color("#9ece6a"),
		lipgloss.Color("#e0af68"),
		lipgloss.Color("#f7768e"),
	}

	DiffAddStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#9ece6a"))
	DiffDelStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#f7768e"))
	DiffHunkStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#7aa2f7"))
)

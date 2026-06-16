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
	ColorMono      = lipgloss.Color("#d7d7d7")
	ColorGit       = lipgloss.Color("#b8ff3d")
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

	BrandMonoStyle = lipgloss.NewStyle().
			Foreground(ColorMono).
			Bold(true)

	BrandGitStyle = lipgloss.NewStyle().
			Foreground(ColorGit).
			Bold(true)

	BrandTitleStyle = lipgloss.NewStyle().
			Foreground(ColorFg).
			Bold(true)

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

	WarningStyle = lipgloss.NewStyle().
			Foreground(ColorWarning).
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
		lipgloss.Color("#b8ff3d"),
		lipgloss.Color("#d7d7d7"),
	}

	GraphCharStyles []lipgloss.Style

	DiffAddStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#9ece6a"))
	DiffDelStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#f7768e"))
	DiffHunkStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#7aa2f7"))
)

func ForegroundStyle(color lipgloss.Color) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(color)
}

func BackgroundStyle(color lipgloss.Color) lipgloss.Style {
	return lipgloss.NewStyle().Background(color)
}

func DiffTabStyle(active bool) lipgloss.Style {
	style := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), true, false, false, false).
		Padding(0, 1)
	if active {
		return style.Foreground(ColorHighlight).BorderForeground(ColorHighlight)
	}
	return style.Foreground(ColorSubtle).BorderForeground(ColorBorder)
}

func init() {
	GraphCharStyles = make([]lipgloss.Style, len(GraphColors))
	for i, c := range GraphColors {
		GraphCharStyles[i] = lipgloss.NewStyle().Foreground(c)
	}
}

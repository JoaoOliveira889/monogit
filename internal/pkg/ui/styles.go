package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	ColorBg        lipgloss.Color
	ColorFg        lipgloss.Color
	ColorHighlight lipgloss.Color
	ColorSelected  lipgloss.Color
	ColorAccent    lipgloss.Color
	ColorSuccess   lipgloss.Color
	ColorError     lipgloss.Color
	ColorWarning   lipgloss.Color
	ColorSubtle    lipgloss.Color
	ColorBorder    lipgloss.Color
	ColorCyan      lipgloss.Color
	ColorOrange    lipgloss.Color
	ColorIndigo    lipgloss.Color
	ColorRose      lipgloss.Color
	ColorEmerald   lipgloss.Color
	ColorAmber     lipgloss.Color
	ColorMono      lipgloss.Color
	ColorGit       lipgloss.Color
)

const (
	IconAhead  = "↑"
	IconBehind = "↓"
	IconDirty  = "✗"
	IconClean  = "✓"
	IconSpace  = " "
)

var (
	LeftPanelStyle    lipgloss.Style
	ActivePanelStyle  lipgloss.Style
	RightPanelStyle   lipgloss.Style
	HeaderStyle       lipgloss.Style
	BrandMonoStyle    lipgloss.Style
	BrandGitStyle     lipgloss.Style
	BrandTitleStyle   lipgloss.Style
	FooterStyle       lipgloss.Style
	FooterKeyStyle    lipgloss.Style
	FooterActionStyle lipgloss.Style
	PanelTitleStyle   lipgloss.Style
	LabelStyle        lipgloss.Style
)

var (
	SelectedItemStyle lipgloss.Style
	NormalItemStyle   lipgloss.Style
	PointerStyle      lipgloss.Style
)

var (
	BranchStyle  lipgloss.Style
	AheadStyle   lipgloss.Style
	BehindStyle  lipgloss.Style
	DirtyStyle   lipgloss.Style
	CleanStyle   lipgloss.Style
	ErrorStyle   lipgloss.Style
	SuccessStyle lipgloss.Style
	WarningStyle lipgloss.Style
	SubtleStyle  lipgloss.Style
	SpinnerStyle lipgloss.Style
	ValueStyle   lipgloss.Style
	InputStyle   lipgloss.Style
)

var (
	GraphColors     []lipgloss.Color
	GraphCharStyles []lipgloss.Style
)

var (
	DiffAddStyle  lipgloss.Style
	DiffDelStyle  lipgloss.Style
	DiffHunkStyle lipgloss.Style
)

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
	ApplyTheme("Tokyo Night")
}

type Theme struct {
	Name      string
	Bg        string
	Fg        string
	Highlight string
	Selected  string
	Accent    string
	Success   string
	Error     string
	Warning   string
	Subtle    string
	Border    string
	Cyan      string
	Orange    string
	Indigo    string
	Rose      string
	Emerald   string
	Amber     string
	Mono      string
	Git       string
}

var Themes = []Theme{
	{
		Name:      "Tokyo Night",
		Bg:        "#1a1b26",
		Fg:        "#a9b1d6",
		Highlight: "#7aa2f7",
		Selected:  "#364a82",
		Accent:    "#bb9af7",
		Success:   "#9ece6a",
		Error:     "#f7768e",
		Warning:   "#e0af68",
		Subtle:    "#565f89",
		Border:    "#414868",
		Cyan:      "#7dcfff",
		Orange:    "#ff9e64",
		Indigo:    "#6366f1",
		Rose:      "#fb7185",
		Emerald:   "#34d399",
		Amber:     "#fbbf24",
		Mono:      "#d7d7d7",
		Git:       "#b8ff3d",
	},
	{
		Name:      "Dracula",
		Bg:        "#282a36",
		Fg:        "#f8f8f2",
		Highlight: "#bd93f9",
		Selected:  "#44475a",
		Accent:    "#ff79c6",
		Success:   "#50fa7b",
		Error:     "#ff5555",
		Warning:   "#f1fa8c",
		Subtle:    "#6272a4",
		Border:    "#44475a",
		Cyan:      "#8be9fd",
		Orange:    "#ffb86c",
		Indigo:    "#6272a4",
		Rose:      "#ff79c6",
		Emerald:   "#50fa7b",
		Amber:     "#f1fa8c",
		Mono:      "#f8f8f2",
		Git:       "#50fa7b",
	},
	{
		Name:      "Nord",
		Bg:        "#2e3440",
		Fg:        "#d8dee9",
		Highlight: "#88c0d0",
		Selected:  "#434c5e",
		Accent:    "#b48ead",
		Success:   "#a3be8c",
		Error:     "#bf616a",
		Warning:   "#ebcb8b",
		Subtle:    "#4c566a",
		Border:    "#434c5e",
		Cyan:      "#8fbcbb",
		Orange:    "#d08770",
		Indigo:    "#5e81ac",
		Rose:      "#b48ead",
		Emerald:   "#a3be8c",
		Amber:     "#ebcb8b",
		Mono:      "#e5e9f0",
		Git:       "#8fbcbb",
	},
	{
		Name:      "Gruvbox",
		Bg:        "#282828",
		Fg:        "#ebdbb2",
		Highlight: "#fe8019",
		Selected:  "#504945",
		Accent:    "#b8bb26",
		Success:   "#b8bb26",
		Error:     "#fb4934",
		Warning:   "#fabd2f",
		Subtle:    "#928374",
		Border:    "#504945",
		Cyan:      "#8ec07c",
		Orange:    "#fe8019",
		Indigo:    "#83a598",
		Rose:      "#d3869b",
		Emerald:   "#b8bb26",
		Amber:     "#fabd2f",
		Mono:      "#ebdbb2",
		Git:       "#b8bb26",
	},
	{
		Name:      "Monokai",
		Bg:        "#272822",
		Fg:        "#f8f8f2",
		Highlight: "#a6e22e",
		Selected:  "#49483e",
		Accent:    "#f92672",
		Success:   "#a6e22e",
		Error:     "#f92672",
		Warning:   "#e6db74",
		Subtle:    "#75715e",
		Border:    "#49483e",
		Cyan:      "#66d9ef",
		Orange:    "#fd971f",
		Indigo:    "#ae81ff",
		Rose:      "#f92672",
		Emerald:   "#a6e22e",
		Amber:     "#e6db74",
		Mono:      "#f8f8f2",
		Git:       "#a6e22e",
	},
	{
		Name:      "One Dark",
		Bg:        "#282c34",
		Fg:        "#abb2bf",
		Highlight: "#61afef",
		Selected:  "#3e4452",
		Accent:    "#c678dd",
		Success:   "#98c379",
		Error:     "#e06c75",
		Warning:   "#d19a66",
		Subtle:    "#5c6370",
		Border:    "#3e4452",
		Cyan:      "#56b6c2",
		Orange:    "#d19a66",
		Indigo:    "#61afef",
		Rose:      "#c678dd",
		Emerald:   "#98c379",
		Amber:     "#e2c08d",
		Mono:      "#abb2bf",
		Git:       "#98c379",
	},
}

func ApplyTheme(name string) {
	var selected Theme
	found := false
	for _, t := range Themes {
		if strings.EqualFold(t.Name, name) {
			selected = t
			found = true
			break
		}
	}
	if !found {
		selected = Themes[0]
	}

	ColorBg = lipgloss.Color(selected.Bg)
	ColorFg = lipgloss.Color(selected.Fg)
	ColorHighlight = lipgloss.Color(selected.Highlight)
	ColorSelected = lipgloss.Color(selected.Selected)
	ColorAccent = lipgloss.Color(selected.Accent)
	ColorSuccess = lipgloss.Color(selected.Success)
	ColorError = lipgloss.Color(selected.Error)
	ColorWarning = lipgloss.Color(selected.Warning)
	ColorSubtle = lipgloss.Color(selected.Subtle)
	ColorBorder = lipgloss.Color(selected.Border)
	ColorCyan = lipgloss.Color(selected.Cyan)
	ColorOrange = lipgloss.Color(selected.Orange)
	ColorIndigo = lipgloss.Color(selected.Indigo)
	ColorRose = lipgloss.Color(selected.Rose)
	ColorEmerald = lipgloss.Color(selected.Emerald)
	ColorAmber = lipgloss.Color(selected.Amber)
	ColorMono = lipgloss.Color(selected.Mono)
	ColorGit = lipgloss.Color(selected.Git)

	GraphColors = []lipgloss.Color{
		ColorHighlight,
		ColorAccent,
		ColorCyan,
		ColorSuccess,
		ColorWarning,
		ColorError,
		ColorIndigo,
		ColorMono,
	}

	GraphCharStyles = make([]lipgloss.Style, len(GraphColors))
	for i, c := range GraphColors {
		GraphCharStyles[i] = lipgloss.NewStyle().Foreground(c)
	}

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

	SelectedItemStyle = lipgloss.NewStyle().
		Background(ColorHighlight).
		Foreground(ColorBg).
		Bold(true)

	NormalItemStyle = lipgloss.NewStyle().
		Foreground(ColorFg)

	PointerStyle = lipgloss.NewStyle().
		Foreground(ColorAccent).
		Bold(true)

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

	DiffAddStyle = lipgloss.NewStyle().Foreground(ColorSuccess)
	DiffDelStyle = lipgloss.NewStyle().Foreground(ColorError)
	DiffHunkStyle = lipgloss.NewStyle().Foreground(ColorHighlight)
}

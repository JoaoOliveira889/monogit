package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

type keyMap struct {
	Up          []string
	Down        []string
	Left        []string
	Right       []string
	Enter       []string
	Esc         []string
	Help        []string
	Quit        []string
	Fetch       []string
	FetchAll    []string
	Pull        []string
	PullAll     []string
	Push        []string
	PushAll     []string
	Commit      []string
	Log         []string
	Files       []string
	Branches    []string
	Graph       []string
	Tab         []string
	Space       []string
	Stash       []string
	StashPop    []string
	Discard     []string
	AddAll      []string
	Undo        []string
	SelectAll   []string
	DeselectAll []string
	HelpAlt     []string
	CommandLog  []string
	Panel1      []string
	Panel2      []string
	Panel3      []string
	CreateBranch []string
	DeleteBranch []string
	OpenEditor   []string
	OpenBrowser  []string
}

var keys = keyMap{
	Up:          []string{"up", "k"},
	Down:        []string{"down", "j"},
	Left:        []string{"left", "h"},
	Right:       []string{"right", "l"},
	Enter:       []string{"enter"},
	Esc:         []string{"esc"},
	Help:        []string{"?"},
	Quit:        []string{"q", "ctrl+c"},
	Fetch:       []string{"f"},
	FetchAll:    []string{"F"},
	Pull:        []string{"p"},
	PullAll:     []string{"P"},
	Push:        []string{"u"},
	PushAll:     []string{"U"},
	Commit:      []string{"c"},
	Log:         []string{"L"},
	Files:       []string{"v"},
	Branches:    []string{"b"},
	Graph:       []string{"g", "G"},
	Tab:         []string{"tab"},
	Space:       []string{" "},
	Stash:       []string{"s"},
	StashPop:    []string{"S"},
	Discard:     []string{"x"},
	AddAll:      []string{"A"},
	Undo:        []string{"z"},
	SelectAll:   []string{"a"},
	DeselectAll: []string{"n"},
	HelpAlt:     []string{"ctrl+p"},
	CommandLog:  []string{"o"},
	Panel1:      []string{"1"},
	Panel2:      []string{"2"},
	Panel3:      []string{"3"},
	CreateBranch: []string{"n"},
	DeleteBranch: []string{"d"},
	OpenEditor:   []string{"e"},
	OpenBrowser:  []string{"w"},
}

func matchesKey(msg tea.KeyMsg, keys ...string) bool {
	s := msg.String()
	for _, k := range keys {
		if s == k {
			return true
		}
	}
	return false
}

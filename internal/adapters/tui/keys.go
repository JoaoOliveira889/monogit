package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

type keyMap struct {
	Up           []string
	Down         []string
	Left         []string
	Right        []string
	Enter        []string
	Esc          []string
	Help         []string
	Quit         []string
	Fetch        []string
	FetchAll     []string
	Pull         []string
	PullAll      []string
	Push         []string
	PushAll      []string
	Commit       []string
	Log          []string
	Files        []string
	Branches     []string
	Graph        []string
	Tab          []string
	Space        []string
	Stash        []string
	StashPop     []string
	StashList    []string
	StashApply   []string
	StashDrop    []string
	Discard      []string
	AddAll       []string
	Undo         []string
	SelectAll    []string
	DeselectAll  []string
	Copy         []string
	Paste        []string
	HelpAlt      []string
	CommandLog   []string
	Panel1       []string
	Panel2       []string
	Panel3       []string
	CreateBranch []string
	DeleteBranch []string
	OpenEditor   []string
	OpenBrowser  []string
	Tag          []string
	ResizeLeft   []string
	ResizeRight  []string
}

var keys = keyMap{
	Up:           []string{"up", "k"},
	Down:         []string{"down", "j"},
	Left:         []string{"left", "h"},
	Right:        []string{"right", "l"},
	Enter:        []string{"enter"},
	Esc:          []string{"esc"},
	Help:         []string{"?"},
	Quit:         []string{"q", "ctrl+c"},
	Fetch:        []string{"f"},
	FetchAll:     []string{"F"},
	Pull:         []string{"p"},
	PullAll:      []string{"P"},
	Push:         []string{"u"},
	PushAll:      []string{"U"},
	Commit:       []string{"c"},
	Log:          []string{"L"},
	Files:        []string{"v"},
	Branches:     []string{"b"},
	Graph:        []string{"g", "G"},
	Tab:          []string{"tab"},
	Space:        []string{" "},
	Stash:        []string{"s"},
	StashPop:     []string{"p"},
	StashList:    []string{"S"},
	StashApply:   []string{"a"},
	StashDrop:    []string{"d"},
	Discard:      []string{"x"},
	AddAll:       []string{"A"},
	Undo:         []string{"z"},
	SelectAll:    []string{"a"},
	DeselectAll:  []string{"n"},
	Copy:         []string{"y"},
	Paste:        []string{"ctrl+v"},
	HelpAlt:      []string{"ctrl+p"},
	CommandLog:   []string{"o"},
	Panel1:       []string{"1"},
	Panel2:       []string{"2"},
	Panel3:       []string{"3"},
	CreateBranch: []string{"n"},
	DeleteBranch: []string{"d"},
	OpenEditor:   []string{"e"},
	OpenBrowser:  []string{"w"},
	Tag:          []string{"t"},
	ResizeLeft:   []string{"<"},
	ResizeRight:  []string{">"},
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

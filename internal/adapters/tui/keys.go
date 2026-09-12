package tui

import (
	tea "charm.land/bubbletea/v2"
)

type keyMap struct {
	Up               []string
	Down             []string
	Left             []string
	Right            []string
	Enter            []string
	Esc              []string
	Help             []string
	Quit             []string
	Fetch            []string
	FetchAll         []string
	Pull             []string
	PullAll          []string
	Push             []string
	PushAll          []string
	Commit           []string
	Files            []string
	Branches         []string
	Graph            []string
	JumpTop          []string
	NextDirty        []string
	PrevDirty        []string
	RepeatAction     []string
	JumpBack         []string
	JumpForward      []string
	Palette          []string
	Tab              []string
	Space            []string
	Stash            []string
	StashPop         []string
	StashList        []string
	StashApply       []string
	StashDrop        []string
	Discard          []string
	Undo             []string
	SelectAll        []string
	DeselectAll      []string
	Copy             []string
	Paste            []string
	HelpAlt          []string
	CommandLog       []string
	WindowPrefix     []string
	Panel1           []string
	Panel2           []string
	Panel3           []string
	CreateBranch     []string
	DeleteBranch     []string
	OpenEditor       []string
	OpenBrowser      []string
	Tag              []string
	TagFilter        []string
	TagAssign        []string
	Search           []string
	StatusFilter     []string
	ResizeLeft       []string
	ResizeRight      []string
	Merge            []string
	ResolveConflicts []string
	Diff             []string
	CompactDiff      []string
	BulkCheckout     []string
	BulkStash        []string
	CherryPick       []string
	Revert           []string
	Config           []string
	ExportLog        []string
	Rebase           []string
	HalfPageDown     []string
	HalfPageUp       []string
	Top              []string
	Bottom           []string
}

var keys = keyMap{
	Up:               []string{"up", "k"},
	Down:             []string{"down", "j"},
	Left:             []string{"left", "h"},
	Right:            []string{"right", "l"},
	Enter:            []string{"enter"},
	Esc:              []string{"esc"},
	Help:             []string{"?"},
	Quit:             []string{"q", "ctrl+c"},
	Fetch:            []string{"f"},
	FetchAll:         []string{"F"},
	Pull:             []string{"p"},
	PullAll:          []string{"P"},
	Push:             []string{"u"},
	PushAll:          []string{"U"},
	Commit:           []string{"c"},
	Files:            []string{"v"},
	Branches:         []string{"b"},
	Graph:            []string{"gl"},
	JumpTop:          []string{"gg"},
	NextDirty:        []string{"}"},
	PrevDirty:        []string{"{"},
	RepeatAction:     []string{"."},
	JumpBack:         []string{"ctrl+o"},
	JumpForward:      []string{"ctrl+i"},
	Palette:          []string{":"},
	Tab:              []string{"tab"},
	Space:            []string{"space", " "},
	Stash:            []string{"s"},
	StashPop:         []string{"p"},
	StashList:        []string{"S"},
	StashApply:       []string{"a"},
	StashDrop:        []string{"d"},
	Discard:          []string{"x"},
	Undo:             []string{"z"},
	SelectAll:        []string{"a"},
	DeselectAll:      []string{"n"},
	Copy:             []string{"y"},
	Paste:            []string{"ctrl+v"},
	HelpAlt:          []string{"ctrl+p"},
	CommandLog:       []string{"o"},
	WindowPrefix:     []string{"ctrl+w"},
	Panel1:           []string{"ctrl+w1"},
	Panel2:           []string{"ctrl+w2"},
	Panel3:           []string{"ctrl+w3"},
	CreateBranch:     []string{"n"},
	DeleteBranch:     []string{"d"},
	OpenEditor:       []string{"e"},
	OpenBrowser:      []string{"w"},
	Tag:              []string{"t"},
	TagFilter:        []string{"ctrl+g"},
	TagAssign:        []string{"ctrl+t"},
	Search:           []string{"/"},
	StatusFilter:     []string{"ctrl+f"},
	ResizeLeft:       []string{"<"},
	ResizeRight:      []string{">"},
	Merge:            []string{"M"},
	ResolveConflicts: []string{"m"},
	Diff:             []string{"d"},
	CompactDiff:      []string{"C"},
	BulkCheckout:     []string{"B"},
	BulkStash:        []string{"Z"},
	CherryPick:       []string{"ctrl+y"},
	Revert:           []string{"ctrl+r"},
	Config:           []string{","},
	ExportLog:        []string{"E"},
	Rebase:           []string{"R"},
	HalfPageDown:     []string{"ctrl+d", "pgdown"},
	HalfPageUp:       []string{"ctrl+u", "pgup"},
	Top:              []string{"home", "gg"},
	Bottom:           []string{"G", "end"},
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

// matchesSeq reports whether a resolved key sequence is bound to an action.
func matchesSeq(seq string, keys ...string) bool {
	for _, k := range keys {
		if seq == k {
			return true
		}
	}
	return false
}

package tui

import (
	"time"

	"monogit/internal/domain"
)

type errMsg struct{ Err error }

func (e errMsg) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return ""
}

type repoScannedMsg struct{ repos []domain.Repository }

type repoStatusMsg struct {
	index   int
	branch  string
	ahead   int
	behind  int
	dirty   bool
	refresh bool
	err     error
}

type fetchDoneMsg struct {
	index  int
	all    bool
	err    error
}

type pullDoneMsg struct {
	index  int
	output string
	err    error
}

type PullResult struct {
	Index  int
	Name   string
	Output string
	Err    error
}

type pullAllDoneMsg struct {
	results []PullResult
}

type pushDoneMsg struct {
	index  int
	output string
	err    error
}

type PushResult struct {
	Index  int
	Name   string
	Output string
	Err    error
}

type pushAllDoneMsg struct {
	results []PushResult
}

type commitDoneMsg struct {
	index  int
	output string
	err    error
}

type gitFilesMsg struct {
	files []domain.FileStatus
}

type gitDiffMsg struct {
	diff string
}

type gitBranchesMsg struct {
	branches []domain.BranchInfo
}

type stashDoneMsg struct {
	index  int
	output string
	err    error
}

type stashPopDoneMsg struct {
	index  int
	output string
	err    error
}

type tickMsg time.Time

type refreshMsg struct{}

type nextStepMsg struct{}

type spinnerTickMsg struct{}

type deleteBranchDoneMsg struct {
	index  int
	output string
	err    error
}

type deleteRemoteBranchDoneMsg struct {
	index  int
	output string
	err    error
}

type clearStatusMsg struct {
	id int
}

type checkoutBranchDoneMsg struct {
	index int
	err   error
}

type openBrowserMsg struct {
	url string
	err error
}

type openEditorMsg struct {
	editor string
	err    error
}

type editorsDetectedMsg struct {
	editors []string
}
type tagDoneMsg struct {
	index  int
	output string
	err    error
}

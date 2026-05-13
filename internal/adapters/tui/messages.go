package tui

import (
	"time"

	"monogit/internal/domain"
)

type errMsg struct{ err error }

type repoScannedMsg struct{ repos []domain.Repository }

type repoStatusMsg struct {
	index  int
	branch string
	ahead  int
	behind int
	dirty   bool
	refresh bool
	err     error
}

type fetchDoneMsg struct {
	index int
	err   error
}

type fetchAllDoneMsg struct{}

type pullDoneMsg struct {
	index  int
	output string
	err    error
}

type pullResult struct {
	index  int
	name   string
	output string
	err    error
}

type pullAllDoneMsg struct {
	results []pullResult
}

type pushDoneMsg struct {
	index  int
	output string
	err    error
}

type pushResult struct {
	index  int
	name   string
	output string
	err    error
}

type pushAllDoneMsg struct {
	results []pushResult
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
type clearStatusMsg struct{}

type checkoutBranchDoneMsg struct {
	index int
	err   error
}

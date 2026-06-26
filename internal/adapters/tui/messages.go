package tui

import (
	"time"

	"github.com/JoaoOliveira889/monogit/internal/domain"
)

type errMsg struct{ Err error }

func (e errMsg) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return ""
}

type repoScannedMsg struct{ repos []domain.Repository }
type startupReposMsg struct{ repos []domain.Repository }

type repoStatusMsg struct {
	index       int
	branch      string
	ahead       int
	behind      int
	dirty       bool
	detached    bool
	hasUpstream bool
	refresh     bool
	err         error
}

type repoDetailMsg struct {
	index          int
	path           string
	branch         string
	ahead          int
	behind         int
	dirty          bool
	detached       bool
	hasUpstream    bool
	hasConflicts   bool
	isStale        bool
	hasUnpushedTag bool
	modified       int
	untracked      int
	lastCommit     string
	log            string
	graph          bool
	needsLog       bool
	err            error
}

type fetchDoneMsg struct {
	index  int
	output string
	err    error
}

type fetchAllResult struct {
	Index  int
	Name   string
	Output string
	Err    error
}

type fetchAllDoneMsg struct {
	results []fetchAllResult
	err     error
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

type gitStashesMsg struct {
	stashes []domain.StashInfo
}

type stashApplyDoneMsg struct {
	index  int
	output string
	err    error
}

type stashDropDoneMsg struct {
	index  int
	output string
	err    error
}

type stashPopIndexDoneMsg struct {
	index  int
	output string
	err    error
}

type stashFilesMsg struct {
	files []string
}

type tickMsg time.Time

type refreshMsg struct{}

type nextStepMsg struct{}

type spinnerTickMsg struct{}
type splashTickMsg struct{}

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

type mergeDoneMsg struct {
	index  int
	branch string
	output string
	err    error
}

type conflictFilesMsg struct {
	files []domain.ConflictFile
}

type compactDiffMsg struct {
	changes []domain.CompactChange
}

type mergetoolDoneMsg struct {
	index  int
	path   string
	file   string
	output string
	err    error
}

type BulkCheckoutResult struct {
	Index  int
	Name   string
	Branch string
	Err    error
}

type checkoutAllDoneMsg struct {
	results []BulkCheckoutResult
}

type BulkStashResult struct {
	Index  int
	Name   string
	Output string
	Err    error
}

type stashAllDoneMsg struct {
	results []BulkStashResult
}

type configSavedMsg struct {
	err error
}

type cherryPickDoneMsg struct {
	index  int
	hash   string
	output string
	err    error
}

type revertDoneMsg struct {
	index  int
	hash   string
	output string
	err    error
}

type repoUnpushedTagMsg struct {
	index          int
	hasUnpushedTag bool
	err            error
}

type exportLogMsg struct {
	path string
	err  error
}


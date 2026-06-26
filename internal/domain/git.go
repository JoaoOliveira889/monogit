package domain

type FileStatus struct {
	Name      string
	Staged    bool
	Modified  bool
	Untracked bool
}

type BranchInfo struct {
	Name      string
	IsRemote  bool
	IsLocal   bool
	IsCurrent bool
}

type Repository struct {
	Name           string
	Path           string
	Branch         string
	Tags           []string
	Ahead          int
	Behind         int
	IsDirty        bool
	IsDetached     bool
	HasUpstream    bool
	HasConflicts   bool
	IsStale        bool
	HasUnpushedTag bool
	Fetching       bool
	Pulling        bool
	Pushing        bool
	Stashing       bool
	Committing     bool
	CheckingOut    bool
	Merging        bool
	Tagging        bool
	LastOutput     string
	Error          string
}

type RepositorySnapshot struct {
	Branch         string
	Ahead          int
	Behind         int
	IsDirty        bool
	IsDetached     bool
	HasUpstream    bool
	HasConflicts   bool
	IsStale        bool
	HasUnpushedTag bool
	ModifiedCount  int
	UntrackedCount int
	LastCommit     string
	LastCommitUnix int64
	Log            string
	LogGraph       bool
}

type BranchManager interface {
	GetBranch(repoPath string) (string, error)
	GetBranches(repoPath string) ([]BranchInfo, error)
	GetAheadBehind(repoPath string) (ahead int, behind int, err error)
	CheckoutBranch(repoPath string, name string) error
	CreateBranch(repoPath string, name string) error
	DeleteBranch(repoPath string, name string) (string, error)
	DeleteRemoteBranch(repoPath string, remote string, name string) (string, error)
}

type StatusReporter interface {
	IsDirty(repoPath string) (bool, error)
	GetStatusFiles(repoPath string) ([]FileStatus, error)
	GetDiff(repoPath string, f FileStatus) (string, error)
}

type LogProvider interface {
	GetGraphLog(repoPath string, n int) (string, error)
	GetSimpleLog(repoPath string, n int) (string, error)
	GetRepositorySnapshot(repoPath string, viewGraph bool, logLines int) (RepositorySnapshot, error)
	GetQuickSnapshot(repoPath string) (RepositorySnapshot, error)
}

type HealthChecker interface {
	HasUpstream(repoPath string) (bool, error)
}

type RemoteOperator interface {
	FetchAll(repoPath string) error
	Pull(repoPath string) (string, error)
	Push(repoPath string) (string, error)
	GetRemoteURL(repoPath string) (string, error)
}

type CommitManager interface {
	AddAndCommit(repoPath string, message string) (string, error)
	Commit(repoPath string, message string) (string, error)
	StageByPattern(repoPath string, pattern string) error
	StageFiles(repoPath string, files []string) error
	UnstageAll(repoPath string) error
	UnstageFile(repoPath string, fileName string) error
	UndoCommit(repoPath string) error
	CherryPick(repoPath string, hash string) (string, error)
	Revert(repoPath string, hash string) (string, error)
}

type StashInfo struct {
	Index   int
	Message string
}

type StashManager interface {
	Stash(repoPath string, message string) (string, error)
	StashPop(repoPath string) (string, error)
	GetStashes(repoPath string) ([]StashInfo, error)
	ApplyStash(repoPath string, index int) (string, error)
	DropStash(repoPath string, index int) (string, error)
	PopStash(repoPath string, index int) (string, error)
	GetStashFiles(repoPath string, index int) ([]string, error)
	GetStashFileDiff(repoPath string, index int, file string) (string, error)
}

type FileDiscarder interface {
	DiscardChanges(repoPath string, f FileStatus) error
}

type TagManager interface {
	CreateTag(repoPath string, name string, message string) (string, error)
	PushTag(repoPath string, name string) (string, error)
}

type ConflictFile struct {
	Name   string
	Status string
}

type CompactChange struct {
	FileName     string
	FunctionName string
	LineRange    string
}

type CommandSpec struct {
	Name string
	Args []string
	Dir  string
}

type MergeOperator interface {
	Merge(repoPath string, branch string) (string, error)
}

type ConflictResolver interface {
	HasConflicts(repoPath string) (bool, error)
	ListConflictingFiles(repoPath string) ([]ConflictFile, error)
	GetCompactDiff(repoPath string, f FileStatus) ([]CompactChange, error)
	OpenMergetool(repoPath string, tool string, file string) (CommandSpec, error)
}

type GitProvider interface {
	BranchManager
	StatusReporter
	LogProvider
	RemoteOperator
	CommitManager
	StashManager
	FileDiscarder
	TagManager
	MergeOperator
	ConflictResolver
	HealthChecker
	HasUnpushedHeadTag(repoPath string) (bool, error)
}

type RepositoryOperator interface {
	GetRepositoryStatus(path string) (Repository, error)
	GetQuickSnapshot(path string) (RepositorySnapshot, error)
	GetRepositorySnapshot(path string, viewGraph bool, logLines int) (RepositorySnapshot, error)
	Fetch(path string) error
	Pull(path string) (string, error)
	Push(path string) (string, error)
	Merge(path string, branch string) (string, error)
	GetRemoteURL(path string) (string, error)
	Commit(path string, message string) (string, error)
	CommitAll(path string, message string) (string, error)
	CommitSelected(path string, files []string, message string) (string, error)
	GetBranches(path string) ([]BranchInfo, error)
	Stash(path string, message string) (string, error)
	StashPop(path string) (string, error)
	GetStashes(path string) ([]StashInfo, error)
	ApplyStash(path string, index int) (string, error)
	DropStash(path string, index int) (string, error)
	PopStash(path string, index int) (string, error)
	GetStashFiles(path string, index int) ([]string, error)
	GetStashFileDiff(path string, index int, file string) (string, error)
	UnstageAll(path string) error
	UndoCommit(path string) error
	StageByPattern(path string, pattern string) error
	AddAll(path string) error
	ToggleFile(path string, file FileStatus) error
	GetFiles(path string) ([]FileStatus, error)
	GetDiff(path string, file FileStatus) (string, error)
	DiscardFile(path string, file FileStatus) error
	GetSimpleLog(path string, n int) (string, error)
	GetGraphLog(path string, n int) (string, error)
	CheckoutBranch(path string, branch string) error
	CreateBranch(path string, branch string) error
	DeleteBranch(path string, branch string) (string, error)
	DeleteRemoteBranch(path string, remote string, branch string) (string, error)
	CreateAndPushTag(path, name, message string) (string, error)
	HasConflicts(path string) (bool, error)
	ListConflictingFiles(path string) ([]ConflictFile, error)
	GetCompactDiff(path string, file FileStatus) ([]CompactChange, error)
	OpenMergetool(path string, tool string, file string) (CommandSpec, error)
	CherryPick(path string, hash string) (string, error)
	Revert(path string, hash string) (string, error)
	HasUnpushedHeadTag(path string) (bool, error)
}

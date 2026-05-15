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
	Name        string
	Path        string
	Branch      string
	Ahead       int
	Behind      int
	IsDirty     bool
	Fetching    bool
	Pulling     bool
	Pushing     bool
	Stashing    bool
	Committing  bool
	CheckingOut bool
	Tagging     bool
	LastOutput  string
	Error       string
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
}

type RemoteOperator interface {
	FetchAll(repoPath string) error
	Pull(repoPath string) (string, error)
	Push(repoPath string) (string, error)
	GetRemoteURL(repoPath string) (string, error)
}

type CommitManager interface {
	AddAndCommit(repoPath string, message string) (string, error)
	StageByPattern(repoPath string, pattern string) error
	UnstageAll(repoPath string) error
	UnstageFile(repoPath string, fileName string) error
	UndoCommit(repoPath string) error
}

type StashManager interface {
	Stash(repoPath string, message string) (string, error)
	StashPop(repoPath string) (string, error)
}

type FileDiscarder interface {
	DiscardChanges(repoPath string, f FileStatus) error
}

type TagManager interface {
	CreateTag(repoPath string, name string, message string) (string, error)
	PushTag(repoPath string, name string) (string, error)
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
}

type RepositoryScanner interface {
	Scan(rootPath string) ([]Repository, error)
}

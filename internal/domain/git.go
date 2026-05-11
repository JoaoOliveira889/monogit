package domain

type FileStatus struct {
	Name      string
	Staged    bool
	Modified  bool
	Untracked bool
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
	LastOutput  string
	Error       string
}

type GitProvider interface {
	GetBranch(repoPath string) (string, error)
	IsDirty(repoPath string) (bool, error)
	GetAheadBehind(repoPath string) (ahead int, behind int, err error)
	FetchAll(repoPath string) error
	Pull(repoPath string) (string, error)
	AddAndCommit(repoPath string, message string) (string, error)
	GetStatusFiles(repoPath string) ([]FileStatus, error)
	GetDiff(repoPath string, f FileStatus) (string, error)
	DiscardChanges(repoPath string, f FileStatus) error
	GetBranches(repoPath string) ([]string, error)
	Push(repoPath string) (string, error)
	CheckoutBranch(repoPath string, name string) error
	CreateBranch(repoPath string, name string) error
	Stash(repoPath string, message string) (string, error)
	StashPop(repoPath string) (string, error)
	UnstageAll(repoPath string) error
	UnstageFile(repoPath string, fileName string) error
	UndoCommit(repoPath string) error
	StageByPattern(repoPath string, pattern string) error
	GetGraphLog(repoPath string, n int) (string, error)
	GetSimpleLog(repoPath string, n int) (string, error)
}

type RepositoryScanner interface {
	Scan(rootPath string) ([]Repository, error)
}

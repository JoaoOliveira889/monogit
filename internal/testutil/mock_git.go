package testutil

import (
	"github.com/JoaoOliveira889/monogit/internal/domain"
)

type MockGitProvider struct {
	GetBranchFunc             func(string) (string, error)
	GetAheadBehindFunc        func(string) (int, int, error)
	IsDirtyFunc               func(string) (bool, error)
	FetchAllFunc              func(string) error
	PullFunc                  func(string) (string, error)
	PushFunc                  func(string) (string, error)
	GetRemoteURLFunc          func(string) (string, error)
	AddAndCommitFunc          func(string, string) (string, error)
	CommitFunc                func(string, string) (string, error)
	CherryPickFunc            func(string, string) (string, error)
	RevertFunc                func(string, string) (string, error)
	GetStatusFilesFunc        func(string) ([]domain.FileStatus, error)
	GetDiffFunc               func(string, domain.FileStatus) (string, error)
	DiscardChangesFunc        func(string, domain.FileStatus) error
	GetBranchesFunc           func(string) ([]domain.BranchInfo, error)
	CheckoutBranchFunc        func(string, string) error
	CreateBranchFunc          func(string, string) error
	StashFunc                 func(string, string) (string, error)
	StashPopFunc              func(string) (string, error)
	UnstageAllFunc            func(string) error
	UnstageFileFunc           func(string, string) error
	UndoCommitFunc            func(string) error
	StageByPatternFunc        func(string, string) error
	StageFilesFunc            func(string, []string) error
	GetGraphLogFunc           func(string, int) (string, error)
	GetSimpleLogFunc          func(string, int) (string, error)
	GetRepositorySnapshotFunc func(string, bool, int) (domain.RepositorySnapshot, error)
	GetQuickSnapshotFunc      func(string) (domain.RepositorySnapshot, error)
	CreateTagFunc             func(string, string, string) (string, error)
	PushTagFunc               func(string, string) (string, error)
	DeleteBranchFunc          func(string, string) (string, error)
	DeleteRemoteBranchFunc    func(string, string, string) (string, error)
	RemoveWorktreeForBranchFunc func(string, string, bool) (string, error)
	GetStashesFunc            func(string) ([]domain.StashInfo, error)
	ApplyStashFunc            func(string, int) (string, error)
	DropStashFunc             func(string, int) (string, error)
	PopStashFunc              func(string, int) (string, error)
	GetStashFilesFunc         func(string, int) ([]string, error)
	GetStashFileDiffFunc      func(string, int, string) (string, error)
	MergeFunc                 func(string, string) (string, error)
	HasConflictsFunc          func(string) (bool, error)
	ListConflictingFilesFunc  func(string) ([]domain.ConflictFile, error)
	GetCompactDiffFunc        func(string, domain.FileStatus) ([]domain.CompactChange, error)
	OpenMergetoolFunc         func(string, string, string) (domain.CommandSpec, error)
	HasUpstreamFunc           func(string) (bool, error)
	HasUnpushedHeadTagFunc    func(string) (bool, error)
}

func (m *MockGitProvider) GetBranch(repoPath string) (string, error) {
	if m.GetBranchFunc != nil {
		return m.GetBranchFunc(repoPath)
	}
	return "", nil
}
func (m *MockGitProvider) GetAheadBehind(repoPath string) (int, int, error) {
	if m.GetAheadBehindFunc != nil {
		return m.GetAheadBehindFunc(repoPath)
	}
	return 0, 0, nil
}
func (m *MockGitProvider) IsDirty(repoPath string) (bool, error) {
	if m.IsDirtyFunc != nil {
		return m.IsDirtyFunc(repoPath)
	}
	return false, nil
}
func (m *MockGitProvider) FetchAll(repoPath string) error {
	if m.FetchAllFunc != nil {
		return m.FetchAllFunc(repoPath)
	}
	return nil
}
func (m *MockGitProvider) Pull(repoPath string) (string, error) {
	if m.PullFunc != nil {
		return m.PullFunc(repoPath)
	}
	return "", nil
}
func (m *MockGitProvider) Push(repoPath string) (string, error) {
	if m.PushFunc != nil {
		return m.PushFunc(repoPath)
	}
	return "", nil
}
func (m *MockGitProvider) GetRemoteURL(repoPath string) (string, error) {
	if m.GetRemoteURLFunc != nil {
		return m.GetRemoteURLFunc(repoPath)
	}
	return "", nil
}
func (m *MockGitProvider) AddAndCommit(repoPath, message string) (string, error) {
	if m.AddAndCommitFunc != nil {
		return m.AddAndCommitFunc(repoPath, message)
	}
	return "", nil
}
func (m *MockGitProvider) Commit(repoPath, message string) (string, error) {
	if m.CommitFunc != nil {
		return m.CommitFunc(repoPath, message)
	}
	return "", nil
}
func (m *MockGitProvider) CherryPick(repoPath, hash string) (string, error) {
	if m.CherryPickFunc != nil {
		return m.CherryPickFunc(repoPath, hash)
	}
	return "", nil
}
func (m *MockGitProvider) Revert(repoPath, hash string) (string, error) {
	if m.RevertFunc != nil {
		return m.RevertFunc(repoPath, hash)
	}
	return "", nil
}
func (m *MockGitProvider) GetStatusFiles(repoPath string) ([]domain.FileStatus, error) {
	if m.GetStatusFilesFunc != nil {
		return m.GetStatusFilesFunc(repoPath)
	}
	return nil, nil
}
func (m *MockGitProvider) GetDiff(repoPath string, f domain.FileStatus) (string, error) {
	if m.GetDiffFunc != nil {
		return m.GetDiffFunc(repoPath, f)
	}
	return "", nil
}
func (m *MockGitProvider) DiscardChanges(repoPath string, f domain.FileStatus) error {
	if m.DiscardChangesFunc != nil {
		return m.DiscardChangesFunc(repoPath, f)
	}
	return nil
}
func (m *MockGitProvider) GetBranches(repoPath string) ([]domain.BranchInfo, error) {
	if m.GetBranchesFunc != nil {
		return m.GetBranchesFunc(repoPath)
	}
	return nil, nil
}
func (m *MockGitProvider) CheckoutBranch(repoPath, name string) error {
	if m.CheckoutBranchFunc != nil {
		return m.CheckoutBranchFunc(repoPath, name)
	}
	return nil
}
func (m *MockGitProvider) CreateBranch(repoPath, name string) error {
	if m.CreateBranchFunc != nil {
		return m.CreateBranchFunc(repoPath, name)
	}
	return nil
}
func (m *MockGitProvider) Stash(repoPath, message string) (string, error) {
	if m.StashFunc != nil {
		return m.StashFunc(repoPath, message)
	}
	return "", nil
}
func (m *MockGitProvider) StashPop(repoPath string) (string, error) {
	if m.StashPopFunc != nil {
		return m.StashPopFunc(repoPath)
	}
	return "", nil
}
func (m *MockGitProvider) UnstageAll(repoPath string) error {
	if m.UnstageAllFunc != nil {
		return m.UnstageAllFunc(repoPath)
	}
	return nil
}
func (m *MockGitProvider) UnstageFile(repoPath, fileName string) error {
	if m.UnstageFileFunc != nil {
		return m.UnstageFileFunc(repoPath, fileName)
	}
	return nil
}
func (m *MockGitProvider) UndoCommit(repoPath string) error {
	if m.UndoCommitFunc != nil {
		return m.UndoCommitFunc(repoPath)
	}
	return nil
}
func (m *MockGitProvider) StageByPattern(repoPath, pattern string) error {
	if m.StageByPatternFunc != nil {
		return m.StageByPatternFunc(repoPath, pattern)
	}
	return nil
}
func (m *MockGitProvider) StageFiles(repoPath string, files []string) error {
	if m.StageFilesFunc != nil {
		return m.StageFilesFunc(repoPath, files)
	}
	return nil
}
func (m *MockGitProvider) GetGraphLog(repoPath string, n int) (string, error) {
	if m.GetGraphLogFunc != nil {
		return m.GetGraphLogFunc(repoPath, n)
	}
	return "", nil
}
func (m *MockGitProvider) GetSimpleLog(repoPath string, n int) (string, error) {
	if m.GetSimpleLogFunc != nil {
		return m.GetSimpleLogFunc(repoPath, n)
	}
	return "", nil
}
func (m *MockGitProvider) GetRepositorySnapshot(repoPath string, viewGraph bool, n int) (domain.RepositorySnapshot, error) {
	if m.GetRepositorySnapshotFunc != nil {
		return m.GetRepositorySnapshotFunc(repoPath, viewGraph, n)
	}
	return domain.RepositorySnapshot{}, nil
}
func (m *MockGitProvider) GetQuickSnapshot(repoPath string) (domain.RepositorySnapshot, error) {
	if m.GetQuickSnapshotFunc != nil {
		return m.GetQuickSnapshotFunc(repoPath)
	}
	return domain.RepositorySnapshot{}, nil
}
func (m *MockGitProvider) CreateTag(repoPath, name, message string) (string, error) {
	if m.CreateTagFunc != nil {
		return m.CreateTagFunc(repoPath, name, message)
	}
	return "", nil
}
func (m *MockGitProvider) PushTag(repoPath, name string) (string, error) {
	if m.PushTagFunc != nil {
		return m.PushTagFunc(repoPath, name)
	}
	return "", nil
}
func (m *MockGitProvider) DeleteBranch(repoPath, name string) (string, error) {
	if m.DeleteBranchFunc != nil {
		return m.DeleteBranchFunc(repoPath, name)
	}
	return "", nil
}
func (m *MockGitProvider) DeleteRemoteBranch(repoPath, remote, name string) (string, error) {
	if m.DeleteRemoteBranchFunc != nil {
		return m.DeleteRemoteBranchFunc(repoPath, remote, name)
	}
	return "", nil
}
func (m *MockGitProvider) RemoveWorktreeForBranch(repoPath, branch string, force bool) (string, error) {
	if m.RemoveWorktreeForBranchFunc != nil {
		return m.RemoveWorktreeForBranchFunc(repoPath, branch, force)
	}
	return "", nil
}
func (m *MockGitProvider) GetStashes(repoPath string) ([]domain.StashInfo, error) {
	if m.GetStashesFunc != nil {
		return m.GetStashesFunc(repoPath)
	}
	return nil, nil
}
func (m *MockGitProvider) ApplyStash(repoPath string, index int) (string, error) {
	if m.ApplyStashFunc != nil {
		return m.ApplyStashFunc(repoPath, index)
	}
	return "", nil
}
func (m *MockGitProvider) DropStash(repoPath string, index int) (string, error) {
	if m.DropStashFunc != nil {
		return m.DropStashFunc(repoPath, index)
	}
	return "", nil
}
func (m *MockGitProvider) PopStash(repoPath string, index int) (string, error) {
	if m.PopStashFunc != nil {
		return m.PopStashFunc(repoPath, index)
	}
	return "", nil
}
func (m *MockGitProvider) GetStashFiles(repoPath string, index int) ([]string, error) {
	if m.GetStashFilesFunc != nil {
		return m.GetStashFilesFunc(repoPath, index)
	}
	return nil, nil
}
func (m *MockGitProvider) GetStashFileDiff(repoPath string, index int, file string) (string, error) {
	if m.GetStashFileDiffFunc != nil {
		return m.GetStashFileDiffFunc(repoPath, index, file)
	}
	return "", nil
}
func (m *MockGitProvider) Merge(repoPath, branch string) (string, error) {
	if m.MergeFunc != nil {
		return m.MergeFunc(repoPath, branch)
	}
	return "merged", nil
}
func (m *MockGitProvider) HasConflicts(repoPath string) (bool, error) {
	if m.HasConflictsFunc != nil {
		return m.HasConflictsFunc(repoPath)
	}
	return false, nil
}
func (m *MockGitProvider) ListConflictingFiles(repoPath string) ([]domain.ConflictFile, error) {
	if m.ListConflictingFilesFunc != nil {
		return m.ListConflictingFilesFunc(repoPath)
	}
	return nil, nil
}
func (m *MockGitProvider) GetCompactDiff(repoPath string, f domain.FileStatus) ([]domain.CompactChange, error) {
	if m.GetCompactDiffFunc != nil {
		return m.GetCompactDiffFunc(repoPath, f)
	}
	return nil, nil
}
func (m *MockGitProvider) OpenMergetool(repoPath, tool, file string) (domain.CommandSpec, error) {
	if m.OpenMergetoolFunc != nil {
		return m.OpenMergetoolFunc(repoPath, tool, file)
	}
	return domain.CommandSpec{Name: "git"}, nil
}
func (m *MockGitProvider) HasUpstream(repoPath string) (bool, error) {
	if m.HasUpstreamFunc != nil {
		return m.HasUpstreamFunc(repoPath)
	}
	return false, nil
}
func (m *MockGitProvider) HasUnpushedHeadTag(repoPath string) (bool, error) {
	if m.HasUnpushedHeadTagFunc != nil {
		return m.HasUnpushedHeadTagFunc(repoPath)
	}
	return false, nil
}

var _ domain.GitProvider = (*MockGitProvider)(nil)

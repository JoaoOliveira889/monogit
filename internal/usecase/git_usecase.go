package usecase

import (
	"github.com/JoaoOliveira889/monogit/internal/domain"
)

type GitUseCase struct {
	git domain.GitProvider
}

func NewGitUseCase(git domain.GitProvider) *GitUseCase {
	return &GitUseCase{git: git}
}

func (uc *GitUseCase) GetRepositoryStatus(path string) (domain.Repository, error) {
	branch, err := uc.git.GetBranch(path)
	if err != nil {
		return domain.Repository{}, err
	}

	ahead, behind, err := uc.git.GetAheadBehind(path)
	if err != nil {
		return domain.Repository{}, err
	}

	isDirty, err := uc.git.IsDirty(path)
	if err != nil {
		return domain.Repository{}, err
	}

	isDetached := branch == "HEAD"
	hasUpstream, _ := uc.git.HasUpstream(path)

	return domain.Repository{
		Path:        path,
		Branch:      branch,
		Ahead:       ahead,
		Behind:      behind,
		IsDirty:     isDirty,
		IsDetached:  isDetached,
		HasUpstream: hasUpstream,
	}, nil
}

func (uc *GitUseCase) GetQuickSnapshot(path string) (domain.RepositorySnapshot, error) {
	return uc.git.GetQuickSnapshot(path)
}

func (uc *GitUseCase) GetRepositorySnapshot(path string, viewGraph bool, logLines int) (domain.RepositorySnapshot, error) {
	return uc.git.GetRepositorySnapshot(path, viewGraph, logLines)
}

func (uc *GitUseCase) Fetch(path string) error {
	return uc.git.FetchAll(path)
}

func (uc *GitUseCase) Pull(path string) (string, error) {
	return uc.git.Pull(path)
}

func (uc *GitUseCase) Merge(path string, branch string) (string, error) {
	return uc.git.Merge(path, branch)
}

func (uc *GitUseCase) Push(path string) (string, error) {
	return uc.git.Push(path)
}

func (uc *GitUseCase) GetRemoteURL(path string) (string, error) {
	return uc.git.GetRemoteURL(path)
}

func (uc *GitUseCase) Commit(path string, message string) (string, error) {
	return uc.git.Commit(path, message)
}

func (uc *GitUseCase) CommitAll(path string, message string) (string, error) {
	return uc.git.AddAndCommit(path, message)
}

func (uc *GitUseCase) CommitSelected(path string, files []string, message string) (string, error) {
	if err := uc.git.UnstageAll(path); err != nil {
		return "", err
	}
	if err := uc.git.StageFiles(path, files); err != nil {
		return "", err
	}
	return uc.git.Commit(path, message)
}

func (uc *GitUseCase) GetBranches(path string) ([]domain.BranchInfo, error) {
	return uc.git.GetBranches(path)
}

func (uc *GitUseCase) Stash(path string, message string) (string, error) {
	return uc.git.Stash(path, message)
}

func (uc *GitUseCase) StashPop(path string) (string, error) {
	return uc.git.StashPop(path)
}

func (uc *GitUseCase) GetStashes(path string) ([]domain.StashInfo, error) {
	return uc.git.GetStashes(path)
}

func (uc *GitUseCase) ApplyStash(path string, index int) (string, error) {
	return uc.git.ApplyStash(path, index)
}

func (uc *GitUseCase) DropStash(path string, index int) (string, error) {
	return uc.git.DropStash(path, index)
}

func (uc *GitUseCase) PopStash(path string, index int) (string, error) {
	return uc.git.PopStash(path, index)
}

func (uc *GitUseCase) GetStashFiles(path string, index int) ([]string, error) {
	return uc.git.GetStashFiles(path, index)
}

func (uc *GitUseCase) UnstageAll(path string) error {
	return uc.git.UnstageAll(path)
}

func (uc *GitUseCase) UndoCommit(path string) error {
	return uc.git.UndoCommit(path)
}

func (uc *GitUseCase) StageByPattern(path string, pattern string) error {
	return uc.git.StageByPattern(path, pattern)
}

func (uc *GitUseCase) AddAll(path string) error {
	return uc.git.StageByPattern(path, ".")
}

func (uc *GitUseCase) ToggleFile(path string, file domain.FileStatus) error {
	if file.Staged {
		return uc.git.UnstageFile(path, file.Name)
	} else {
		return uc.git.StageByPattern(path, file.Name)
	}
}

func (uc *GitUseCase) GetFiles(path string) ([]domain.FileStatus, error) {
	return uc.git.GetStatusFiles(path)
}

func (uc *GitUseCase) GetDiff(path string, file domain.FileStatus) (string, error) {
	return uc.git.GetDiff(path, file)
}

func (uc *GitUseCase) DiscardFile(path string, file domain.FileStatus) error {
	return uc.git.DiscardChanges(path, file)
}

func (uc *GitUseCase) GetSimpleLog(path string, n int) (string, error) {
	return uc.git.GetSimpleLog(path, n)
}

func (uc *GitUseCase) GetGraphLog(path string, n int) (string, error) {
	return uc.git.GetGraphLog(path, n)
}

func (uc *GitUseCase) CheckoutBranch(path string, branch string) error {
	return uc.git.CheckoutBranch(path, branch)
}

func (uc *GitUseCase) CreateBranch(path string, branch string) error {
	return uc.git.CreateBranch(path, branch)
}

func (uc *GitUseCase) DeleteBranch(path string, branch string) (string, error) {
	return uc.git.DeleteBranch(path, branch)
}

func (uc *GitUseCase) DeleteRemoteBranch(path string, remote string, branch string) (string, error) {
	return uc.git.DeleteRemoteBranch(path, remote, branch)
}

func (uc *GitUseCase) CreateAndPushTag(path, name, message string) (string, error) {
	out1, err := uc.git.CreateTag(path, name, message)
	if err != nil {
		return out1, err
	}

	out2, err := uc.git.PushTag(path, name)
	if err != nil {
		return out1 + "\n" + out2, err
	}

	return out1 + "\n" + out2, nil
}

func (uc *GitUseCase) HasConflicts(path string) (bool, error) {
	return uc.git.HasConflicts(path)
}

func (uc *GitUseCase) ListConflictingFiles(path string) ([]domain.ConflictFile, error) {
	return uc.git.ListConflictingFiles(path)
}

func (uc *GitUseCase) GetCompactDiff(path string, file domain.FileStatus) ([]domain.CompactChange, error) {
	return uc.git.GetCompactDiff(path, file)
}

func (uc *GitUseCase) OpenMergetool(path string, tool string, file string) (domain.CommandSpec, error) {
	return uc.git.OpenMergetool(path, tool, file)
}

package usecase

import (
	"monogit/internal/domain"
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

	return domain.Repository{
		Path:    path,
		Branch:  branch,
		Ahead:   ahead,
		Behind:  behind,
		IsDirty: isDirty,
	}, nil
}

func (uc *GitUseCase) Fetch(path string) error {
	return uc.git.FetchAll(path)
}

func (uc *GitUseCase) Pull(path string) (string, error) {
	return uc.git.Pull(path)
}

func (uc *GitUseCase) Push(path string) (string, error) {
	return uc.git.Push(path)
}

func (uc *GitUseCase) Commit(path string, message string) (string, error) {
	return uc.git.AddAndCommit(path, message)
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

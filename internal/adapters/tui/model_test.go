package tui

import (
	"monogit/internal/domain"
	"monogit/internal/usecase"
	"testing"
	"time"

)

type mockGitProvider struct {
	domain.GitProvider
	getFilesFunc     func(string) ([]domain.FileStatus, error)
	getSimpleLogFunc func(string, int) (string, error)
}

func (m *mockGitProvider) GetStatusFiles(p string) ([]domain.FileStatus, error) {
	if m.getFilesFunc != nil {
		return m.getFilesFunc(p)
	}
	return nil, nil
}
func (m *mockGitProvider) GetSimpleLog(p string, n int) (string, error) {
	if m.getSimpleLogFunc != nil {
		return m.getSimpleLogFunc(p, n)
	}
	return "", nil
}

func TestNewModel(t *testing.T) {
	uc := usecase.NewGitUseCase(&mockGitProvider{})
	m := NewModel("/root", 30*time.Second, uc)
	if m.rootPath != "/root" {
		t.Errorf("expected /root, got %s", m.rootPath)
	}
}

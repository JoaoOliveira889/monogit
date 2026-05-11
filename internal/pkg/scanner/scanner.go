package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"monogit/internal/domain"
)

func ScanForRepos(rootPath string) ([]domain.Repository, error) {
	absRoot, err := filepath.Abs(rootPath)
	if err != nil {
		return nil, fmt.Errorf("resolve path: %w", err)
	}

	entries, err := os.ReadDir(absRoot)
	if err != nil {
		return nil, fmt.Errorf("read directory: %w", err)
	}

	var repos []domain.Repository
	for _, entry := range entries {
		if !entry.IsDir() || entry.Name()[0] == '.' {
			continue
		}

		dirPath := filepath.Join(absRoot, entry.Name())
		gitDir := filepath.Join(dirPath, ".git")

		info, err := os.Stat(gitDir)
		if err != nil || !info.IsDir() {
			continue
		}

		repos = append(repos, domain.Repository{
			Name: entry.Name(),
			Path: dirPath,
		})
	}

	slices.SortFunc(repos, func(a, b domain.Repository) int {
		if a.Name < b.Name {
			return -1
		}
		if a.Name > b.Name {
			return 1
		}
		return 0
	})

	return repos, nil
}

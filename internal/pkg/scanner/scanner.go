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

	var repos []domain.Repository
	err = filepath.WalkDir(absRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if !d.IsDir() {
			return nil
		}

		if d.Name() == "node_modules" || d.Name() == "vendor" || (d.Name() == ".git" && path != filepath.Join(absRoot, ".git")) {
			return filepath.SkipDir
		}

		gitPath := filepath.Join(path, ".git")
		if _, err := os.Stat(gitPath); err == nil {
		relPath, err := filepath.Rel(absRoot, path)
			if err != nil || relPath == "." {
				relPath = filepath.Base(path)
			}

			repos = append(repos, domain.Repository{
				Name: relPath,
				Path: path,
			})
			return filepath.SkipDir
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("scan directory: %w", err)
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

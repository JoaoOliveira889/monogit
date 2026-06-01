package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/JoaoOliveira889/monogit/internal/domain"
)

func ScanForRepos(rootPath string, repoTags map[string][]string, excludes []string) ([]domain.Repository, error) {
	absRoot, err := filepath.Abs(rootPath)
	if err != nil {
		return nil, fmt.Errorf("resolve path: %w", err)
	}

	excludeSet := make(map[string]struct{}, len(excludes))
	for _, name := range excludes {
		name = strings.TrimSpace(name)
		if name != "" {
			excludeSet[name] = struct{}{}
		}
	}

	var repos []domain.Repository
	err = filepath.WalkDir(absRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if !d.IsDir() {
			return nil
		}

		if shouldSkipDir(d.Name(), path, absRoot, excludeSet) {
			return filepath.SkipDir
		}

		gitPath := filepath.Join(path, ".git")
		if _, err := os.Stat(gitPath); err == nil {
			relPath, err := filepath.Rel(absRoot, path)
			if err != nil || relPath == "." {
				relPath = filepath.Base(path)
			}

			repo := domain.Repository{
				Name: relPath,
				Path: path,
			}
			if tags, ok := repoTags[path]; ok {
				repo.Tags = tags
			}

			repos = append(repos, repo)
			return nil
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

func shouldSkipDir(name, path, root string, excludeSet map[string]struct{}) bool {
	if _, ok := excludeSet[name]; !ok {
		return false
	}
	if name != ".git" {
		return true
	}
	return path != filepath.Join(root, ".git")
}

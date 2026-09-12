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

		if shouldSkipDir(d.Name(), excludeSet) {
			return filepath.SkipDir
		}

		gitPath := filepath.Join(path, ".git")
		if isGitRepository(gitPath) {
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
			if path != absRoot {
				return filepath.SkipDir
			}
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

func isGitRepository(gitPath string) bool {
	info, err := os.Lstat(gitPath)
	if err != nil || info.Mode()&os.ModeSymlink != 0 {
		return false
	}
	if info.IsDir() {
		return true
	}
	if !info.Mode().IsRegular() || info.Size() > 4096 {
		return false
	}
	data, err := os.ReadFile(gitPath)
	if err != nil {
		return false
	}
	value := strings.TrimSpace(string(data))
	if !strings.HasPrefix(value, "gitdir: ") || strings.ContainsAny(value, "\x00\r\n") {
		return false
	}
	target := strings.TrimSpace(strings.TrimPrefix(value, "gitdir: "))
	if !filepath.IsAbs(target) {
		target = filepath.Join(filepath.Dir(gitPath), target)
	}
	targetInfo, err := os.Stat(filepath.Clean(target))
	return err == nil && targetInfo.IsDir()
}

// AlwaysExcluded lists directories that never contain a repository worth
// showing. Users extend this through the scan_excludes config key; they cannot
// remove entries, because walking into them is always wasted work.
var AlwaysExcluded = map[string]struct{}{
	".git":           {},
	"node_modules":   {},
	".venv":          {},
	"vendor":         {},
	".idea":          {},
	".vscode":        {},
	"dist":           {},
	"target":         {},
	"bin":            {},
	"obj":            {},
	"__pycache__":    {},
	".tox":           {},
	".eggs":          {},
	".gradle":        {},
	".terraform":     {},
	"bazel-bin":      {},
	"bazel-out":      {},
	"bazel-testlogs": {},
}

func shouldSkipDir(name string, excludeSet map[string]struct{}) bool {
	if _, ok := AlwaysExcluded[name]; ok {
		return true
	}
	_, excluded := excludeSet[name]
	return excluded
}

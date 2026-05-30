package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/JoaoOliveira889/monogit/internal/domain"
)

const startupCacheFileName = "startup_cache.json"

type startupCacheFile struct {
	Roots map[string]startupCacheEntry `json:"roots"`
}

type startupCacheEntry struct {
	UpdatedAt time.Time          `json:"updated_at"`
	Repos     []startupCacheRepo `json:"repos"`
}

type startupCacheRepo struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

func GetStartupCachePath() string {
	return filepath.Join(filepath.Dir(GetConfigPath()), startupCacheFileName)
}

func LoadStartupRepos(rootPath string, repoTags map[string][]string) ([]domain.Repository, error) {
	absRoot, err := filepath.Abs(rootPath)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(GetStartupCachePath())
	if err != nil {
		return nil, err
	}

	var cache startupCacheFile
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, err
	}

	entry, ok := cache.Roots[absRoot]
	if !ok {
		return nil, os.ErrNotExist
	}

	repos := make([]domain.Repository, 0, len(entry.Repos))
	for _, cached := range entry.Repos {
		repo := domain.Repository{
			Name: cached.Name,
			Path: cached.Path,
		}
		if tags, ok := repoTags[cached.Path]; ok {
			repo.Tags = append([]string(nil), tags...)
		}
		repos = append(repos, repo)
	}

	return repos, nil
}

func SaveStartupRepos(rootPath string, repos []domain.Repository) error {
	absRoot, err := filepath.Abs(rootPath)
	if err != nil {
		return err
	}

	path := GetStartupCachePath()
	cache := startupCacheFile{}

	if data, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(data, &cache)
	}
	if cache.Roots == nil {
		cache.Roots = make(map[string]startupCacheEntry)
	}

	cachedRepos := make([]startupCacheRepo, 0, len(repos))
	for _, repo := range repos {
		cachedRepos = append(cachedRepos, startupCacheRepo{
			Name: repo.Name,
			Path: repo.Path,
		})
	}

	cache.Roots[absRoot] = startupCacheEntry{
		UpdatedAt: time.Now().UTC(),
		Repos:     cachedRepos,
	}

	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

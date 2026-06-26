package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	LeftPanelRatio float64             `json:"left_panel_ratio"`
	RepoTags       map[string][]string `json:"repo_tags"`
	MergeTool      string              `json:"merge_tool"`
	ScanExcludes   []string            `json:"scan_excludes"`
	Concurrency    int                 `json:"concurrency"`
	Theme          string              `json:"theme"`
}

var defaultConfig = Config{
	LeftPanelRatio: 0.30,
	RepoTags:       make(map[string][]string),
	ScanExcludes:   []string{"node_modules", "vendor", ".git", ".idea", ".vscode", "dist", "build", "coverage", ".next", ".turbo"},
	Concurrency:    5,
	Theme:          "Tokyo Night",
}

func GetConfigPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		homeDir, _ := os.UserHomeDir()
		configDir = filepath.Join(homeDir, ".config")
	}
	return filepath.Join(configDir, "monogit", "config.json")
}

func LoadConfig() Config {
	path := GetConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return defaultConfig
	}
	if info, statErr := os.Stat(path); statErr == nil && info.Mode().Perm() != 0600 {
		_ = os.Chmod(path, 0600)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return defaultConfig
	}

	if cfg.LeftPanelRatio < 0.1 || cfg.LeftPanelRatio > 0.9 {
		cfg.LeftPanelRatio = defaultConfig.LeftPanelRatio
	}

	if cfg.RepoTags == nil {
		cfg.RepoTags = make(map[string][]string)
	}
	if len(cfg.ScanExcludes) == 0 {
		cfg.ScanExcludes = append([]string(nil), defaultConfig.ScanExcludes...)
	}
	if cfg.Concurrency < 1 || cfg.Concurrency > 50 {
		cfg.Concurrency = defaultConfig.Concurrency
	}
	if cfg.Theme == "" {
		cfg.Theme = defaultConfig.Theme
	}
	return cfg
}

func SaveConfig(cfg Config) error {
	path := GetConfigPath()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

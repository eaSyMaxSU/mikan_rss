package main

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config is the whole program configuration, loaded from a YAML file.
type Config struct {
	PikPak struct {
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"pikpak"`

	// FolderID is the PikPak folder the torrents are added into.
	// Leave empty to drop them into PikPak's default "My Pack" / download root.
	FolderID string `yaml:"folder_id"`

	// RSS is the list of mikan RSS feed URLs to pull from.
	RSS []string `yaml:"rss"`

	// StateFile remembers which torrents were already submitted, so re-running
	// does not add duplicates. Defaults to "seen.json" in the current folder.
	StateFile string `yaml:"state_file"`

	// LogFile is where the run log is written (in addition to stdout).
	// Defaults to "mikan_rss.log" in the current folder.
	LogFile string `yaml:"log_file"`
}

// loadConfig reads and validates the YAML config file.
func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config %s: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config %s: %w", path, err)
	}

	cfg.PikPak.Username = strings.TrimSpace(cfg.PikPak.Username)
	cfg.PikPak.Password = strings.TrimSpace(cfg.PikPak.Password)
	cfg.FolderID = strings.TrimSpace(cfg.FolderID)

	if cfg.PikPak.Username == "" || cfg.PikPak.Password == "" {
		return nil, fmt.Errorf("pikpak.username and pikpak.password are required in %s", path)
	}

	if cfg.StateFile == "" {
		cfg.StateFile = "seen.json"
	}
	if cfg.LogFile == "" {
		cfg.LogFile = "mikan_rss.log"
	}

	// Tidy up the feed list: trim blanks and drop empties.
	feeds := cfg.RSS[:0]
	for _, u := range cfg.RSS {
		if u = strings.TrimSpace(u); u != "" {
			feeds = append(feeds, u)
		}
	}
	cfg.RSS = feeds

	return &cfg, nil
}

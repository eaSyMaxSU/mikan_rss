package config

import (
	"fmt"
	"os"
	"strings"

	"mikan-rss/internal/filter"
	"mikan-rss/internal/qbittorrent"

	"gopkg.in/yaml.v3"
)

// Config is the whole program configuration, loaded from a YAML file.
type Config struct {
	RSS         []string            `yaml:"rss"`
	Filters     filter.Rules        `yaml:"filters"`
	QBitTorrent qbittorrent.Config  `yaml:"qbittorrent"`
	StateFile   string              `yaml:"state_file"`
	LogFile     string              `yaml:"log_file"`
}

// Load reads and validates the YAML config file.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config %s: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config %s: %w", path, err)
	}

	if cfg.StateFile == "" {
		cfg.StateFile = "data/seen.json"
	}
	if cfg.LogFile == "" {
		cfg.LogFile = "data/mikan_rss.log"
	}

	feeds := cfg.RSS[:0]
	for _, u := range cfg.RSS {
		if u = strings.TrimSpace(u); u != "" {
			feeds = append(feeds, u)
		}
	}
	cfg.RSS = feeds

	cfg.Filters.Normalize()
	if err := cfg.Filters.Validate(); err != nil {
		return nil, err
	}

	cfg.QBitTorrent.URL = strings.TrimSpace(cfg.QBitTorrent.URL)
	cfg.QBitTorrent.Username = strings.TrimSpace(cfg.QBitTorrent.Username)
	cfg.QBitTorrent.Password = strings.TrimSpace(cfg.QBitTorrent.Password)
	cfg.QBitTorrent.Category = strings.TrimSpace(cfg.QBitTorrent.Category)
	cfg.QBitTorrent.SavePath = strings.TrimSpace(cfg.QBitTorrent.SavePath)

	return &cfg, nil
}
package app

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"mikan-rss/internal/config"
	"mikan-rss/internal/qbittorrent"
	"mikan-rss/internal/rss"
	"mikan-rss/internal/state"
)

// Run pulls every feed and adds torrents that pass the filters to qBittorrent.
func Run(logger *log.Logger, cfg *config.Config, showAll, printOnly, verbose bool) error {
	if len(cfg.RSS) == 0 {
		logger.Printf("no rss feeds configured, nothing to do")
		return nil
	}

	var qbt *qbittorrent.Client
	if printOnly {
		logger.Printf("print mode: torrent URLs go to stdout")
	} else {
		client, err := qbittorrent.NewClient(cfg.QBitTorrent)
		if err != nil {
			return err
		}
		qbt = client
		logger.Printf("connected to qBittorrent at %s", cfg.QBitTorrent.URL)
	}

	seen, err := state.Load(cfg.StateFile)
	if err != nil {
		return err
	}

	var matched, added, skippedSeen, skippedFilter, noTorrent, failed int

	for _, feedURL := range cfg.RSS {
		title, items, err := rss.Fetch(feedURL)
		if err != nil {
			logger.Printf("[feed] %s -> ERROR: %v", feedURL, err)
			continue
		}
		logger.Printf("[feed] %s (%q): %d items", feedURL, title, len(items))

		for _, it := range items {
			url := it.TorrentURL()
			if url == "" {
				noTorrent++
				continue
			}

			ok, err := cfg.Filters.Matches(it.Title)
			if err != nil {
				return err
			}
			if !ok {
				skippedFilter++
				logger.Printf("  - filtered: %s", it.Title)
				continue
			}

			matched++
			if !showAll && seen[it.ID()] {
				skippedSeen++
				continue
			}

			if printOnly {
				if verbose {
					fmt.Printf("%s\t%s\n", it.Title, url)
				} else {
					fmt.Println(url)
				}
			} else if err := qbt.AddTorrent(url, cfg.QBitTorrent); err != nil {
				logger.Printf("  ! add failed: %s -> %v", it.Title, err)
				failed++
				continue
			}

			seen[it.ID()] = true
			added++
			logger.Printf("  + added: %s", it.Title)
		}
	}

	if !showAll {
		if err := state.Save(cfg.StateFile, seen); err != nil {
			return err
		}
	}

	logger.Printf(
		"done: %d added, %d matched, %d skipped (seen), %d skipped (filter), %d without torrent, %d failed",
		added, matched, skippedSeen, skippedFilter, noTorrent, failed,
	)
	return nil
}

// SetupLogger writes to stderr and the log file (appended).
func SetupLogger(path string) (*log.Logger, func(), error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, nil, err
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, nil, err
	}
	w := io.MultiWriter(os.Stderr, f)
	logger := log.New(w, "", log.LstdFlags)
	return logger, func() { _ = f.Close() }, nil
}
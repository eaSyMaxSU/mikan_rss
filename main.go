// Command mikan-rss adds torrents from mikan RSS feeds into a PikPak folder.
//
// It runs once and exits — it does not schedule itself. Run it manually (or
// from your own cron/launchd) whenever you want to pull new episodes.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"io"
	"log"
	"os"
	"sort"
	"time"

	"mikan-rss/pikpakgo"
)

// pauseBetweenTasks is a small courtesy delay between offline-download submits.
const pauseBetweenTasks = 1 * time.Second

func main() {
	configPath := flag.String("config", "config.yaml", "path to the YAML config file")
	listFolders := flag.Bool("list", false, "list PikPak folders (with their IDs) and exit; use to find folder_id")
	listParent := flag.String("parent", "", "parent folder ID to list when using -list (default: root)")
	flag.Parse()

	cfg, err := loadConfig(*configPath)
	if err != nil {
		// Config isn't loaded yet, so just log to stderr.
		log.Fatalf("config error: %v", err)
	}

	logger, closeLog, err := setupLogger(cfg.LogFile)
	if err != nil {
		log.Fatalf("cannot open log file %s: %v", cfg.LogFile, err)
	}
	defer closeLog()

	client, err := pikpakgo.NewPikPakClient(cfg.PikPak.Username, cfg.PikPak.Password)
	if err != nil {
		logger.Fatalf("create pikpak client: %v", err)
	}
	logger.Printf("logging in to PikPak as %s ...", cfg.PikPak.Username)
	if err := client.Login(); err != nil {
		logger.Fatalf("pikpak login failed: %v", err)
	}
	logger.Printf("login ok")

	if *listFolders {
		if err := printFolders(logger, client, *listParent); err != nil {
			logger.Fatalf("list folders: %v", err)
		}
		return
	}

	if err := run(logger, client, cfg); err != nil {
		logger.Fatalf("run failed: %v", err)
	}
}

// run pulls every feed and submits any new torrents to the configured folder.
func run(logger *log.Logger, client *pikpakgo.PikPakClient, cfg *Config) error {
	if len(cfg.RSS) == 0 {
		logger.Printf("no rss feeds configured, nothing to do")
		return nil
	}

	seen, err := loadSeen(cfg.StateFile)
	if err != nil {
		return err
	}

	var added, skipped, failed int

feeds:
	for _, feedURL := range cfg.RSS {
		title, items, err := fetchFeed(feedURL)
		if err != nil {
			logger.Printf("[feed] %s -> ERROR: %v", feedURL, err)
			continue
		}
		logger.Printf("[feed] %s (%q): %d items", feedURL, title, len(items))

		for _, it := range items {
			url := it.torrentURL()
			if url == "" {
				continue // not a torrent entry
			}
			if seen[it.id()] {
				skipped++
				continue
			}

			_, err := client.OfflineDownload(it.Title, url, cfg.FolderID)
			if err != nil {
				// Stop early on account-wide limits — retrying won't help today.
				if errors.Is(err, pikpakgo.ErrDailyCreateLimit) {
					logger.Printf("  ! daily offline-download limit reached, stopping: %v", err)
					break feeds
				}
				if errors.Is(err, pikpakgo.ErrSpaceNotEnough) {
					logger.Printf("  ! pikpak storage full, stopping: %v", err)
					break feeds
				}
				logger.Printf("  ! add failed: %s -> %v", it.Title, err)
				failed++
				continue
			}

			seen[it.id()] = true
			added++
			logger.Printf("  + added: %s", it.Title)
			time.Sleep(pauseBetweenTasks)
		}
	}

	if err := saveSeen(cfg.StateFile, seen); err != nil {
		return err
	}

	logger.Printf("done: %d added, %d skipped (already seen), %d failed", added, skipped, failed)
	return nil
}

// printFolders lists folders under parentID so the user can find a folder_id.
func printFolders(logger *log.Logger, client *pikpakgo.PikPakClient, parentID string) error {
	files, err := client.FileListAll(parentID)
	if err != nil {
		return err
	}
	where := "root"
	if parentID != "" {
		where = parentID
	}
	logger.Printf("folders under %s:", where)
	for _, f := range files {
		if f.Kind == pikpakgo.KindOfFolder {
			logger.Printf("  %s  %s", f.ID, f.Name)
		}
	}
	return nil
}

// --- run log -------------------------------------------------------------

// setupLogger writes to both stdout and the log file (appended) in the cwd.
func setupLogger(path string) (*log.Logger, func(), error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, nil, err
	}
	w := io.MultiWriter(os.Stdout, f)
	logger := log.New(w, "", log.LstdFlags)
	return logger, func() { _ = f.Close() }, nil
}

// --- dedup state ---------------------------------------------------------

// loadSeen reads the set of already-submitted torrent IDs. A missing file is
// treated as an empty set.
func loadSeen(path string) (map[string]bool, error) {
	seen := map[string]bool{}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return seen, nil
	}
	if err != nil {
		return nil, err
	}
	var ids []string
	if err := json.Unmarshal(data, &ids); err != nil {
		return nil, err
	}
	for _, id := range ids {
		seen[id] = true
	}
	return seen, nil
}

// saveSeen writes the set back as a sorted JSON array.
func saveSeen(path string, seen map[string]bool) error {
	ids := make([]string, 0, len(seen))
	for id := range seen {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	data, err := json.MarshalIndent(ids, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

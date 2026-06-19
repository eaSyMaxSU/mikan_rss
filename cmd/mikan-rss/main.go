// Command mikan-rss fetches mikan RSS feeds, applies filter rules, and adds
// matching torrents to qBittorrent.
package main

import (
	"flag"
	"log"

	"mikan-rss/internal/app"
	"mikan-rss/internal/config"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to the YAML config file")
	showAll := flag.Bool("all", false, "process torrents even if already seen")
	printOnly := flag.Bool("print", false, "print torrent URLs to stdout instead of adding to qBittorrent")
	verbose := flag.Bool("v", false, "with -print, show title and URL (tab-separated)")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	logger, closeLog, err := app.SetupLogger(cfg.LogFile)
	if err != nil {
		log.Fatalf("cannot open log file %s: %v", cfg.LogFile, err)
	}
	defer closeLog()

	if err := app.Run(logger, cfg, *showAll, *printOnly, *verbose); err != nil {
		logger.Fatalf("run failed: %v", err)
	}
}
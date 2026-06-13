package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// rssFeed / rssItem cover just the fields we need from a mikan RSS 2.0 feed.
type rssFeed struct {
	Channel struct {
		Title string    `xml:"title"`
		Items []rssItem `xml:"item"`
	} `xml:"channel"`
}

type rssItem struct {
	Title     string `xml:"title"`
	Link      string `xml:"link"`
	GUID      string `xml:"guid"`
	Enclosure struct {
		URL  string `xml:"url,attr"`
		Type string `xml:"type,attr"`
	} `xml:"enclosure"`
}

// torrentURL is the .torrent download link PikPak should fetch.
func (it rssItem) torrentURL() string {
	return strings.TrimSpace(it.Enclosure.URL)
}

// id is the stable identifier used for de-duplication across runs.
func (it rssItem) id() string {
	if u := it.torrentURL(); u != "" {
		return u
	}
	if g := strings.TrimSpace(it.GUID); g != "" {
		return g
	}
	return strings.TrimSpace(it.Link)
}

var httpClient = &http.Client{Timeout: 30 * time.Second}

// fetchFeed downloads and parses one RSS feed, returning its channel title and items.
func fetchFeed(url string) (string, []rssItem, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", nil, err
	}
	req.Header.Set("User-Agent", "mikan-rss/1.0 (+https://github.com)")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("unexpected status %s", resp.Status)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 32<<20)) // 32 MiB safety cap
	if err != nil {
		return "", nil, err
	}

	var feed rssFeed
	if err := xml.Unmarshal(body, &feed); err != nil {
		return "", nil, fmt.Errorf("parse rss: %w", err)
	}

	return strings.TrimSpace(feed.Channel.Title), feed.Channel.Items, nil
}

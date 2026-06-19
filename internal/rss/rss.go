package rss

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Item is one entry from a mikan RSS feed.
type Item struct {
	Title     string `xml:"title"`
	Link      string `xml:"link"`
	GUID      string `xml:"guid"`
	Enclosure struct {
		URL  string `xml:"url,attr"`
		Type string `xml:"type,attr"`
	} `xml:"enclosure"`
}

// TorrentURL is the .torrent download link for this RSS item.
func (it Item) TorrentURL() string {
	return strings.TrimSpace(it.Enclosure.URL)
}

// ID is the stable identifier used for de-duplication across runs.
func (it Item) ID() string {
	if u := it.TorrentURL(); u != "" {
		return u
	}
	if g := strings.TrimSpace(it.GUID); g != "" {
		return g
	}
	return strings.TrimSpace(it.Link)
}

type feed struct {
	Channel struct {
		Title string `xml:"title"`
		Items []Item `xml:"item"`
	} `xml:"channel"`
}

var httpClient = &http.Client{Timeout: 30 * time.Second}

// Fetch downloads and parses one RSS feed, returning its channel title and items.
func Fetch(url string) (string, []Item, error) {
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

	body, err := io.ReadAll(io.LimitReader(resp.Body, 32<<20))
	if err != nil {
		return "", nil, err
	}

	var parsed feed
	if err := xml.Unmarshal(body, &parsed); err != nil {
		return "", nil, fmt.Errorf("parse rss: %w", err)
	}

	return strings.TrimSpace(parsed.Channel.Title), parsed.Channel.Items, nil
}
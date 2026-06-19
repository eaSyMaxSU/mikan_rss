package qbittorrent

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
)

// Config holds Web UI connection settings.
type Config struct {
	URL      string `yaml:"url"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Category string `yaml:"category"`
	SavePath string `yaml:"save_path"`
	Paused   bool   `yaml:"paused"`
}

// Client talks to the qBittorrent Web API.
type Client struct {
	base   string
	client *http.Client
}

// NewClient connects to the qBittorrent Web UI.
func NewClient(cfg Config) (*Client, error) {
	base := strings.TrimRight(strings.TrimSpace(cfg.URL), "/")
	if base == "" {
		return nil, fmt.Errorf("qbittorrent.url is required")
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	c := &Client{
		base: base,
		client: &http.Client{
			Timeout: 30 * time.Second,
			Jar:     jar,
		},
	}

	username := strings.TrimSpace(cfg.Username)
	password := strings.TrimSpace(cfg.Password)
	if username != "" || password != "" {
		if err := c.login(username, password); err != nil {
			return nil, err
		}
	}

	return c, nil
}

func (c *Client) login(username, password string) error {
	body := url.Values{
		"username": {username},
		"password": {password},
	}
	resp, err := c.client.PostForm(c.base+"/api/v2/auth/login", body)
	if err != nil {
		return fmt.Errorf("qbittorrent login: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("qbittorrent login: %w", err)
	}
	text := strings.TrimSpace(string(data))
	switch {
	case resp.StatusCode == http.StatusNoContent:
		return nil
	case resp.StatusCode == http.StatusOK && text == "Ok.":
		return nil
	default:
		if text == "" {
			text = resp.Status
		}
		return fmt.Errorf("qbittorrent login failed: %s", text)
	}
}

// AddTorrent submits a torrent URL to qBittorrent.
func (c *Client) AddTorrent(torrentURL string, cfg Config) error {
	body := url.Values{"urls": {torrentURL}}
	if cat := strings.TrimSpace(cfg.Category); cat != "" {
		body.Set("category", cat)
	}
	if path := strings.TrimSpace(cfg.SavePath); path != "" {
		body.Set("savepath", path)
	}
	if cfg.Paused {
		body.Set("paused", "true")
	}

	resp, err := c.client.PostForm(c.base+"/api/v2/torrents/add", body)
	if err != nil {
		return fmt.Errorf("qbittorrent add: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("qbittorrent add: %w", err)
	}
	text := strings.TrimSpace(string(data))
	switch resp.StatusCode {
	case http.StatusOK:
		if text == "Ok." {
			return nil
		}
	case http.StatusAccepted:
		var result struct {
			FailureCount int `json:"failure_count"`
		}
		if err := json.Unmarshal(data, &result); err == nil && result.FailureCount == 0 {
			return nil
		}
	}

	if text == "" {
		text = resp.Status
	}
	return fmt.Errorf("qbittorrent add failed: %s", text)
}
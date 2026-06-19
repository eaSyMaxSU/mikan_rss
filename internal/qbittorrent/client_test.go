package qbittorrent

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClient(t *testing.T) {
	var added []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v2/auth/login":
			if err := r.ParseForm(); err != nil {
				t.Fatal(err)
			}
			if r.Form.Get("username") != "admin" || r.Form.Get("password") != "secret" {
				http.Error(w, "Fails.", http.StatusOK)
				return
			}
			w.Write([]byte("Ok."))
		case "/api/v2/torrents/add":
			if err := r.ParseForm(); err != nil {
				t.Fatal(err)
			}
			added = append(added, r.Form.Get("urls"))
			if r.Form.Get("category") != "anime" {
				t.Errorf("category = %q, want anime", r.Form.Get("category"))
			}
			if r.Form.Get("savepath") != "/downloads" {
				t.Errorf("savepath = %q, want /downloads", r.Form.Get("savepath"))
			}
			if r.Form.Get("paused") != "true" {
				t.Errorf("paused = %q, want true", r.Form.Get("paused"))
			}
			w.Write([]byte("Ok."))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	cfg := Config{
		URL:      srv.URL,
		Username: "admin",
		Password: "secret",
		Category: "anime",
		SavePath: "/downloads",
		Paused:   true,
	}
	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	url := "https://mikanani.me/Download/20240101/aaaa1111.torrent"
	if err := client.AddTorrent(url, cfg); err != nil {
		t.Fatalf("AddTorrent: %v", err)
	}
	if len(added) != 1 || added[0] != url {
		t.Fatalf("added = %v, want [%s]", added, url)
	}
}

func TestClientLoginFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Fails."))
	}))
	defer srv.Close()

	_, err := NewClient(Config{
		URL:      srv.URL,
		Username: "bad",
		Password: "bad",
	})
	if err == nil || !strings.Contains(err.Error(), "login failed") {
		t.Fatalf("expected login error, got %v", err)
	}
}
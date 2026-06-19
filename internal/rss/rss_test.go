package rss

import (
	"encoding/xml"
	"testing"
)

const sampleMikanRSS = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Mikan Project - 某番组</title>
    <link>https://mikanani.me</link>
    <description>Mikan Project - 某番组</description>
    <item>
      <guid isPermaLink="false">aaaa1111</guid>
      <link>https://mikanani.me/Home/Episode/aaaa1111</link>
      <title>[字幕组] 某番组 - 01 [1080p]</title>
      <enclosure type="application/x-bittorrent" length="123456"
        url="https://mikanani.me/Download/20240101/aaaa1111.torrent"/>
      <torrent xmlns="https://mikanani.me/0.1/">
        <link>https://mikanani.me/Home/Episode/aaaa1111</link>
        <contentLength>123456</contentLength>
        <pubDate>2024-01-01T12:00:00</pubDate>
      </torrent>
    </item>
    <item>
      <guid isPermaLink="false">bbbb2222</guid>
      <link>https://mikanani.me/Home/Episode/bbbb2222</link>
      <title>[字幕组] 某番组 - 02 [1080p]</title>
      <enclosure type="application/x-bittorrent" length="234567"
        url="https://mikanani.me/Download/20240108/bbbb2222.torrent"/>
    </item>
  </channel>
</rss>`

func TestParseMikanRSS(t *testing.T) {
	var parsed feed
	if err := xml.Unmarshal([]byte(sampleMikanRSS), &parsed); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got, want := parsed.Channel.Title, "Mikan Project - 某番组"; got != want {
		t.Errorf("channel title = %q, want %q", got, want)
	}
	if len(parsed.Channel.Items) != 2 {
		t.Fatalf("got %d items, want 2", len(parsed.Channel.Items))
	}

	first := parsed.Channel.Items[0]
	if want := "https://mikanani.me/Download/20240101/aaaa1111.torrent"; first.TorrentURL() != want {
		t.Errorf("TorrentURL = %q, want %q", first.TorrentURL(), want)
	}
	if first.ID() != first.TorrentURL() {
		t.Errorf("ID should fall back to torrent URL, got %q", first.ID())
	}
	if first.Title != "[字幕组] 某番组 - 01 [1080p]" {
		t.Errorf("unexpected title %q", first.Title)
	}
}
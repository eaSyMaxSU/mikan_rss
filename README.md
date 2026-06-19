# mikan-rss

A minimal Go program that fetches [mikan](https://mikanani.me) RSS feeds,
applies title filter rules, and adds matching torrents to
[qBittorrent](https://www.qbittorrent.org) via its Web API.

## Project layout

```
cmd/mikan-rss/          program entry point
internal/
  app/                  orchestration and logging
  config/               YAML config loading
  filter/               title include/exclude rules
  qbittorrent/          qBittorrent Web API client
  rss/                  mikan RSS fetch and parse
  state/                seen-torrent persistence
configs/                example configuration
data/                   runtime state and logs (git-ignored)
```

## Setup

Requires Go 1.21+ and qBittorrent with Web UI enabled.

```sh
go mod tidy
go build -o bin/mikan-rss ./cmd/mikan-rss

cp configs/config.example.yaml config.yaml
# edit config.yaml with your feeds and qBittorrent login

./bin/mikan-rss
```

Enable qBittorrent Web UI under **Tools → Preferences → Web UI**.

## Configuration

See `configs/config.example.yaml`. Runtime files default to `data/seen.json`
and `data/mikan_rss.log`.

### Filter rules

- `include`: when non-empty, the title must match **at least one** pattern.
- `exclude`: if **any** pattern matches, the item is dropped.
- `exclude_unless`: drop when `pattern` matches but `unless` does not. Useful
  for blocking `繁日双语` while keeping `简繁内封` (do not put plain `繁` in
  `exclude`, because that also blocks `简繁`).
- `regex`: when `true`, patterns are regular expressions; otherwise
  case-insensitive substring matches.

## Usage

```
./bin/mikan-rss                 # add new torrents to qBittorrent
./bin/mikan-rss -config x.yaml  # use a different config file
./bin/mikan-rss -all            # process all matches, including already-seen
./bin/mikan-rss -print          # print URLs to stdout instead of adding
./bin/mikan-rss -print -v       # print "title<TAB>url"
```

Point your own `cron` / `launchd` at the binary to run on a schedule.
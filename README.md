# mikan-rss

A minimal Go program that adds torrents from [mikan](https://mikanani.me) RSS
feeds into a [PikPak](https://mypikpak.com) folder as offline downloads.

It's a small, command-line reimagining of
[PikPak-RSS](https://github.com/okami-horo/PikPak-RSS): no GUI, no scheduler.
You run it manually; it pulls every feed once and exits.

## What it does

1. Reads everything from a single `config.yaml` (PikPak account, target folder
   ID, and the list of RSS feeds).
2. Logs in to PikPak (using the patched [`pikpak-go`](pikpakgo/) client, which
   handles PikPak's current captcha-token sign-in flow).
3. For each feed, submits every new `.torrent` to your PikPak folder as an
   offline download.
4. Remembers what it already added (`seen.json`) so re-running never creates
   duplicates.
5. Writes a run log to `mikan_rss.log` in the current folder (and to stdout).

It is **not** automatic — it only does work when you run it.

## Setup

Requires Go 1.21+.

```sh
# 1. fetch dependencies & build
go mod tidy
go build -o mikan-rss .

# 2. create your config
cp config.example.yaml config.yaml
#   ...then edit config.yaml with your PikPak login and feeds

# 3. find your target folder's ID (lists folders in your PikPak root)
./mikan-rss -list

# 4. put that ID in config.yaml as folder_id, then run
./mikan-rss
```

## Configuration (`config.yaml`)

```yaml
pikpak:
  username: "your_email_or_phone"
  password: "your_password"

folder_id: "PIKPAK_FOLDER_ID"   # empty = default download location

rss:
  - "https://mikanani.me/RSS/MyBangumi?token=YOUR_TOKEN"
  - "https://mikanani.me/RSS/Bangumi?bangumiId=123&subgroupid=456"
```

Optional keys: `state_file` (default `seen.json`) and `log_file`
(default `mikan_rss.log`).

## Usage

```
./mikan-rss                 # run once: add new torrents from all feeds
./mikan-rss -config x.yaml  # use a different config file
./mikan-rss -list           # list folders in your PikPak root + their IDs
./mikan-rss -list -parent <folderID>   # list folders inside another folder
```

To run it on a schedule, point your own `cron` / `launchd` / Task Scheduler at
the binary — the program itself never schedules anything.

## Notes

- The dedup state is keyed on the torrent download URL, stored in `seen.json`.
  Delete that file if you want to re-add everything.
- PikPak enforces a daily offline-download creation limit; if it's hit the run
  stops early and logs it. Just run again the next day.
- `config.yaml`, `seen.json`, and `*.log` are git-ignored so you don't commit
  your password.

## Credits

The `pikpakgo/` package is vendored from
[AnimeX](https://github.com/YinBuLiao/AnimeX)'s patched copy of
[`kanghengliu/pikpak-go`](https://github.com/kanghengliu/pikpak-go).

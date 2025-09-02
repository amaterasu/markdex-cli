# markdex CLI

Command-line interface for interacting with Markdex bookmarks.

## Install / Build Locally

Prerequisites: Go 1.22+.

1. Clone repo:
   git clone https://github.com/amaterasu/markdex-cli.git
   cd markdex-cli

2. Build binary into bin/:
   make build
   ls bin/markdex

3. Install into your GOPATH/bin (usually ~/go/bin) so it's on PATH:
   make install
   markdex --help

Or without Makefile:
   go install -ldflags "-s -w" ./...

## Version Info Injection
Makefile injects version, commit, and build date. To see it (after adding version command):
   markdex version

## Usage Examples

Set API base URL:
   markdex config set --api https://your.api.example
Show config path:
   markdex config path
List bookmarks:
   markdex list
   markdex ls
Search (now supports positional argument or -s flag):
   markdex list rust
   markdex ls rust
   markdex list -s rust
   markdex ls -s rust
Filter by tag:
   markdex list -t programming
   markdex ls -t programming
Open by index:
   markdex open 3
Fuzzy pick (requires fzf):
   markdex pick -s golang
Open by hash prefix:
   markdex open-hash abc

Add bookmark (AI enrichment):
   markdex add --ai --source-file inbox.md https://example.com/some/page

Add bookmark manually with fields:
   markdex add -T "Some Title" -t web,reference -d "Some Description" -f inbox.md https://example.com/ref

Output created bookmark JSON:
   markdex add --ai --json https://example.com/interesting

JSON output:
   markdex list --json | jq '.[] | {title, url}'

## Config File
Stored at: ~/.config/markdex/config.toml

Example:
```toml
apiBase = "https://your.api.example"
```

## Cache
Bookmark list cache stored under your OS user cache dir (5 min TTL). Use --no-cache to bypass.

## Cross Compilation
Example:
   GOOS=linux GOARCH=amd64 make build
Binary appears at bin/markdex (rename if you want OS/arch suffix).

## Contributing
PRs welcome. Please run "go fmt ./..." and ensure "go build ./..." succeeds.

## License
MIT (add LICENSE file if not present).

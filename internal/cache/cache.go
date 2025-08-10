package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/amaterasu/markdex-cli/internal/api"
)

type diskCache struct {
	Path string
	ttl  time.Duration
}

type entry struct {
	Items []api.Bookmark `json:"items"`
	TS    int64          `json:"ts"`
}

func New() *diskCache {
	return &diskCache{Path: filepath.Join(userCacheDir(), "bookmarks.json"), ttl: 5 * time.Minute}
}

func (c *diskCache) Read() ([]api.Bookmark, bool) {
	b, err := os.ReadFile(c.Path)
	if err != nil {
		return nil, false
	}
	var e entry
	if json.Unmarshal(b, &e) != nil {
		return nil, false
	}
	if time.Since(time.Unix(e.TS, 0)) > c.ttl {
		return nil, false
	}
	return e.Items, true
}

func (c *diskCache) Write(items []api.Bookmark) {
	_ = os.MkdirAll(filepath.Dir(c.Path), 0o755)
	b, _ := json.Marshal(entry{Items: items, TS: time.Now().Unix()})
	_ = os.WriteFile(c.Path, b, 0o644)
}

func userCacheDir() string {
	if d, err := os.UserCacheDir(); err == nil {
		return filepath.Join(d, "markdex")
	}
	return ".markdex-cache"
}

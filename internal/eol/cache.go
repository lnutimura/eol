package eol

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type cacheEntry struct {
	ETag     string          `json:"etag"`
	Body     json.RawMessage `json:"body"`
	CachedAt time.Time       `json:"cached_at"`
}

func cacheKey(rawURL string) string {
	sum := sha256.Sum256([]byte(rawURL))
	return hex.EncodeToString(sum[:])
}

func (c *Client) cachePath(key string) string {
	dir := c.cacheDir()
	if dir == "" {
		return ""
	}
	return filepath.Join(dir, key+".json")
}

func (c *Client) cacheDir() string {
	if c.CacheDir != "" {
		return c.CacheDir
	}
	dir, err := os.UserCacheDir()
	if err != nil || dir == "" {
		return ""
	}
	return filepath.Join(dir, "eol")
}

func (c *Client) readCache(key string) *cacheEntry {
	path := c.cachePath(key)
	if path == "" {
		return nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var entry cacheEntry
	if err := json.Unmarshal(b, &entry); err != nil {
		return nil
	}
	return &entry
}

func (c *Client) writeCache(key string, entry cacheEntry) {
	path := c.cachePath(key)
	if path == "" {
		return
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return
	}
	b, err := json.Marshal(entry)
	if err != nil {
		return
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return
	}
	_ = os.Rename(tmp, path)
}

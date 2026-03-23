package services

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"coati/internal/core/ports"
)

type CachedData struct {
	Content   []byte    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

type CachedGistFetcher struct {
	delegate ports.GistFetcher
	cacheDir string
	ttl      time.Duration
}

func NewCachedGistFetcher(delegate ports.GistFetcher, cacheDir string, ttl time.Duration) *CachedGistFetcher {
	return &CachedGistFetcher{
		delegate: delegate,
		cacheDir: cacheDir,
		ttl:      ttl,
	}
}

func (c *CachedGistFetcher) Fetch(gistID, token string) ([]byte, error) {
	cachePath := filepath.Join(c.cacheDir, gistID+".json")

	// 1. Try Cache
	if data, err := os.ReadFile(cachePath); err == nil {
		var cached CachedData
		if err := json.Unmarshal(data, &cached); err == nil {
			if time.Since(cached.Timestamp) < c.ttl {
				return cached.Content, nil
			}
		}
	}

	// 2. Fetch from Delegate
	content, err := c.delegate.Fetch(gistID, token)
	if err != nil {
		return nil, err
	}

	// 3. Save to Cache (best-effort; errors are non-fatal)
	c.saveToCache(cachePath, content)

	return content, nil
}

func (c *CachedGistFetcher) saveToCache(cachePath string, content []byte) {
	if err := os.MkdirAll(c.cacheDir, 0755); err != nil {
		return
	}
	data, err := json.Marshal(CachedData{Content: content, Timestamp: time.Now()})
	if err != nil {
		return
	}
	_ = os.WriteFile(cachePath, data, 0644)
}

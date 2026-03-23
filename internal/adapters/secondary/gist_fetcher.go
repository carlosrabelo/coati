package secondary

// Package secondary provides filesystem and HTTP adapter implementations.

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type GistFetcher struct{}

func NewGistFetcher() *GistFetcher {
	return &GistFetcher{}
}

type GistResponse struct {
	Files map[string]struct {
		Content string `json:"content"`
	} `json:"files"`
}

func (g *GistFetcher) Fetch(gistID, token string) ([]byte, error) {
	url := fmt.Sprintf("https://api.github.com/gists/%s", gistID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if token != "" {
		req.Header.Set("Authorization", "token "+token)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch gist: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("github api error (status: %d): %s", resp.StatusCode, string(body))
	}

	var gist GistResponse
	if err := json.NewDecoder(resp.Body).Decode(&gist); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// We assume there's at least one file and we take the first one we find
	// In a real scenario we might want to let the user specify the filename or default to hosts.yml
	for _, file := range gist.Files {
		return []byte(file.Content), nil
	}

	return nil, fmt.Errorf("gist contains no files")
}

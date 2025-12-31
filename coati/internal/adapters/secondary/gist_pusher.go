package secondary

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type GistPusher struct{}

func NewGistPusher() *GistPusher {
	return &GistPusher{}
}

func (g *GistPusher) Push(gistID, token, gistFile string, content []byte) error {
	if gistFile == "" {
		name, err := g.firstFileName(gistID, token)
		if err != nil {
			return err
		}
		gistFile = name
	}

	body, err := json.Marshal(map[string]any{
		"files": map[string]any{
			gistFile: map[string]string{
				"content": string(content),
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to encode request: %w", err)
	}

	url := fmt.Sprintf("https://api.github.com/gists/%s", gistID)
	req, err := http.NewRequest("PATCH", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to push gist: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("github api error (status: %d): %s", resp.StatusCode, string(b))
	}

	return nil
}

func (g *GistPusher) firstFileName(gistID, token string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/gists/%s", gistID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch gist: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("github api error (status: %d): %s", resp.StatusCode, string(b))
	}

	var gist GistResponse
	if err := json.NewDecoder(resp.Body).Decode(&gist); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}
	for name := range gist.Files {
		return name, nil
	}
	return "", fmt.Errorf("gist contains no files")
}

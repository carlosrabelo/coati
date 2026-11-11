package services

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigManager_LoadConfig(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)
	cm := NewConfigManager(logger, nil)

	t.Run("missing file returns zero config", func(t *testing.T) {
		tmp := t.TempDir()
		t.Setenv("COATI_CONFIG_DIR", tmp)

		cfg, err := cm.LoadConfig()
		require.NoError(t, err)
		assert.Equal(t, "", cfg.GistID)
		assert.Equal(t, "", cfg.GitHubToken)
		assert.Nil(t, cfg.AllowedHooks)
	})

	t.Run("valid yaml", func(t *testing.T) {
		tmp := t.TempDir()
		t.Setenv("COATI_CONFIG_DIR", tmp)
		path := filepath.Join(tmp, "config.yaml")
		require.NoError(t, os.WriteFile(path, []byte(`
gist_id: gid-1
github_token: tok-1
allowed_hooks:
  - dnsmasq
`), 0600))

		cfg, err := cm.LoadConfig()
		require.NoError(t, err)
		assert.Equal(t, "gid-1", cfg.GistID)
		assert.Equal(t, "tok-1", cfg.GitHubToken)
		assert.Equal(t, []string{"dnsmasq"}, cfg.AllowedHooks)
	})

	t.Run("invalid yaml", func(t *testing.T) {
		tmp := t.TempDir()
		t.Setenv("COATI_CONFIG_DIR", tmp)
		path := filepath.Join(tmp, "config.yaml")
		require.NoError(t, os.WriteFile(path, []byte(": not yaml ["), 0600))

		_, err := cm.LoadConfig()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse config file:")
	})
}

func TestConfigManager_FetchGist(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)
	mock := new(MockGistFetcher)
	cm := NewConfigManager(logger, mock)

	t.Run("empty gist id", func(t *testing.T) {
		_, err := cm.FetchGist("", "tok", "")
		require.Error(t, err)
		assert.Equal(t, "gist ID is empty", err.Error())
	})

	t.Run("empty token", func(t *testing.T) {
		_, err := cm.FetchGist("gid", "", "")
		require.Error(t, err)
		assert.Equal(t, "github token is empty", err.Error())
	})

	t.Run("delegates to fetcher", func(t *testing.T) {
		mock.On("Fetch", "gid", "tok", "hosts.yaml").Return([]byte("ok"), nil).Once()
		got, err := cm.FetchGist("gid", "tok", "hosts.yaml")
		require.NoError(t, err)
		assert.Equal(t, []byte("ok"), got)
		mock.AssertExpectations(t)
	})
}

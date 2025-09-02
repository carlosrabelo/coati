package services

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigDir_DefaultAndOverride(t *testing.T) {
	t.Setenv("COATI_CONFIG_DIR", "")
	assert.Equal(t, "/etc/coati", ConfigDir())
	assert.Equal(t, "/etc/coati/config.yaml", AppConfigPath())

	tmp := t.TempDir()
	t.Setenv("COATI_CONFIG_DIR", tmp)
	assert.Equal(t, tmp, ConfigDir())
	assert.Equal(t, filepath.Join(tmp, "config.yaml"), AppConfigPath())
}

func TestDataDir_DefaultAndOverride(t *testing.T) {
	t.Setenv("COATI_DATA_DIR", "")
	assert.Equal(t, "/var/lib/coati", DataDir())
	assert.Equal(t, "/var/lib/coati/hosts.yaml", HostsYAMLPath())

	tmp := t.TempDir()
	t.Setenv("COATI_DATA_DIR", tmp)
	assert.Equal(t, tmp, DataDir())
	assert.Equal(t, filepath.Join(tmp, "hosts.yaml"), HostsYAMLPath())
}

func TestCacheDir_DefaultAndOverride(t *testing.T) {
	t.Setenv("COATI_CACHE_DIR", "")
	assert.Equal(t, "/var/cache/coati", CacheDir())
	assert.Equal(t, "/var/cache/coati/gists", GistCacheDir())

	tmp := t.TempDir()
	t.Setenv("COATI_CACHE_DIR", tmp)
	assert.Equal(t, tmp, CacheDir())
	assert.Equal(t, filepath.Join(tmp, "gists"), GistCacheDir())
}

func TestSSHConfigPath(t *testing.T) {
	assert.Equal(t, "/home/alice/.ssh/config", SSHConfigPath("/home/alice"))
}

func TestEffectiveUsername_SUDO_USER(t *testing.T) {
	t.Setenv("SUDO_USER", "alice")
	name, err := EffectiveUsername()
	require.NoError(t, err)
	assert.Equal(t, "alice", name)
}

func TestResolveTargetUsers_DefaultAndList(t *testing.T) {
	t.Setenv("SUDO_USER", "alice")
	users, err := ResolveTargetUsers(nil)
	require.NoError(t, err)
	assert.Equal(t, []string{"alice"}, users)

	users, err = ResolveTargetUsers([]string{"bob", " root ", "bob", ""})
	require.NoError(t, err)
	assert.Equal(t, []string{"bob", "root"}, users)
}

func TestResolveTargetUsers_EmptyList(t *testing.T) {
	_, err := ResolveTargetUsers([]string{"", "  "})
	require.Error(t, err)
}

func TestExpandHome(t *testing.T) {
	assert.Equal(t, "/home/alice/.ssh/config", ExpandHome("~/.ssh/config", "/home/alice"))
	assert.Equal(t, "/etc/hosts", ExpandHome("/etc/hosts", "/home/alice"))
}

func TestHomeFor_CurrentUser(t *testing.T) {
	t.Setenv("SUDO_USER", "")
	u, err := os.UserHomeDir()
	require.NoError(t, err)
	name, err := EffectiveUsername()
	require.NoError(t, err)
	home, err := HomeFor(name)
	require.NoError(t, err)
	assert.Equal(t, u, home)
}

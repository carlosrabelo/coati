package app

import (
	"log/slog"
	"os"
	"os/user"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"coati/coati/internal/core/services"
)

func TestNormalizeScope(t *testing.T) {
	s, err := normalizeScope("all")
	require.NoError(t, err)
	assert.Equal(t, ScopeAll, s)

	s, err = normalizeScope("system")
	require.NoError(t, err)
	assert.Equal(t, ScopeSystem, s)

	_, err = normalizeScope("nope")
	require.Error(t, err)
}

func TestNormalizeScope_DefaultByPrivilege(t *testing.T) {
	s, err := normalizeScope("")
	require.NoError(t, err)
	if services.RunningAsRoot() {
		assert.Equal(t, ScopeAll, s)
	} else {
		assert.Equal(t, ScopeUser, s)
	}
}

func TestResolveEffectiveScope_NonRoot(t *testing.T) {
	if services.RunningAsRoot() {
		t.Skip("running as root")
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	app, err := New(logger, Config{
		OutputHostsFile: "/etc/hosts",
		Scope:           ScopeUser,
	})
	require.NoError(t, err)

	_, err = app.resolveEffectiveScope(ScopeAll)
	require.Error(t, err)

	_, err = app.resolveEffectiveScope(ScopeSystem)
	require.Error(t, err)

	scope, err := app.resolveEffectiveScope(ScopeUser)
	require.NoError(t, err)
	assert.Equal(t, ScopeUser, scope)

	// Custom output path keeps requested scope (dev/tests).
	app.cfg.OutputHostsFile = filepath.Join(t.TempDir(), "hosts")
	scope, err = app.resolveEffectiveScope(ScopeAll)
	require.NoError(t, err)
	assert.Equal(t, ScopeAll, scope)
}

func TestApplication_ScopeUser_WritesOnlySSH(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("COATI_CONFIG_DIR", filepath.Join(tmp, "etc"))
	t.Setenv("COATI_DATA_DIR", filepath.Join(tmp, "lib"))
	t.Setenv("COATI_CACHE_DIR", filepath.Join(tmp, "cache"))
	t.Setenv("SUDO_USER", "")

	hostsList := filepath.Join(tmp, "hosts.yaml")
	require.NoError(t, os.WriteFile(hostsList, []byte(`
hosts:
  - ip: 10.0.0.10
    hostname: scope-host
`), 0644))

	outHosts := filepath.Join(tmp, "etc-hosts")
	outSSH := filepath.Join(tmp, "ssh-config")

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	cfg := Config{
		HostsListFile:    hostsList,
		OutputHostsFile:  outHosts,
		OutputConfigFile: outSSH,
		Scope:            ScopeUser,
		Backup:           false,
	}

	application, err := New(logger, cfg)
	require.NoError(t, err)
	require.NoError(t, application.Run())

	assert.NoFileExists(t, outHosts)
	require.FileExists(t, outSSH)
	data, err := os.ReadFile(outSSH)
	require.NoError(t, err)
	assert.Contains(t, string(data), "scope-host")
}

func TestApplication_ScopeSystem_WritesOnlyHosts(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("COATI_CONFIG_DIR", filepath.Join(tmp, "etc"))
	t.Setenv("COATI_DATA_DIR", filepath.Join(tmp, "lib"))
	t.Setenv("COATI_CACHE_DIR", filepath.Join(tmp, "cache"))

	hostsList := filepath.Join(tmp, "hosts.yaml")
	require.NoError(t, os.WriteFile(hostsList, []byte(`
hosts:
  - ip: 10.0.0.10
    hostname: scope-host
`), 0644))

	outHosts := filepath.Join(tmp, "etc-hosts")
	outSSH := filepath.Join(tmp, "ssh-config")

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	cfg := Config{
		HostsListFile:    hostsList,
		OutputHostsFile:  outHosts,
		OutputConfigFile: outSSH,
		Scope:            ScopeSystem,
		Backup:           false,
	}

	application, err := New(logger, cfg)
	require.NoError(t, err)
	require.NoError(t, application.Run())

	require.FileExists(t, outHosts)
	assert.NoFileExists(t, outSSH)
	data, err := os.ReadFile(outHosts)
	require.NoError(t, err)
	assert.Contains(t, string(data), "scope-host")
}

func TestApplication_ScopeAll_WritesBoth(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("COATI_CONFIG_DIR", filepath.Join(tmp, "etc"))
	t.Setenv("COATI_DATA_DIR", filepath.Join(tmp, "lib"))
	t.Setenv("COATI_CACHE_DIR", filepath.Join(tmp, "cache"))

	hostsList := filepath.Join(tmp, "hosts.yaml")
	require.NoError(t, os.WriteFile(hostsList, []byte(`
hosts:
  - ip: 10.0.0.10
    hostname: scope-host
`), 0644))

	outHosts := filepath.Join(tmp, "etc-hosts")
	outSSH := filepath.Join(tmp, "ssh-config")

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	cfg := Config{
		HostsListFile:    hostsList,
		OutputHostsFile:  outHosts,
		OutputConfigFile: outSSH,
		Scope:            ScopeAll,
		Backup:           false,
	}

	application, err := New(logger, cfg)
	require.NoError(t, err)
	require.NoError(t, application.Run())

	require.FileExists(t, outHosts)
	require.FileExists(t, outSSH)
}

func TestApplication_AutoResolveSSHPath(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("COATI_CONFIG_DIR", filepath.Join(tmp, "etc"))
	t.Setenv("COATI_DATA_DIR", filepath.Join(tmp, "lib"))
	t.Setenv("COATI_CACHE_DIR", filepath.Join(tmp, "cache"))

	current, err := user.Current()
	require.NoError(t, err)
	t.Setenv("SUDO_USER", current.Username)

	// Point home resolution at a fake home by using OutputConfigFile empty and
	// writing into the real user's path would be invasive. Instead verify
	// ResolveTargetUsers + SSHConfigPath wiring via resolveSSHTargets.
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	cfg := Config{Scope: ScopeUser, Backup: false}
	application, err := New(logger, cfg)
	require.NoError(t, err)

	users, err := services.ResolveTargetUsers(nil)
	require.NoError(t, err)
	targets, err := application.resolveSSHTargets(users)
	require.NoError(t, err)
	require.Len(t, targets, 1)
	assert.Equal(t, current.Username, targets[0].username)
	assert.Equal(t, services.SSHConfigPath(current.HomeDir), targets[0].path)
}

func TestIsPermissionError(t *testing.T) {
	assert.True(t, isPermissionError(os.ErrPermission))
	assert.False(t, isPermissionError(os.ErrNotExist))
}

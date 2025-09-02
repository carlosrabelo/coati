package services

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

const (
	defaultConfigDir = "/etc/coati"
	defaultDataDir   = "/var/lib/coati"
	defaultCacheDir  = "/var/cache/coati"
	defaultHostsFile = "/etc/hosts"
)

// RunningAsRoot reports whether the process effective UID is 0.
func RunningAsRoot() bool {
	return os.Geteuid() == 0
}

// ConfigDir returns the application config directory (/etc/coati).
// Override with COATI_CONFIG_DIR.
func ConfigDir() string {
	if dir := os.Getenv("COATI_CONFIG_DIR"); dir != "" {
		return dir
	}
	return defaultConfigDir
}

// AppConfigPath returns the path to config.yaml.
func AppConfigPath() string {
	return filepath.Join(ConfigDir(), "config.yaml")
}

// DataDir returns the persistent data directory (/var/lib/coati).
// Override with COATI_DATA_DIR.
func DataDir() string {
	if dir := os.Getenv("COATI_DATA_DIR"); dir != "" {
		return dir
	}
	return defaultDataDir
}

// HostsYAMLPath returns the canonical local hosts YAML copy.
func HostsYAMLPath() string {
	return filepath.Join(DataDir(), "hosts.yaml")
}

// CacheDir returns the regenerable cache directory (/var/cache/coati).
// Override with COATI_CACHE_DIR.
func CacheDir() string {
	if dir := os.Getenv("COATI_CACHE_DIR"); dir != "" {
		return dir
	}
	return defaultCacheDir
}

// GistCacheDir returns the Gist HTTP cache directory.
func GistCacheDir() string {
	return filepath.Join(CacheDir(), "gists")
}

// DefaultHostsOutput returns the default /etc/hosts output path.
func DefaultHostsOutput() string {
	return defaultHostsFile
}

// SSHConfigPath returns ~/.ssh/config for the given home directory.
func SSHConfigPath(home string) string {
	return filepath.Join(home, ".ssh", "config")
}

// EffectiveUsername returns SUDO_USER when set (and non-empty), otherwise
// the current process user.
func EffectiveUsername() (string, error) {
	if sudoUser := strings.TrimSpace(os.Getenv("SUDO_USER")); sudoUser != "" {
		return sudoUser, nil
	}
	u, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("failed to resolve current user: %w", err)
	}
	return u.Username, nil
}

// HomeFor looks up the home directory for a username.
func HomeFor(username string) (string, error) {
	u, err := user.Lookup(username)
	if err != nil {
		return "", fmt.Errorf("failed to lookup user %q: %w", username, err)
	}
	if u.HomeDir == "" {
		return "", fmt.Errorf("user %q has empty home directory", username)
	}
	return u.HomeDir, nil
}

// ResolveTargetUsers returns the list of usernames whose SSH config should be
// updated. When users is empty, returns the effective user (SUDO_USER-aware).
func ResolveTargetUsers(users []string) ([]string, error) {
	if len(users) == 0 {
		name, err := EffectiveUsername()
		if err != nil {
			return nil, err
		}
		return []string{name}, nil
	}
	out := make([]string, 0, len(users))
	seen := make(map[string]bool, len(users))
	for _, u := range users {
		u = strings.TrimSpace(u)
		if u == "" || seen[u] {
			continue
		}
		seen[u] = true
		out = append(out, u)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no valid usernames provided")
	}
	return out, nil
}

// ExpandHome expands a leading ~/ using the given home directory.
func ExpandHome(path, home string) string {
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(home, path[2:])
	}
	return path
}

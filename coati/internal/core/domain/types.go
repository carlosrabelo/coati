package domain

import (
	"fmt"
	"net"
)

type GlobalConfig struct {
	Defaults            SSHDefaults   `yaml:"defaults"`
	Hosts               []HostConfig  `yaml:"hosts"`
	CNAMEs              []CNAMEConfig `yaml:"cnames"`
	PTRs                []PTRConfig   `yaml:"ptrs"`
	SimplifiedModeHosts []string      `yaml:"simplified_mode_hosts"`
	PostHooks           []string      `yaml:"post_hooks"`
}

type AppConfig struct {
	GistID      string `yaml:"gist_id"`
	GitHubToken string `yaml:"github_token"`
}

type CNAMEConfig struct {
	Target  string   `yaml:"target"`
	Aliases []string `yaml:"alias"`
}

type PTRConfig struct {
	IP     string `yaml:"ip"`
	Domain string `yaml:"domain"`
}

type HostConfig struct {
	// Common / Hosts file
	Hostname string   `yaml:"hostname"`
	IP       string   `yaml:"ip"`
	Aliases  []string `yaml:"aliases,omitempty"`
	DNSName  string   `yaml:"dns_name,omitempty"`
	Comment  string   `yaml:"comment,omitempty"`
	// SSH specific
	User         string            `yaml:"user,omitempty"`
	Port         int               `yaml:"port,omitempty"`
	IdentityFile string            `yaml:"identity_file,omitempty"`
	Options      map[string]string `yaml:"options,omitempty"`
}

type SSHDefaults struct {
	User         string            `yaml:"user,omitempty"`
	Port         int               `yaml:"port,omitempty"`
	IdentityFile string            `yaml:"identity_file,omitempty"`
	Options      map[string]string `yaml:"options,omitempty"`
}

// HostEntry represents a single line parsed from an existing /etc/hosts file.
type HostEntry struct {
	IP       string
	Hostname string
	Aliases  []string
	Comment  string
}

// SSHConfig represents a Host block parsed from an existing ~/.ssh/config file.
type SSHConfig struct {
	Host         string
	HostName     string
	User         string
	Port         int
	IdentityFile string
	Options      map[string]string
}

func (c *GlobalConfig) Validate() error {
	v := NewValidator()
	for i, h := range c.Hosts {
		// The ip field accepts a valid IP address (for /etc/hosts generation) or a hostname
		// (for SSH-only entries, e.g. "github.com", which are skipped in /etc/hosts).
		if errIP := v.ValidateIP(h.IP); errIP != nil {
			if errHost := v.ValidateHostname(h.IP); errHost != nil {
				return fmt.Errorf("host[%d]: ip field %q is neither a valid IP address nor a valid hostname", i, h.IP)
			}
		}
		if err := v.ValidateHostname(h.Hostname); err != nil {
			return fmt.Errorf("host[%d]: invalid hostname %q: %w", i, h.Hostname, err)
		}
		for _, alias := range h.Aliases {
			if err := v.ValidateHostname(alias); err != nil {
				return fmt.Errorf("host[%d]: invalid alias %q: %w", i, alias, err)
			}
		}
	}
	seenHostnames := make(map[string]bool)
	seenIPs := make(map[string]bool)
	for i, h := range c.Hosts {
		if seenHostnames[h.Hostname] {
			return fmt.Errorf("host[%d]: duplicate hostname %q", i, h.Hostname)
		}
		seenHostnames[h.Hostname] = true
		if net.ParseIP(h.IP) != nil {
			if seenIPs[h.IP] {
				return fmt.Errorf("host[%d]: duplicate ip %q", i, h.IP)
			}
			seenIPs[h.IP] = true
		}
	}
	if c.Defaults.Port != 0 && (c.Defaults.Port < 1 || c.Defaults.Port > 65535) {
		return fmt.Errorf("invalid default port: %d", c.Defaults.Port)
	}
	return nil
}

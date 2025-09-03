package services

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"coati/coati/internal/core/domain"
)

type SSHGenerator struct {
	config domain.GlobalConfig
}

func NewSSHGenerator(config domain.GlobalConfig) *SSHGenerator {
	return &SSHGenerator{config: config}
}

func (g *SSHGenerator) GenerateSSHConfig() ([]byte, error) {
	var buf bytes.Buffer

	seenHosts := make(map[string]bool)

	// Sort a copy of hosts by IP (string order), then by Hostname for ties.
	hosts := make([]domain.HostConfig, len(g.config.Hosts))
	copy(hosts, g.config.Hosts)
	sort.Slice(hosts, func(i, j int) bool {
		if hosts[i].IP != hosts[j].IP {
			return hosts[i].IP < hosts[j].IP
		}
		return hosts[i].Hostname < hosts[j].Hostname
	})

	if len(hosts) > 0 {
		buf.WriteString("# === Source: Configuration ===\n")
	}

	for _, h := range hosts {
		if seenHosts[h.Hostname] {
			continue
		}

		patterns := []string{h.Hostname}
		patterns = append(patterns, h.Aliases...)

		config := domain.SSHConfig{
			Host:         strings.Join(patterns, " "),
			HostName:     h.IP,
			User:         h.User,
			Port:         h.Port,
			IdentityFile: h.IdentityFile,
			Options:      h.Options,
		}

		writeSSHConfig(&buf, config)
		seenHosts[h.Hostname] = true
	}

	// Host * block — only emitted when at least one default is set
	defaults := g.config.Defaults
	if defaults.User != "" || defaults.Port != 0 || defaults.IdentityFile != "" || len(defaults.Options) > 0 {
		buf.WriteString("# === Source: Defaults ===\n")
		buf.WriteString("Host *\n")
		if defaults.User != "" {
			buf.WriteString(fmt.Sprintf("    User %s\n", defaults.User))
		}
		if defaults.Port != 0 {
			buf.WriteString(fmt.Sprintf("    Port %d\n", defaults.Port))
		}
		if defaults.IdentityFile != "" {
			buf.WriteString(fmt.Sprintf("    IdentityFile %s\n", defaults.IdentityFile))
		}
		for _, k := range sortedKeys(defaults.Options) {
			buf.WriteString(fmt.Sprintf("    %s %s\n", k, defaults.Options[k]))
		}
	}

	out := make([]byte, buf.Len())
	copy(out, buf.Bytes())
	return out, nil
}

func writeSSHConfig(buf *bytes.Buffer, c domain.SSHConfig) {
	buf.WriteString(fmt.Sprintf("Host %s\n", c.Host))
	if c.HostName != "" {
		buf.WriteString(fmt.Sprintf("    HostName %s\n", c.HostName))
	}
	if c.User != "" {
		buf.WriteString(fmt.Sprintf("    User %s\n", c.User))
	}
	if c.Port != 0 {
		buf.WriteString(fmt.Sprintf("    Port %d\n", c.Port))
	}
	if c.IdentityFile != "" {
		buf.WriteString(fmt.Sprintf("    IdentityFile %s\n", c.IdentityFile))
	}
	for _, k := range sortedKeys(c.Options) {
		buf.WriteString(fmt.Sprintf("    %s %s\n", k, c.Options[k]))
	}
	buf.WriteString("\n")
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

package services

import (
	"net"
	"strings"

	"coati/coati/internal/core/domain"
	"coati/coati/internal/core/ports"
)

type Importer struct {
	hostsParser ports.HostsParser
	sshParser   ports.SSHParser
}

func NewImporter(hostsParser ports.HostsParser, sshParser ports.SSHParser) *Importer {
	return &Importer{
		hostsParser: hostsParser,
		sshParser:   sshParser,
	}
}

func (im *Importer) Import(hostsContent []byte, sshContent []byte) (domain.GlobalConfig, error) {
	var config domain.GlobalConfig

	var hostEntries []domain.HostEntry
	var err error
	if len(hostsContent) > 0 {
		hostEntries, err = im.hostsParser.ParseHosts(hostsContent)
		if err != nil {
			return config, err
		}
	}

	var sshConfigs []domain.SSHConfig
	if len(sshContent) > 0 {
		sshConfigs, err = im.sshParser.ParseSSHConfig(sshContent)
		if err != nil {
			return config, err
		}
	}

	// We'll keep track of hosts we've built.
	// Map from IP to index in config.Hosts to quickly find IP-based hosts
	ipMap := make(map[string]int)
	// Map from Hostname/Alias to index in config.Hosts to quickly find hostname-based hosts
	nameMap := make(map[string]int)

	// Helper to add a host and update maps
	addHost := func(h domain.HostConfig) {
		idx := len(config.Hosts)
		config.Hosts = append(config.Hosts, h)
		if h.IP != "" && net.ParseIP(h.IP) != nil {
			ipMap[h.IP] = idx
		}
		if h.Hostname != "" {
			nameMap[strings.ToLower(h.Hostname)] = idx
		}
		for _, alias := range h.Aliases {
			nameMap[strings.ToLower(alias)] = idx
		}
	}

	// 1. Process Hosts File Entries
	for _, entry := range hostEntries {
		// Skip standard loopback mappings to avoid pollution and validation failures
		if isStandardLoopback(entry.IP, entry.Hostname) {
			continue
		}

		// Check if Hostname is already in our list
		if idx, exists := nameMap[strings.ToLower(entry.Hostname)]; exists {
			// Merge aliases and comments into existing host config
			existing := &config.Hosts[idx]
			if existing.IP == "" {
				existing.IP = entry.IP
				ipMap[entry.IP] = idx
			}
			if entry.Comment != "" {
				if existing.Comment != "" {
					existing.Comment += "; " + entry.Comment
				} else {
					existing.Comment = entry.Comment
				}
			}
			for _, alias := range entry.Aliases {
				if !containsCaseInsensitive(existing.Aliases, alias) && strings.ToLower(alias) != strings.ToLower(existing.Hostname) {
					existing.Aliases = append(existing.Aliases, alias)
					nameMap[strings.ToLower(alias)] = idx
				}
			}
			continue
		}

		// Create a new HostConfig
		h := domain.HostConfig{
			IP:       entry.IP,
			Hostname: entry.Hostname,
			Aliases:  entry.Aliases,
			Comment:  entry.Comment,
		}
		addHost(h)
	}

	// 2. Process SSH Config Entries
	for _, sc := range sshConfigs {
		// Clean up host patterns
		patterns := strings.Fields(sc.Host)
		if len(patterns) == 0 {
			continue
		}

		// Skip wildcards or global configs (Host *)
		if len(patterns) == 1 && patterns[0] == "*" {
			// Map global settings to Defaults
			if sc.User != "" {
				config.Defaults.User = sc.User
			}
			if sc.Port != 0 {
				config.Defaults.Port = sc.Port
			}
			if sc.IdentityFile != "" {
				config.Defaults.IdentityFile = sc.IdentityFile
			}
			if len(sc.Options) > 0 {
				if config.Defaults.Options == nil {
					config.Defaults.Options = make(map[string]string)
				}
				for k, v := range sc.Options {
					config.Defaults.Options[k] = v
				}
			}
			continue
		}

		// Try to match this SSH config to an existing HostConfig by host patterns
		var targetIdx = -1

		for _, pat := range patterns {
			if idx, ok := nameMap[strings.ToLower(pat)]; ok {
				targetIdx = idx
				break
			}
		}

		if targetIdx != -1 {
			// Merge SSH config into existing HostConfig
			h := &config.Hosts[targetIdx]
			if h.User == "" {
				h.User = sc.User
			}
			if h.Port == 0 {
				h.Port = sc.Port
			}
			if h.IdentityFile == "" {
				h.IdentityFile = sc.IdentityFile
			}
			if len(sc.Options) > 0 {
				if h.Options == nil {
					h.Options = make(map[string]string)
				}
				for k, v := range sc.Options {
					if _, ok := h.Options[k]; !ok {
						h.Options[k] = v
					}
				}
			}
			// Add any missing host patterns as aliases
			for _, pat := range patterns {
				patLower := strings.ToLower(pat)
				if patLower != strings.ToLower(h.Hostname) && !containsCaseInsensitive(h.Aliases, pat) {
					h.Aliases = append(h.Aliases, pat)
					nameMap[patLower] = targetIdx
				}
			}
		} else {
			// Create a new HostConfig for this SSH block
			hostname := patterns[0]
			var aliases []string
			if len(patterns) > 1 {
				aliases = patterns[1:]
			}

			ipVal := sc.HostName
			if ipVal == "" {
				// Fallback to hostname if no HostName parameter is set (e.g. Host name is the target)
				ipVal = hostname
			}

			h := domain.HostConfig{
				Hostname:     hostname,
				IP:           ipVal,
				Aliases:      aliases,
				User:         sc.User,
				Port:         sc.Port,
				IdentityFile: sc.IdentityFile,
				Options:      sc.Options,
			}
			addHost(h)
		}
	}

	return config, nil
}

func isStandardLoopback(ip, hostname string) bool {
	ip = strings.TrimSpace(ip)
	hostname = strings.ToLower(strings.TrimSpace(hostname))

	if ip == "127.0.0.1" && (hostname == "localhost" || hostname == "localhost.localdomain") {
		return true
	}
	if ip == "::1" && (hostname == "localhost" || hostname == "ip6-localhost" || hostname == "ip6-loopback") {
		return true
	}
	// Other standard IPv6 loopback / multicast lines
	if ip == "fe00::0" || ip == "ff00::0" || ip == "ff02::1" || ip == "ff02::2" {
		return true
	}
	return false
}

func containsCaseInsensitive(slice []string, val string) bool {
	for _, item := range slice {
		if strings.EqualFold(item, val) {
			return true
		}
	}
	return false
}

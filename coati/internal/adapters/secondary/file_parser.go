package secondary

// Package secondary provides filesystem and HTTP adapter implementations.

import (
	"bufio"
	"bytes"
	"strconv"
	"strings"

	"coati/coati/internal/core/domain"
)

type HostFileParser struct{}

func NewHostFileParser() *HostFileParser {
	return &HostFileParser{}
}

func (p *HostFileParser) ParseHosts(content []byte) ([]domain.HostEntry, error) {
	var entries []domain.HostEntry
	scanner := bufio.NewScanner(bytes.NewReader(content))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		var comment string
		if idx := strings.Index(line, "#"); idx != -1 {
			comment = strings.TrimSpace(line[idx+1:])
			line = strings.TrimSpace(line[:idx])
		}

		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		entry := domain.HostEntry{
			IP:       parts[0],
			Hostname: parts[1],
			Comment:  comment,
		}

		if len(parts) > 2 {
			entry.Aliases = parts[2:]
		}

		entries = append(entries, entry)
	}

	return entries, scanner.Err()
}

type SSHFileParser struct{}

func NewSSHFileParser() *SSHFileParser {
	return &SSHFileParser{}
}

func (p *SSHFileParser) ParseSSHConfig(content []byte) ([]domain.SSHConfig, error) {
	var configs []domain.SSHConfig
	var current *domain.SSHConfig

	scanner := bufio.NewScanner(bytes.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		var originalKey string
		var key, value string
		if idx := strings.Index(line, "="); idx != -1 {
			firstWord := strings.Fields(line[:idx])
			if len(firstWord) == 1 {
				originalKey = firstWord[0]
				key = strings.ToLower(originalKey)
				value = strings.TrimSpace(line[idx+1:])
			}
		}
		if key == "" {
			parts := strings.Fields(line)
			if len(parts) < 2 {
				continue
			}
			originalKey = parts[0]
			key = strings.ToLower(originalKey)
			value = strings.Join(parts[1:], " ")
		}

		if value == "" {
			continue
		}

		// Trim surrounding quotes if present
		if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
			(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
			value = value[1 : len(value)-1]
		}

		if key == "host" {
			if current != nil {
				configs = append(configs, *current)
			}
			current = &domain.SSHConfig{
				Host:    value,
				Options: make(map[string]string),
			}
			continue
		}

		if current == nil {
			continue
		}

		switch key {
		case "hostname":
			current.HostName = value
		case "user":
			current.User = value
		case "port":
			if p, err := strconv.Atoi(value); err == nil {
				current.Port = p
			}
		case "identityfile":
			current.IdentityFile = value
		default:
			current.Options[originalKey] = value
		}
	}

	if current != nil {
		configs = append(configs, *current)
	}

	return configs, scanner.Err()
}

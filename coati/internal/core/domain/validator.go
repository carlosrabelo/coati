package domain

import (
	"fmt"
	"net"
	"regexp"
	"strings"
)

var (
	// Regex for valid hostname (RFC 1123)
	hostnameRegex = regexp.MustCompile(`^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])$`)
	// OpenSSH Host patterns may include * and ? in labels (e.g. 10.13.50.*, *.lab.local).
	hostPatternRegex = regexp.MustCompile(`^([a-zA-Z0-9*?]|[a-zA-Z0-9*?][a-zA-Z0-9\-.*?]*[a-zA-Z0-9*?])(\.([a-zA-Z0-9*?]|[a-zA-Z0-9*?][a-zA-Z0-9\-.*?]*[a-zA-Z0-9*?]))*$`)
)

type Validator struct{}

func NewValidator() *Validator {
	return &Validator{}
}

// IsSSHHostPattern reports whether name is an OpenSSH Host pattern (* or ?).
func IsSSHHostPattern(name string) bool {
	return strings.ContainsAny(name, "*?")
}

// ValidateIP checks if the string is a valid IPv4 or IPv6 address.
func (v *Validator) ValidateIP(ip string) error {
	if ip == "" {
		return fmt.Errorf("ip cannot be empty")
	}
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return fmt.Errorf("invalid ip address: %s", ip)
	}
	return nil
}

// ValidateHostname checks if the string is a valid hostname or SSH Host pattern.
func (v *Validator) ValidateHostname(hostname string) error {
	if hostname == "" {
		return fmt.Errorf("hostname cannot be empty")
	}
	if len(hostname) > 255 {
		return fmt.Errorf("hostname too long: %s", hostname)
	}
	if strings.HasPrefix(hostname, "-") || strings.HasSuffix(hostname, "-") {
		return fmt.Errorf("hostname cannot start or end with hyphen: %s", hostname)
	}
	if IsSSHHostPattern(hostname) {
		if !hostPatternRegex.MatchString(hostname) {
			return fmt.Errorf("invalid hostname format: %s", hostname)
		}
		return nil
	}
	if !hostnameRegex.MatchString(hostname) {
		return fmt.Errorf("invalid hostname format: %s", hostname)
	}
	return nil
}

// ValidateHostEntry validates a single host entry from config.
func (v *Validator) ValidateHostEntry(h HostConfig) error {
	if IsSSHHostPattern(h.Hostname) {
		if strings.TrimSpace(h.IP) != "" {
			return fmt.Errorf("SSH host pattern %q requires empty ip", h.Hostname)
		}
	} else if err := v.ValidateIP(h.IP); err != nil {
		return err
	}
	if err := v.ValidateHostname(h.Hostname); err != nil {
		return err
	}
	for _, alias := range h.Aliases {
		if err := v.ValidateHostname(alias); err != nil {
			return fmt.Errorf("invalid alias '%s' for host '%s': %w", alias, h.Hostname, err)
		}
	}
	return nil
}

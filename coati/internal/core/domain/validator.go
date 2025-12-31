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
)

type Validator struct{}

func NewValidator() *Validator {
	return &Validator{}
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

// ValidateHostname checks if the string is a valid hostname.
func (v *Validator) ValidateHostname(hostname string) error {
	if hostname == "" {
		return fmt.Errorf("hostname cannot be empty")
	}
	if len(hostname) > 255 {
		return fmt.Errorf("hostname too long: %s", hostname)
	}
	if !hostnameRegex.MatchString(hostname) {
		return fmt.Errorf("invalid hostname format: %s", hostname)
	}
	if strings.HasPrefix(hostname, "-") || strings.HasSuffix(hostname, "-") {
		return fmt.Errorf("hostname cannot start or end with hyphen: %s", hostname)
	}
	return nil
}

// ValidateHostEntry validates a single host entry from config.
func (v *Validator) ValidateHostEntry(h HostConfig) error {
	if err := v.ValidateIP(h.IP); err != nil {
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

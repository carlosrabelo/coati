package services

// Package services provides business logic for generating hosts and SSH configurations.

import (
	"bytes"
	"fmt"
	"net"
	"sort"
	"strings"
	"text/tabwriter"

	"coati/internal/core/domain"
	"coati/internal/core/ports"
)

type HostsGenerator struct {
	config               domain.GlobalConfig
	fileReader           ports.FileReader
	hostsTemplateContent string
}

func NewHostsGenerator(
	config domain.GlobalConfig,
	reader ports.FileReader,
	hostsTemplateContent string,
) *HostsGenerator {
	return &HostsGenerator{
		config:               config,
		fileReader:           reader,
		hostsTemplateContent: hostsTemplateContent,
	}
}

func (g *HostsGenerator) GenerateHosts() ([]byte, error) {
	hostname := g.readHostname()

	var buf bytes.Buffer
	// use tabwidth 8, padding 1, padchar '\t', flags 0 to align using tabs
	// minwidth 24 ensures short IPs (::1) get 3 tabs (to reach 24)
	// while normal IPs (127.0.0.1) get 2 tabs, aligning columns.
	w := tabwriter.NewWriter(&buf, 24, 8, 1, '\t', 0)

	lastLineWasBlank := g.writeTemplate(w, hostname)

	seen := map[string]bool{
		"127.0.0.1": true, "127.0.1.1": true, "::1": true,
		"fe00::0": true, "ff00::0": true, "ff02::1": true, "ff02::2": true,
	}

	// Sort a copy of hosts by IP (network byte order) to avoid mutating the shared config slice.
	hosts := make([]domain.HostConfig, len(g.config.Hosts))
	copy(hosts, g.config.Hosts)
	sort.Slice(hosts, func(i, j int) bool {
		ip1 := net.ParseIP(hosts[i].IP)
		ip2 := net.ParseIP(hosts[j].IP)
		return bytes.Compare(ip1, ip2) < 0
	})

	if len(hosts) > 0 {
		if !lastLineWasBlank {
			fmt.Fprintln(w, "")
		}
		fmt.Fprintln(w, "# === Source: Configuration ===")
	}

	// Build PTR map
	ptrMap := make(map[string]string)
	for _, p := range g.config.PTRs {
		ptrMap[p.IP] = p.Domain
	}

	// Check if we are in Simplified Mode
	simpleMode := false
	for _, h := range g.config.SimplifiedModeHosts {
		if h == hostname {
			simpleMode = true
			break
		}
	}

	for _, h := range hosts {
		// Skip if hostname matches system hostname (already added by template)
		if h.Hostname == hostname {
			continue
		}
		// Only add to /etc/hosts if it is a valid IP
		if net.ParseIP(h.IP) == nil || seen[h.IP] {
			continue
		}
		writeHostEntry(w, h.IP, h.Hostname, g.buildAliases(h, simpleMode, ptrMap), h.Comment)
		seen[h.IP] = true
	}

	w.Flush()

	out := make([]byte, buf.Len())
	copy(out, buf.Bytes())
	return out, nil
}

// readHostname reads /etc/hostname and returns the trimmed value.
// Returns "localhost" if the file is missing or unreadable.
func (g *HostsGenerator) readHostname() string {
	b, _ := g.fileReader.ReadFile("/etc/hostname")
	if h := strings.TrimSpace(string(b)); h != "" {
		return h
	}
	return "localhost"
}

// writeTemplate writes the template section to w and returns whether the last
// written line was blank (used to decide if a separator is needed before the
// configuration section).
func (g *HostsGenerator) writeTemplate(w *tabwriter.Writer, hostname string) bool {
	if g.hostsTemplateContent == "" {
		// Fallback to Ubuntu defaults if no template
		fmt.Fprintln(w, "# === Source: Template (Default) ===")
		fmt.Fprintln(w, "127.0.0.1\tlocalhost")
		fmt.Fprintf(w, "127.0.1.1\t%s\n", hostname)
		fmt.Fprintln(w, "\n# The following lines are desirable for IPv6 capable hosts")
		fmt.Fprintln(w, "::1\tip6-localhost\tip6-loopback")
		fmt.Fprintln(w, "fe00::0\tip6-localnet")
		fmt.Fprintln(w, "ff00::0\tip6-mcastprefix")
		fmt.Fprintln(w, "ff02::1\tip6-allnodes")
		fmt.Fprintln(w, "ff02::2\tip6-allrouters")
		return false
	}

	fmt.Fprintln(w, "# === Source: Template ===")
	tplContent := strings.ReplaceAll(g.hostsTemplateContent, "<hostname>", hostname)

	lastLineWasBlank := false
	for _, line := range strings.Split(tplContent, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			if lastLineWasBlank {
				continue
			}
			lastLineWasBlank = true
			fmt.Fprintln(w, "")
			continue
		}
		lastLineWasBlank = false

		if strings.HasPrefix(trimmed, "#") {
			fmt.Fprintln(w, line)
			continue
		}
		// Attempt to align host lines: Tab between IP and Host, space for the rest
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			aligned := fmt.Sprintf("%s\t%s", parts[0], parts[1])
			if len(parts) > 2 {
				aligned += " " + strings.Join(parts[2:], " ")
			}
			fmt.Fprintln(w, aligned)
		} else {
			fmt.Fprintln(w, line)
		}
	}
	return lastLineWasBlank
}

// buildAliases assembles the deduplicated alias list for a host entry:
// PTR/DNSName first, then explicit aliases, then CNAME expansions (all
// skipped in simplified mode). The canonical hostname is never included.
func (g *HostsGenerator) buildAliases(h domain.HostConfig, simpleMode bool, ptrMap map[string]string) []string {
	var all []string

	if !simpleMode {
		if ptrDomain, ok := ptrMap[h.IP]; ok {
			all = append(all, ptrDomain)
		} else if h.DNSName != "" {
			all = append(all, h.DNSName)
		}
	}

	all = append(all, h.Aliases...)

	if !simpleMode {
		for _, c := range g.config.CNAMEs {
			if strings.EqualFold(c.Target, h.Hostname) || strings.EqualFold(c.Target, h.DNSName) {
				all = append(all, c.Aliases...)
			}
		}
	}

	// Deduplicate, excluding the canonical hostname
	seen := map[string]bool{h.Hostname: true}
	result := all[:0:0]
	for _, alias := range all {
		if !seen[alias] {
			result = append(result, alias)
			seen[alias] = true
		}
	}
	return result
}

func writeHostEntry(w *tabwriter.Writer, ip, canonical string, aliases []string, comment string) {
	line := fmt.Sprintf("%s\t%s", ip, canonical)
	if len(aliases) > 0 {
		line += " " + strings.Join(aliases, " ")
	}
	if comment != "" {
		line += " # " + comment
	}
	fmt.Fprintln(w, line)
}

package domain

type GlobalConfig struct {
	Defaults            SSHDefaults   `yaml:"defaults"`
	Hosts               []HostConfig  `yaml:"hosts"`
	CNAMEs              []CNAMEConfig `yaml:"cnames"`
	PTRs                []PTRConfig   `yaml:"ptrs"`
	SimplifiedModeHosts []string      `yaml:"simplified_mode_hosts"`
	PostHooks           []string      `yaml:"post_hooks"`
}

type AppConfig struct {
	GistID       string   `yaml:"gist_id"`
	GitHubToken  string   `yaml:"github_token"`
	AllowedHooks []string `yaml:"allowed_hooks,omitempty"`
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

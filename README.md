# Coati

Coati is a modern CLI tool for managing local `/etc/hosts` and SSH configurations (`~/.ssh/config`). It allows you to define your infrastructure in a clean YAML format and generate the necessary system files automatically.

## Highlights

- Define hosts, aliases, and SSH options in a single YAML file
- Automatically generates `/etc/hosts` with proper formatting and alignment
- Generates `~/.ssh/config` from the same host definitions
- Fetch configuration from a private GitHub Gist for portable, shareable setups
- Caches Gist responses locally to reduce network calls
- Strict validation of IP addresses and hostnames
- Run custom commands after successful configuration generation via hooks
- Preview changes without writing to disk with dry-run mode
- Auto-completion support for bash, zsh, fish, and PowerShell

## Table of Contents

- [Highlights](#highlights)
- [Overview](#overview)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Usage](#usage)
- [Configuration](#configuration)
- [Project Layout](#project-layout)
- [Development](#development)
- [Testing](#testing)
- [Troubleshooting](#troubleshooting)
- [Contributing](#contributing)
- [License](#license)

## Overview

Coati simplifies infrastructure management by providing a single source of truth for host definitions. Instead of manually editing `/etc/hosts` and `~/.ssh/config`, you define your servers once in a YAML file and Coati handles the rest. This approach ensures consistency across your systems, reduces manual errors, and makes it easy to share configurations with your team.

## Prerequisites

- **Go 1.25+** (for building from source)
- **YAML file** (configuration file)
- **Write permissions** for `/etc/hosts` (requires sudo)
- **Write permissions** for `~/.ssh/config`

## Installation

### From Source

```bash
git clone https://github.com/carlosrabelo/coati
cd coati
make install
```

### Using Go Install

```bash
go install github.com/carlosrabelo/coati/cmd/coati@latest
```

## Quick Start

Get started in less than 2 minutes:

1. Create a configuration file:
```bash
cat > hosts.yaml << 'EOF'
hosts:
  - hostname: my-server
    ip: 192.168.1.100
EOF
```

2. Run Coati:
```bash
sudo coati --hosts-list hosts.yaml --output-hosts /etc/hosts
```

3. Verify:
```bash
cat /etc/hosts
# Output: 192.168.1.100    my-server
```

## Usage

### Basic Usage

1. Create a configuration file (e.g., `hosts.yaml`).
2. Run Coati:

```bash
sudo coati --hosts-list hosts.yaml --output-hosts /etc/hosts --output-config ~/.ssh/config
```

### Configuration Format

```yaml
defaults:
  user: ubuntu
  port: 22
  identity_file: ~/.ssh/id_rsa

hosts:
  - hostname: web-prod
    ip: 192.168.1.10
    aliases: [www, portal]
    user: admin

  - hostname: db-prod
    ip: 192.168.1.20

post_hooks:
  - "sudo systemctl restart dnsmasq"
```

### Advanced Commands

- **Dry Run**: Preview changes with colored output.
  ```bash
  coati --dry-run
  ```

- **Force Refresh**: Bypass cache and fetch from Gist.
  ```bash
  coati --force-refresh
  ```

- **Verbose Mode**: Enable debug logging.
  ```bash
  coati --verbose
  ```

- **Shell Completion**: Generate auto-completion scripts.
  ```bash
  source <(coati completion bash)
  ```

## Configuration

### Default Configuration

A default configuration is provided in `cfg/config.yaml`:

```yaml
defaults:
  user: ubuntu
  port: 22
  identity_file: ~/.ssh/id_rsa

hosts:
  - hostname: web-prod
    ip: 192.168.1.10
    aliases: [www, portal]
    user: admin

  - hostname: db-prod
    ip: 192.168.1.20

post_hooks:
  - "sudo systemctl restart dnsmasq"
```

### Environment Variables

- `GITHUB_TOKEN`: GitHub token for Gist access
- `COATI_CONFIG_DIR`: Custom configuration directory path

## Project Layout

```
coati/
├── bin/                    ← Compiled binaries
├── cfg/                    ← Configuration files
│   └── config.yaml         ← Default configuration
├── cmd/                    ← CLI entry point
│   └── coati/              ← Main application
├── internal/               ← Internal packages
│   ├── adapters/           ← Port implementations
│   │   └── secondary/      ← Outbound adapters
│   ├── core/               ← Business logic
│   │   ├── domain/         ← Domain models
│   │   ├── ports/          ← Interfaces
│   │   └── services/       ← Application services
│   └── templates/          ← Embedded templates
├── make/                   ← Automation scripts
│   ├── build.sh            ← Build project
│   ├── test.sh             ← Run tests
│   ├── clean.sh            ← Clean artifacts
│   ├── install.sh          ← Install binary
│   └── uninstall.sh        ← Remove binary
├── out/                    ← Generated output files
├── test/                   ← Integration tests
│   └── testdata/           ← Test fixtures
├── Makefile                ← Build automation
├── README.md               ← English documentation
└── README-PT.md            ← Portuguese documentation
```

## Development

```bash
make build      # Compile binary to bin/coati
make test       # Run all tests
make quality    # Format, vet, and lint
make install    # Install to ~/.local/bin
```

## Testing

### Running Tests

```bash
make test
```

Or directly:

```bash
./run/test.sh
```

### Test Coverage

```bash
go test -cover ./...
```

### Test Structure

- **Unit tests**: `**/*_test.go`
- **Integration tests**: `test/integration_test.go`
- **Test data**: `test/testdata/`

### Current Coverage

- `cmd/coati`: ~50% (basic tests for hook validation)
- `internal/adapters/secondary`: ~67%
- `internal/core/domain`: ~85%
- `internal/core/services`: ~95%
- `internal/core/ports`: 0% (interfaces only)
- `internal/templates`: 0% (embedded templates)

## Troubleshooting

### Issue: "command not found"

**Solution**: Ensure installation completed successfully:
```bash
which coati
# Should show: /usr/local/bin/coati or ~/.local/bin/coati
```

If not found, reinstall:
```bash
make install
```

### Issue: "permission denied when writing /etc/hosts"

**Solution**: Run Coati with sudo:
```bash
sudo coati --hosts-list hosts.yaml --output-hosts /etc/hosts
```

### Issue: "hook validation failed"

**Solution**: Check that the hook command is in the allowlist:
```bash
# Allowed commands: systemctl, service, docker, kubectl, nginx, apache2, httpd
# Commands cannot contain: ;, &, |
```

### Issue: "connection refused when fetching from Gist"

**Solution**: Check your GitHub token and network connection:
```bash
export GITHUB_TOKEN=your_token_here
coati --verbose
```

### Issue: "cache not expiring"

**Solution**: Force refresh to bypass cache:
```bash
coati --force-refresh
```

## Contributing

Contributions are welcome! Please follow these guidelines:

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/your-feature`
3. Make your changes
4. Write tests for new functionality
5. Ensure all tests pass: `make test`
6. Format your code: `gofmt -w .`
7. Run linters: `go vet ./...`
8. Commit your changes: `git commit -m "feat: description"`
9. Push to branch: `git push origin feature/your-feature`
10. Open a Pull Request

### Code Style

- Follow standard Go conventions
- Keep functions focused and small
- Add package documentation
- Write tests for all public functions
- Use structured logging with `slog`

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

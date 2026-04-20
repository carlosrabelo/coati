# Coati

CLI that generates `/etc/hosts` and `~/.ssh/config` from a YAML definition stored locally or fetched from a private GitHub Gist.

## Highlights

- Define hosts, aliases, and SSH options in a single YAML file
- Generates `/etc/hosts` with proper formatting and column alignment
- Generates `~/.ssh/config` from the same host definitions
- Pull config from a private GitHub Gist; push local changes back with `coati push`
- Caches Gist responses locally to reduce network calls
- Strict validation rejects duplicate hostnames and IPs before writing any file
- Merge mode preserves existing file content in named `# BEGIN ORIGINAL` sections
- Check mode shows a unified diff before any file is written
- Dry-run mode previews generated content without touching disk
- Shell completion for bash, zsh, fish, and PowerShell

## Table of Contents

- [Overview](#overview)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Usage](#usage)
- [Configuration](#configuration)
- [Project Layout](#project-layout)
- [Development](#development)
- [Contributing](#contributing)
- [License](#license)

## Overview

Coati provides a single source of truth for host definitions. Instead of manually editing `/etc/hosts` and `~/.ssh/config`, define your servers once in a YAML file and Coati handles the rest. Store the YAML in a private GitHub Gist to keep it synchronized across machines.

## Prerequisites

- **Go 1.25+** — required to build from source; [download](https://go.dev/dl/)
- **Write permissions** for `/etc/hosts` (requires sudo)
- **Write permissions** for `~/.ssh/config`

## Installation

### Build from Source

```bash
git clone https://github.com/carlosrabelo/coati
cd coati
make build
```

Install to `~/.local/bin` (no root required):

```bash
make install
```

### Using Go Install

```bash
go install github.com/carlosrabelo/coati/cmd/coati@latest
```

## Quick Start

1. Create a YAML config file:

```bash
cat > hosts.yaml << 'EOF'
hosts:
  - hostname: my-server
    ip: 192.168.1.100
EOF
```

2. Process and verify:

```bash
coati process --hosts-list hosts.yaml --output-hosts data/gen/etc/hosts
cat data/gen/etc/hosts
# 192.168.1.100    my-server
```

## Usage

### process

Generate `/etc/hosts` and `~/.ssh/config` from a YAML file:

```bash
coati process --hosts-list hosts.yaml
```

Write directly to system paths (requires sudo for `/etc/hosts`):

```bash
coati process --output-hosts /etc/hosts --output-config ~/.ssh/config
```

### pull / push

Download Gist content to `data/src/gist.txt`:

```bash
coati pull
```

Upload `data/src/gist.txt` back to the Gist:

```bash
coati push
```

Both commands read `--gist-id` and `--github-token` from flags, the `GITHUB_TOKEN` environment variable, or the saved config at `/etc/coati/config.yaml`.

### Advanced flags

- **Dry Run**: Preview generated content without writing.
  ```bash
  coati process --dry-run
  ```

- **Check**: Show a unified diff between current files and what would be written.
  ```bash
  coati process --check
  coati process --check --merge
  ```

- **Merge**: Preserve existing content in a `# BEGIN ORIGINAL` section; manage only the `# BEGIN COATI` section. Safe to run repeatedly.
  ```bash
  sudo coati process --merge --output-hosts /etc/hosts
  ```

- **Gist File**: Select a specific file when the Gist contains multiple files.
  ```bash
  coati process --gist-id abc123 --gist-file work.yaml
  ```

- **Force Refresh**: Bypass the local cache and fetch from Gist.
  ```bash
  coati process --force-refresh
  ```

- **Shell Completion**: Install auto-completion for your shell.
  ```bash
  coati completion bash
  coati completion zsh
  coati completion fish
  ```

## Configuration

### YAML format

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
  - "systemctl restart dnsmasq"
```

### Save Gist credentials

Run once to store your Gist ID and token in `/etc/coati/config.yaml`:

```bash
coati process --gist-id YOUR_GIST_ID --github-token YOUR_TOKEN --save-config
```

After saving, `coati pull`, `coati push`, and `coati process` all work without flags.

### Environment variables

- `GITHUB_TOKEN` — GitHub token for Gist access
- `COATI_CONFIG_DIR` — override the default config directory (`/etc/coati`)

## Project Layout

```
bin/                        # Compiled binaries (git-ignored)
data/
  src/gist.txt              # Local copy of the Gist (written by coati pull)
  gen/etc/hosts             # Generated hosts file (written by coati process)
  gen/ssh/config            # Generated SSH config (written by coati process)
coati/
  cmd/coati/                # CLI entry point
  internal/adapters/        # Port implementations (filesystem, GitHub API)
  internal/core/domain/     # Domain types and validation
  internal/core/ports/      # Interfaces
  internal/core/services/   # Business logic (generators, cache, config)
  internal/templates/       # Embedded default templates
make/                       # Build and install scripts
```

## Development

```bash
make build      # Compile binary to bin/coati
make test       # Run all tests
make quality    # Format, vet, and lint
make install    # Install to ~/.local/bin
make apply      # Build, process, and apply config to /etc/hosts and ~/.ssh/config
```

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feat/description`
3. Commit with Conventional Commits: `git commit -m "feat: add X"`
4. Push and open a pull request

## License

This project is licensed under the MIT License — see [LICENSE](LICENSE) for details.

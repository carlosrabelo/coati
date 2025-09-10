# Coati

CLI that generates `/etc/hosts` and `~/.ssh/config` from a YAML definition stored locally or fetched from a private GitHub Gist.

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.25%2B-blue.svg)](https://go.dev/)

## Highlights

- Define hosts, aliases, and SSH options in a single YAML file
- Generate `/etc/hosts` and `~/.ssh/config` from the same definitions
- Sync a private GitHub Gist with `coati pull` / `coati push`
- System and user scopes (`--scope all|system|user`) with `SUDO_USER` awareness
- Single system data store in `/var/lib/coati` and cache in `/var/cache/coati`
- Import, merge, check, dry-run, and automatic `.bak` backups
- System-wide shell completion via `sudo coati completion bash|zsh|fish`

## Prerequisites

- **Go 1.25+** — required to build from source; [download](https://go.dev/dl/)
- **Root** — required for `pull`/`push`, `/etc/hosts`, and `completion`
- **Readable** `/var/lib/coati/hosts.yaml` — for non-root `--scope user`

## Installation

### Build from Source

```bash
git clone https://github.com/carlosrabelo/coati.git
cd coati
make build
```

Install to `~/.local/bin` (default), or system-wide to `/usr/local/bin` (sudo only for the copy):

```bash
make install
make install-system
make uninstall
make uninstall-system
```

## Usage

Full guides (also on [GitHub Pages](https://carlosrabelo.github.io/coati/)): [User Guide](docs/GUIDE.md) · [Gist Schema](docs/GIST.md)

```bash
# One-time system setup (root)
sudo mkdir -p /etc/coati /var/lib/coati /var/cache/coati
sudo tee /etc/coati/config.yaml >/dev/null <<'EOF'
gist_id: "<gist-id>"
github_token: "<pat>"
EOF
sudo chmod 600 /etc/coati/config.yaml

sudo coati pull
sudo coati process                  # --scope all (hosts + SSH of SUDO_USER)
coati process                       # --scope user (SSH only)
sudo coati completion bash
```

## Configuration

Credentials and hooks live in `/etc/coati/config.yaml` (mode `0600`):

```yaml
gist_id: "..."
github_token: "..."
allowed_hooks:
  - dnsmasq
```

Overrides:

| Env | Default |
|-----|---------|
| `COATI_CONFIG_DIR` | `/etc/coati` |
| `COATI_DATA_DIR` | `/var/lib/coati` |
| `COATI_CACHE_DIR` | `/var/cache/coati` |
| `GITHUB_TOKEN` | (optional; overrides file token) |

## Project Layout

```
coati/cmd/coati/         # CLI entry point
coati/internal/          # App, domain, services, adapters
bin/                     # Compiled binaries (git-ignored)
.make/                   # Build and install scripts
testdata/                # Test fixtures
docs/                    # User guides + GitHub Pages
```

Runtime (system-only):

```
/etc/coati/config.yaml
/var/lib/coati/hosts.yaml
/var/cache/coati/gists/
```

## Development

```bash
make build             # Compile binary to bin/coati
make test              # Run all tests
make quality           # Format, vet, and lint
make install           # Install binary to ~/.local/bin
make install-system    # Install binary to /usr/local/bin
make uninstall         # Remove from ~/.local/bin
make uninstall-system  # Remove from /usr/local/bin
```

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feat/description`
3. Commit with Conventional Commits: `git commit -m "feat: add X"`
4. Push and open a pull request

## License

This project is licensed under the MIT License — see [LICENSE](LICENSE) for details.

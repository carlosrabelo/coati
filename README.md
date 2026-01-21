# Coati

CLI that generates `/etc/hosts` and `~/.ssh/config` from a YAML definition stored locally or fetched from a private GitHub Gist.

## Highlights

- Define hosts, aliases, and SSH options in a single YAML file
- Generates `/etc/hosts` with proper formatting and column alignment
- Generates `~/.ssh/config` from the same host definitions
- Import command bootstraps configuration from existing `/etc/hosts` and `~/.ssh/config`
- Automatic file backups with permission preservation
- Pull config from a private GitHub Gist; push local changes back with `coati push`
- Caches Gist responses locally to reduce network calls
- Strict validation rejects duplicate hostnames before writing any file
- Merge mode preserves existing file content in named `# BEGIN ORIGINAL` sections
- Check mode shows a unified diff before any file is written
- Dry-run mode previews generated content without touching disk
- Shell completion for bash, zsh, fish, and PowerShell

---

## Documentation

For full details on using Coati, please refer to the following guides:

*   **[User Guide](docs/GUIDE.md)**: Details CLI commands, command-line flags, configuration options, automatic backup logic, and post-execution hooks.
*   **[Gist Schema Reference](docs/GIST.md)**: Explains the YAML file structure (defaults, hosts, CNAMEs, PTRs, post-hooks).

---

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

---

## Project Layout

```
bin/                        # Compiled binaries (ignored by git)
data/
  src/gist.txt              # Local Gist copy (written by coati pull)
  gen/etc/hosts             # Generated hosts file (written by coati process)
  gen/ssh/config            # Generated SSH config (written by coati process)
docs/                       # Comprehensive documentation and guides
coati/
  cmd/coati/                # CLI entry point
  internal/adapters/        # Port implementations (filesystem, GitHub API)
  internal/core/domain/     # Domain types and validation
  internal/core/ports/      # Interfaces
  internal/core/services/   # Business logic (generators, cache, config)
  internal/templates/       # Embedded default templates
.make/                      # Build and installation scripts
```

## Development

```bash
make build      # Compile the binary to bin/coati
make test       # Run all tests
make quality    # Format, check and lint code
make install    # Install to ~/.local/bin
make apply      # Compile, process and apply config to /etc/hosts and ~/.ssh/config
```

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feat/description`
3. Commit using Conventional Commits: `git commit -m "feat: add X"`
4. Push and open a pull request

## License

This project is licensed under the MIT License — see the [LICENSE](LICENSE) file for details.

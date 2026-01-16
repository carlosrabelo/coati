# Coati User Guide

This guide covers the CLI usage, commands, options, and advanced features of Coati.

---

## Table of Contents
- [Core Workflow](#core-workflow)
- [CLI Commands](#cli-commands)
  - [process](#process)
  - [import](#import)
  - [pull](#pull)
  - [push](#push)
  - [completion](#completion)
- [Advanced Features](#advanced-features)
  - [Automatic Backups](#automatic-backups)
  - [Post-Execution Hooks & Security](#post-execution-hooks--security)
  - [Merge Mode](#merge-mode)
  - [Check Mode](#check-mode)
- [GitHub Gist Synchronization](#github-gist-synchronization)

---

## Core Workflow

Coati uses a single YAML file (stored locally or fetched from a private GitHub Gist) as the source of truth to generate your `/etc/hosts` and `~/.ssh/config` files.

```
                  ┌──────────────┐
                  │ GitHub Gist  │
                  └──────┬───────┘
                         │ coati pull / process
                         ▼
┌────────────┐     ┌──────────────┐     ┌─────────────────┐
│ hosts.yaml ├────►│    Coati     ├────►│ /etc/hosts       │
└────────────┘     └──────────────┘     │ ~/.ssh/config   │
                                        └─────────────────┘
```

---

## CLI Commands

### `process`
Generates `/etc/hosts` and `~/.ssh/config` from the YAML configuration.

**Syntax**:
```bash
coati process [flags]
```

**Common Flags**:
*   `--hosts-list <path>`: Local YAML configuration file (default: `/etc/coati/hosts.yaml`).
*   `--output-hosts <path>`: Destination hosts file (default: `data/gen/etc/hosts`). To apply to system, use `/etc/hosts` (requires sudo).
*   `--output-config <path>`: Destination SSH config file (default: `data/gen/ssh/config`). To apply to system, use `~/.ssh/config`.
*   `--hosts-template <path>`: Path to a custom hosts template.
*   `--gist-id <id>`: GitHub Gist ID to fetch the config from.
*   `--gist-file <filename>`: The filename inside the Gist if it contains multiple files.
*   `--github-token <token>`: GitHub Personal Access Token.
*   `--save-config`: Securely saves `--gist-id` and `--github-token` in `/etc/coati/config.yaml` so they don't have to be specified on subsequent commands.
*   `--dry-run`: Prints generated content to stdout without writing any files.
*   `--check`: Displays a unified diff showing what would change without modifying files.
*   `--merge`: Preserves existing file content (wrapping it inside `# BEGIN ORIGINAL` / `# END ORIGINAL` markers) and updates only the `# BEGIN COATI` / `# END COATI` blocks.
*   `--backup`: Creates a `.bak` backup copy of the target files before overwriting (default: `true`).
*   `--allow-unsafe-hooks`: Bypasses command validation for post-execution hooks.
*   `--allowed-hooks <cmd1,cmd2>`: Comma-separated list of additional allowed commands for post-execution hooks.
*   `--force-refresh`, `-f`: Bypasses the local cache when fetching from Gist.
*   `--verbose`, `-v`: Enables debug log level.

---

### `import`
Imports your existing `/etc/hosts` and `~/.ssh/config` files and merges them into a single, clean `hosts.yaml` configuration.

**Syntax**:
```bash
coati import [flags]
```

**Flags**:
*   `--hosts-file <path>`: Path to the hosts file (default: `/etc/hosts`).
*   `--ssh-file <path>`: Path to the SSH config file (default: `~/.ssh/config`).
*   `--output <path>`: Path to save the generated YAML configuration (default: `hosts.yaml`). Pass `-` to print directly to stdout.

**How it works**:
1. Parses `/etc/hosts` to extract IP addresses, hostnames, aliases, and line comments.
2. Skips standard system loopback addresses (e.g. `127.0.0.1 localhost`, `::1 localhost`, etc.) to prevent configuration bloat.
3. Parses `~/.ssh/config` to extract host blocks and their properties (`HostName`, `User`, `Port`, `IdentityFile`, `Options`).
4. Merges them: If an SSH host corresponds to a hostname or alias in the hosts file, it consolidates them into a single entry. If no IP exists, it retains the entry as an SSH-only host.
5. Saves the output as a fully valid Coati YAML schema ready to be used or pushed to a Gist.

---

### `pull`
Downloads the remote Gist content and saves it to a local configuration file.

**Syntax**:
```bash
coati pull [flags]
```

**Flags**:
*   `--gist-id <id>`: GitHub Gist ID.
*   `--github-token <token>`: GitHub Token.
*   `--output <path>`: Local file to write (default: `data/src/gist.txt`).

---

### `push`
Uploads the local YAML configuration back to your GitHub Gist.

**Syntax**:
```bash
coati push [flags]
```

**Flags**:
*   `--gist-id <id>`: GitHub Gist ID.
*   `--github-token <token>`: GitHub Token.
*   `--input <path>`: Local file to upload (default: `data/src/gist.txt`).

---

### `completion`
Generates auto-completion scripts for your shell.

**Syntax**:
```bash
coati completion [bash|zsh|fish|powershell]
```

---

## Advanced Features

### Automatic Backups
By default, the `process` command creates a backup of target files before modifying them:
*   `/etc/hosts` is backed up to `/etc/hosts.bak`.
*   `~/.ssh/config` is backed up to `~/.ssh/config.bak`.
*   **File Permissions**: The backup copy preserves the original file's permissions (usually `0644` for hosts, `0600` for SSH config).
*   **Opt-out**: Bypassed by passing `--backup=false`.

### Post-Execution Hooks & Security
Post-execution hooks (`post_hooks` section in Gist YAML) allow you to run commands (like restarting a local DNS server) after Coati completes successfully.

For security, the commands you can run are restricted to a safe list:
`systemctl`, `service`, `docker`, `kubectl`, `nginx`, `apache2`, `httpd`.

If you need to execute other commands:
1.  **Command line override**: Use `--allowed-hooks` to pass additional allowed command names:
    ```bash
    coati process --allowed-hooks dnsmasq,unbound
    ```
2.  **Local config configuration**: Add `allowed_hooks` inside `/etc/coati/config.yaml`:
    ```yaml
    gist_id: ...
    github_token: ...
    allowed_hooks:
      - dnsmasq
      - custom-script
    ```
3.  **Complete Bypass**: Pass `--allow-unsafe-hooks` to skip name checks entirely. Note that characters like `;`, `&`, and `|` are always blocked in hook arguments to prevent shell injection.

### Merge Mode
Using `--merge` wraps original/manual entries in the target files in a named section:
```
# BEGIN ORIGINAL
# My manual host entries...
# END ORIGINAL

# BEGIN COATI
# Managed by Coati...
# END COATI
```
This is safe to run repeatedly.

### Check Mode
Using `--check` lets you view the diff in standard unified diff format:
```diff
--- /etc/hosts
+++ /etc/hosts
@@ -3,4 +3,5 @@
 127.0.0.1 localhost
+192.168.1.50 db-prod
```

---

## GitHub Gist Synchronization

1.  Generate a GitHub Personal Access Token (PAT) with `gist` permission scope.
2.  Create a secret Gist containing a single YAML file (e.g. `hosts.yaml`).
3.  Run Coati once to save your credentials locally:
    ```bash
    coati process --gist-id <your_gist_id> --github-token <your_pat> --save-config
    ```
4.  Run Coati normally. It will fetch the remote Gist, validate it, cache it, and compile the configurations.

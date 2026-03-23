# Configuration Reference

Coati uses a single YAML structure to define your infrastructure. This content should be stored in your private GitHub Gist.

> **Note**: The filename inside the Gist (e.g., `hosts.yaml`) does not strictly matter; Coati will read the first file it finds in the Gist. However, using `.yaml` or `.yml` extension is recommended for syntax highlighting.

## Structure Overview

The configuration is divided into three main sections:
1.  **defaults**: Global settings for SSH generation.
2.  **hosts**: List of your machines/servers.
3.  **cnames**: Aliases that point to other hostnames.

## 1. Defaults
These settings apply to `Host *` in the generated SSH config, serving as fallbacks.

```yaml
defaults:
  user: root                     # Default SSH user
  port: 22                       # Default SSH port
  identity_file: ~/.ssh/id_rsa   # Default private key
  options:                       # Arbitrary SSH options (Key: Value)
    StrictHostKeyChecking: "no"
    ServerAliveInterval: "120"
```

## 2. Hosts
This list defines your actual endpoints. Each entry generates a line in `/etc/hosts` AND a `Host` block in `~/.ssh/config`.

```yaml
hosts:
  - ip: "10.0.0.50"              # [Required] IP Address
    hostname: "db-prod"          # [Required] Canonical Hostname
    
    # [Optional] List of aliases. 
    # Added to /etc/hosts line and SSH 'Host' line.
    aliases: ["db", "postgresql"] 
    
    # [Optional] SSH User for this specific host
    user: "admin"
    
    # [Optional] SSH Port
    port: 2222
    
    # [Optional] Specific Key
    identity_file: "~/.ssh/db_key"
    
    # [Optional] DNS Name for PTR/Reverse DNS logic (adds to aliases)
    dns_name: "ec2-10-0-0-50.compute.amazonaws.com"
    
    # [Optional] Extra Arbitrary SSH Options
    options:
      ForwardAgent: "yes"
      
    # [Optional] Comment added to /etc/hosts line
    comment: "Primary Database"
```

## 3. CNAMEs (Aliases)
Use this to create convenient shortcuts to existing hosts without duplicating IPs.

```yaml
cnames:
  - target: "db-prod"            # Must match a 'hostname' defined above
    alias:                       # Note: Key is 'alias' (singular)
      - "writer.db" 
      - "reader.db"
```

## 4. PTRs (Reverse DNS)
Map an IP to a domain. This helps when you want to look up a domain by IP (adds to host aliases).

```yaml
ptrs:
  - ip: "200.0.35.148"
    domain: "p148.venkon.org"
```

## 5. Simplified Mode
List of hostnames that should NOT have extra Aliases/CNAMEs/PTRs processed. Useful for specific hosts where you want strict control.

```yaml
simplified_mode_hosts:
  - "m492721"
```

## Full Example

```yaml
defaults:
  user: ubuntu
  identity_file: ~/.ssh/id_rsa

hosts:
  - ip: "192.168.1.10"
    hostname: "web-01"
    aliases: ["www"]

  - ip: "192.168.1.11"
    hostname: "web-02"

cnames:
  - target: "web-01"
    alias: 
      - "primary"
      - "loadbalancer"

ptrs:
  - ip: "192.168.1.10"
    domain: "web-01.internal"

simplified_mode_hosts:
  - "web-02"
```

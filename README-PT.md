# Coati

CLI que gera `/etc/hosts` e `~/.ssh/config` a partir de uma definição YAML armazenada localmente ou obtida de um GitHub Gist privado.

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.25%2B-blue.svg)](https://go.dev/)
[![Release](https://img.shields.io/github/release/carlosrabelo/coati.svg)](https://github.com/carlosrabelo/coati/releases)

## Destaques

- Defina hosts, aliases e opções SSH em um único arquivo YAML
- Gere `/etc/hosts` e `~/.ssh/config` a partir das mesmas definições
- Sincronize um Gist privado com `coati pull` / `coati push`
- Escopos system e user (`--scope all|system|user`) com suporte a `SUDO_USER`
- Uma única fonte de dados em `/var/lib/coati` e cache em `/var/cache/coati`
- Import, merge, check, dry-run e backups `.bak` automáticos
- Autocompletar system-wide via `sudo coati completion bash|zsh|fish`

## Pré-requisitos

- **Go 1.25+** — necessário para compilar a partir do código-fonte; [download](https://go.dev/dl/)
- **Root** — necessário para `pull`/`push`, `/etc/hosts` e `completion`
- **`/var/lib/coati/hosts.yaml` legível** — para `--scope user` sem root

## Instalação

### A partir do Código-Fonte

```bash
git clone https://github.com/carlosrabelo/coati.git
cd coati
make build
```

Instale em `~/.local/bin` (padrão), ou defina `PREFIX` para um prefixo próprio. A instalação no sistema é via `.deb`:

```bash
make install
make uninstall
make package   # só para CLIs compilados — veja Desenvolvimento
sudo dpkg -i dist/*.deb
```

## Uso

Guias completos (também em [GitHub Pages](https://carlosrabelo.github.io/coati/)): [Guia de Uso](docs/GUIDE-PT.md) · [Sintaxe do Gist](docs/GIST-PT.md)

```bash
# Setup do sistema (root, uma vez)
sudo mkdir -p /etc/coati /var/lib/coati /var/cache/coati
sudo tee /etc/coati/config.yaml >/dev/null <<'EOF'
gist_id: "<gist-id>"
github_token: "<pat>"
EOF
sudo chmod 600 /etc/coati/config.yaml

sudo coati pull
sudo coati process                  # --scope all (hosts + SSH do SUDO_USER)
coati process                       # --scope user (só SSH)
sudo coati completion bash
```

## Configuração

Credenciais e hooks ficam em `/etc/coati/config.yaml` (modo `0600`):

```yaml
gist_id: "..."
github_token: "..."
allowed_hooks:
  - dnsmasq
```

Overrides:

| Env | Padrão |
|-----|--------|
| `COATI_CONFIG_DIR` | `/etc/coati` |
| `COATI_DATA_DIR` | `/var/lib/coati` |
| `COATI_CACHE_DIR` | `/var/cache/coati` |
| `GITHUB_TOKEN` | (opcional; sobrescreve o token do arquivo) |

## Estrutura do Projeto

```
coati/cmd/coati/         # Ponto de entrada CLI
coati/internal/          # App, domain, services, adapters
bin/                     # Binários compilados (ignorado pelo git)
dist/                    # Pacotes .deb (ignorado pelo git)
.make/                   # Scripts de build, instalação e empacotamento
testdata/                # Fixtures de teste
docs/                    # Guias de uso + GitHub Pages
```

Em runtime (só sistema):

```
/etc/coati/config.yaml
/var/lib/coati/hosts.yaml
/var/cache/coati/gists/
```

## Desenvolvimento

```bash
make build           # Compila o binário em bin/coati
make test            # Executa todos os testes
make quality         # Formata, verifica e executa o linter
make install         # Instala o binário em ~/.local/bin
make uninstall       # Remove de ~/.local/bin
make package         # Gera o .deb em dist/ (binário em /usr/bin)
make deb             # Atalho para package
```

## Contribuição

1. Fork o repositório
2. Crie uma branch de feature: `git checkout -b feat/descricao`
3. Commit com Conventional Commits: `git commit -m "feat: add X"`
4. Push e abra um pull request

## Licença

Este projeto está licenciado sob a Licença MIT — veja [LICENSE](LICENSE) para detalhes.

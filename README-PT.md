# Coati

CLI que gera `/etc/hosts` e `~/.ssh/config` a partir de uma definição YAML armazenada localmente ou obtida de um GitHub Gist privado.

## Destaques

- Defina hosts, aliases e opções SSH em um único arquivo YAML
- Gera `/etc/hosts` com formatação e alinhamento de colunas adequados
- Gera `~/.ssh/config` a partir das mesmas definições de host
- Baixe a config de um GitHub Gist privado; envie alterações locais de volta com `coati push`
- Armazena respostas do Gist localmente para reduzir chamadas de rede
- Validação rigorosa rejeita hostnames e IPs duplicados antes de escrever qualquer arquivo
- Modo merge preserva o conteúdo existente em seções nomeadas `# BEGIN ORIGINAL`
- Modo check exibe um diff unificado antes de qualquer arquivo ser escrito
- Modo dry-run visualiza o conteúdo gerado sem tocar o disco
- Suporte a auto-completar para bash, zsh, fish e PowerShell

## Sumário

- [Visão Geral](#visão-geral)
- [Pré-requisitos](#pré-requisitos)
- [Instalação](#instalação)
- [Início Rápido](#início-rápido)
- [Uso](#uso)
- [Configuração](#configuração)
- [Estrutura do Projeto](#estrutura-do-projeto)
- [Desenvolvimento](#desenvolvimento)
- [Contribuição](#contribuição)
- [Licença](#licença)

## Visão Geral

Coati fornece uma única fonte de verdade para definições de hosts. Em vez de editar manualmente `/etc/hosts` e `~/.ssh/config`, defina seus servidores uma vez em um arquivo YAML e o Coati cuida do resto. Armazene o YAML em um GitHub Gist privado para mantê-lo sincronizado entre máquinas.

## Pré-requisitos

- **Go 1.25+** — necessário para compilar a partir do código-fonte; [download](https://go.dev/dl/)
- **Permissões de escrita** para `/etc/hosts` (requer sudo)
- **Permissões de escrita** para `~/.ssh/config`

## Instalação

### A partir do Código-Fonte

```bash
git clone https://github.com/carlosrabelo/coati
cd coati
make build
```

Instalar em `~/.local/bin` (sem root):

```bash
make install
```

### Usando Go Install

```bash
go install github.com/carlosrabelo/coati/cmd/coati@latest
```

## Início Rápido

1. Crie um arquivo de configuração YAML:

```bash
cat > hosts.yaml << 'EOF'
hosts:
  - hostname: my-server
    ip: 192.168.1.100
EOF
```

2. Processe e verifique:

```bash
coati process --hosts-list hosts.yaml --output-hosts data/gen/etc/hosts
cat data/gen/etc/hosts
# 192.168.1.100    my-server
```

## Uso

### process

Gera `/etc/hosts` e `~/.ssh/config` a partir de um arquivo YAML:

```bash
coati process --hosts-list hosts.yaml
```

Escreve diretamente nos caminhos do sistema (requer sudo para `/etc/hosts`):

```bash
coati process --output-hosts /etc/hosts --output-config ~/.ssh/config
```

### pull / push

Baixa o conteúdo do Gist para `data/src/gist.txt`:

```bash
coati pull
```

Envia `data/src/gist.txt` de volta para o Gist:

```bash
coati push
```

Ambos os comandos leem `--gist-id` e `--github-token` de flags, da variável de ambiente `GITHUB_TOKEN`, ou da config salva em `/etc/coati/config.yaml`.

### Flags avançadas

- **Dry Run**: Visualiza o conteúdo gerado sem escrever.
  ```bash
  coati process --dry-run
  ```

- **Check**: Exibe um diff unificado entre os arquivos atuais e o que seria escrito.
  ```bash
  coati process --check
  coati process --check --merge
  ```

- **Merge**: Preserva o conteúdo existente em uma seção `# BEGIN ORIGINAL`; gerencia apenas a seção `# BEGIN COATI`. Seguro para executar repetidamente.
  ```bash
  sudo coati process --merge --output-hosts /etc/hosts
  ```

- **Gist File**: Seleciona um arquivo específico quando o Gist contém múltiplos arquivos.
  ```bash
  coati process --gist-id abc123 --gist-file work.yaml
  ```

- **Forçar Atualização**: Ignora o cache local e busca do Gist.
  ```bash
  coati process --force-refresh
  ```

- **Shell Completion**: Instala o auto-completar para o seu shell.
  ```bash
  coati completion bash
  coati completion zsh
  coati completion fish
  ```

## Configuração

### Formato YAML

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

### Salvar credenciais do Gist

Execute uma vez para armazenar o ID do Gist e o token em `/etc/coati/config.yaml`:

```bash
coati process --gist-id SEU_GIST_ID --github-token SEU_TOKEN --save-config
```

Após salvar, `coati pull`, `coati push` e `coati process` funcionam sem flags.

### Variáveis de ambiente

- `GITHUB_TOKEN` — token do GitHub para acesso ao Gist
- `COATI_CONFIG_DIR` — sobrescreve o diretório de configuração padrão (`/etc/coati`)

## Estrutura do Projeto

```
bin/                        # Binários compilados (ignorado pelo git)
data/
  src/gist.txt              # Cópia local do Gist (escrita por coati pull)
  gen/etc/hosts             # Arquivo hosts gerado (escrito por coati process)
  gen/ssh/config            # SSH config gerado (escrito por coati process)
coati/
  cmd/coati/                # Ponto de entrada CLI
  internal/adapters/        # Implementações de ports (filesystem, GitHub API)
  internal/core/domain/     # Tipos de domínio e validação
  internal/core/ports/      # Interfaces
  internal/core/services/   # Lógica de negócio (geradores, cache, config)
  internal/templates/       # Templates padrão embutidos
make/                       # Scripts de build e instalação
```

## Desenvolvimento

```bash
make build      # Compila o binário em bin/coati
make test       # Executa todos os testes
make quality    # Formata, verifica e executa o linter
make install    # Instala em ~/.local/bin
make apply      # Compila, processa e aplica a config em /etc/hosts e ~/.ssh/config
```

## Contribuição

1. Fork o repositório
2. Crie uma branch de feature: `git checkout -b feat/descricao`
3. Commit com Conventional Commits: `git commit -m "feat: add X"`
4. Push e abra um pull request

## Licença

Este projeto está licenciado sob a Licença MIT — veja o arquivo [LICENSE](LICENSE) para detalhes.

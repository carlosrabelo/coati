# Coati

CLI que gera `/etc/hosts` e `~/.ssh/config` a partir de uma definição YAML armazenada localmente ou obtida de um GitHub Gist privado.

## Destaques

- Defina hosts, aliases e opções SSH em um único arquivo YAML
- Gera `/etc/hosts` com formatação e alinhamento de colunas adequados
- Gera `~/.ssh/config` a partir das mesmas definições de host
- O comando de importação extrai as configurações atuais do `/etc/hosts` e `~/.ssh/config`
- Backups automáticos dos arquivos modificados preservando as permissões originais
- Baixe a config de um GitHub Gist privado; envie alterações locais de volta com `coati push`
- Armazena respostas do Gist localmente para reduzir chamadas de rede
- Validação rigorosa rejeita hostnames e IPs duplicados antes de escrever qualquer arquivo
- Modo merge preserva o conteúdo existente em seções nomeadas `# BEGIN ORIGINAL`
- Modo check exibe um diff unificado antes de qualquer arquivo ser escrito
- Modo dry-run visualiza o conteúdo gerado sem tocar o disco
- Suporte a auto-completar para bash, zsh, fish e PowerShell

---

## Documentação

Para informações completas sobre o uso do Coati, consulte os seguintes guias:

*   **[Guia de Uso](docs/GUIDE-PT.md)**: Detalha comandos CLI, flags de linha de comando, opções de configuração, lógica de backups automáticos e hooks de pós-execução.
*   **[Referência de Sintaxe do Gist](docs/GIST-PT.md)**: Explica a estrutura do arquivo YAML (defaults, hosts, CNAMEs, PTRs, post-hooks).

---

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

---

## Estrutura do Projeto

```
bin/                        # Binários compilados (ignorado pelo git)
data/
  src/gist.txt              # Cópia local do Gist (escrita por coati pull)
  gen/etc/hosts             # Arquivo hosts gerado (escrito por coati process)
  gen/ssh/config            # SSH config gerado (escrito por coati process)
docs/                       # Documentação abrangente e guias de uso
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

Este projeto está licenciado sob a Licença MIT — veja o arquivo [LICENSE](LICENSE) file for details.

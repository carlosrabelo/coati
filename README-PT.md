# Coati

Coati é uma ferramenta CLI moderna para gerenciar arquivos locais `/etc/hosts` e configurações SSH (`~/.ssh/config`). Permite definir sua infraestrutura em um formato YAML limpo e gerar os arquivos de sistema necessários automaticamente.

## Destaques

- Defina hosts, aliases e opções SSH em um único arquivo YAML
- Gera `/etc/hosts` automaticamente com formatação e alinhamento adequados
- Gera `~/.ssh/config` a partir das mesmas definições de host
- Busca configuração de um GitHub Gist privado; selecione um arquivo específico com `--gist-file`
- Armazena respostas do Gist localmente para reduzir chamadas de rede
- Validação rigorosa rejeita hostnames e IPs duplicados antes de escrever qualquer arquivo
- Executa comandos personalizados após a geração bem-sucedida da configuração via hooks
- Visualize alterações sem escrever em disco com o modo dry-run
- Veja exatamente o que mudaria com um diff unificado antes de confirmar com `--check`
- Mescla entradas geradas em arquivos existentes, preservando o conteúdo original em seções nomeadas
- Suporte a auto-completar para bash, zsh, fish e PowerShell

## Sumário

- [Destaques](#destaques)
- [Visão Geral](#visão-geral)
- [Pré-requisitos](#pré-requisitos)
- [Instalação](#instalação)
- [Início Rápido](#início-rápido)
- [Uso](#uso)
- [Configuração](#configuração)
- [Estrutura do Projeto](#estrutura-do-projeto)
- [Desenvolvimento](#desenvolvimento)
- [Testes](#testes)
- [Solução de Problemas](#solução-de-problemas)
- [Contribuindo](#contribuindo)
- [Licença](#licença)

## Visão Geral

Coati simplifica o gerenciamento de infraestrutura fornecendo uma única fonte de verdade para definições de hosts. Em vez de editar manualmente `/etc/hosts` e `~/.ssh/config`, você define seus servidores uma vez em um arquivo YAML e o Coati cuida do resto. Essa abordagem garante consistência em seus sistemas, reduz erros manuais e facilita o compartilhamento de configurações com sua equipe.

## Pré-requisitos

- **Go 1.25+** (para compilar a partir do código-fonte)
- **Arquivo YAML** (arquivo de configuração)
- **Permissões de escrita** para `/etc/hosts` (requer sudo)
- **Permissões de escrita** para `~/.ssh/config`

## Instalação

### A partir do Código-Fonte

```bash
git clone https://github.com/carlosrabelo/coati
cd coati
make install
```

### Usando Go Install

```bash
go install github.com/carlosrabelo/coati/cmd/coati@latest
```

## Início Rápido

Comece em menos de 2 minutos:

1. Crie um arquivo de configuração:
```bash
cat > hosts.yaml << 'EOF'
hosts:
  - hostname: my-server
    ip: 192.168.1.100
EOF
```

2. Execute o Coati:
```bash
sudo coati apply --hosts-list hosts.yaml --output-hosts /etc/hosts
```

3. Verifique:
```bash
cat /etc/hosts
# Saída: 192.168.1.100    my-server
```

## Uso

### Uso Básico

1. Crie um arquivo de configuração (ex: `hosts.yaml`).
2. Execute o Coati:

```bash
sudo coati apply --hosts-list hosts.yaml --output-hosts /etc/hosts --output-config ~/.ssh/config
```

### Formato de Configuração

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

### Comandos Avançados

- **Dry Run**: Visualize o conteúdo gerado completo com saída colorida, sem escrever.
  ```bash
  coati apply --dry-run
  ```

- **Check**: Exibe um diff unificado entre os arquivos atuais e o que seria escrito.
  ```bash
  coati apply --check
  coati apply --check --merge   # diff do resultado mesclado
  ```

- **Merge**: Preserva o conteúdo existente em uma seção `# BEGIN ORIGINAL` e gerencia
  apenas a seção `# BEGIN COATI`. Seguro para executar repetidamente.
  ```bash
  sudo coati apply --merge --output-hosts /etc/hosts
  ```

- **Gist File**: Seleciona um arquivo específico quando o Gist contém múltiplos arquivos YAML.
  ```bash
  coati apply --gist-id abc123 --gist-file work.yaml
  ```

- **Forçar Atualização**: Ignora o cache e busca do Gist.
  ```bash
  coati apply --force-refresh
  ```

- **Modo Detalhado**: Ativa logs de debug.
  ```bash
  coati apply --verbose
  ```

- **Shell Completion**: Instala o auto-completar para o seu shell.
  ```bash
  coati completion bash
  coati completion zsh
  coati completion fish
  ```

## Configuração

### Configuração Padrão

Uma configuração padrão é fornecida em `data/cfg/config.yaml`:

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

### Variáveis de Ambiente

- `GITHUB_TOKEN`: Token do GitHub para acesso ao Gist
- `COATI_CONFIG_DIR`: Diretório de configuração personalizado

## Estrutura do Projeto

```
coati/
├── bin/                    ← Binários compilados
├── data/                   ← Arquivos de configuração e saída
│   ├── cfg/                ← Arquivos de configuração
│   │   └── config.yaml     ← Configuração padrão
│   └── out/                ← Arquivos de saída gerados
├── cmd/                    ← Ponto de entrada CLI
│   └── coati/              ← Aplicação principal
├── internal/               ← Pacotes internos
│   ├── adapters/           ← Implementações de portas
│   │   └── secondary/      ← Adaptadores de saída
│   ├── core/               ← Lógica de negócios
│   │   ├── domain/         ← Modelos de domínio
│   │   ├── ports/          ← Interfaces
│   │   └── services/       ← Serviços de aplicação
│   └── templates/          ← Templates embarcados
├── make/                   ← Scripts de automação
│   ├── build.sh            ← Compilar projeto
│   ├── test.sh             ← Executar testes
│   ├── clean.sh            ← Limpar artefatos
│   ├── install.sh          ← Instalar binário
│   └── uninstall.sh        ← Remover binário
├── test/                   ← Testes de integração
│   └── testdata/           ← Dados de teste
├── Makefile                ← Automação de build
├── README.md               ← Documentação em inglês
└── README-PT.md            ← Documentação em português
```

## Desenvolvimento

```bash
make apply      # Compila, gera e aplica a config em /etc/hosts e ~/.ssh/config
make build      # Compila o binário em bin/coati
make test       # Executa todos os testes
make quality    # Formata, verifica e executa o linter
make install    # Instala em ~/.local/bin
```

## Testes

### Executando Testes

```bash
make test
```

Ou diretamente:

```bash
./run/test.sh
```

### Cobertura de Testes

```bash
go test -cover ./...
```

### Estrutura de Testes

- **Testes unitários**: `**/*_test.go`
- **Testes de integração**: `test/integration_test.go`
- **Dados de teste**: `test/testdata/`

### Cobertura Atual

- `cmd/coati`: ~50% (testes básicos para validação de hooks)
- `internal/adapters/secondary`: ~67%
- `internal/core/domain`: ~85%
- `internal/core/services`: ~95%
- `internal/core/ports`: 0% (apenas interfaces)
- `internal/templates`: 0% (templates embarcados)

## Solução de Problemas

### Problema: "comando não encontrado"

**Solução**: Verifique se a instalação foi concluída com sucesso:
```bash
which coati
# Deve mostrar: /usr/local/bin/coati ou ~/.local/bin/coati
```

Se não for encontrado, reinstale:
```bash
make install
```

### Problema: "permissão negada ao escrever /etc/hosts"

**Solução**: Execute o Coati com sudo:
```bash
sudo coati --hosts-list hosts.yaml --output-hosts /etc/hosts
```

### Problema: "falha na validação do hook"

**Solução**: Verifique se o comando do hook está na lista permitida:
```bash
# Comandos permitidos: systemctl, service, docker, kubectl, nginx, apache2, httpd
# Comandos não podem conter: ;, &, |
```

### Problema: "conexão recusada ao buscar do Gist"

**Solução**: Verifique seu token do GitHub e conexão de rede:
```bash
export GITHUB_TOKEN=seu_token_aqui
coati --verbose
```

### Problema: "cache não expira"

**Solução**: Force a atualização para ignorar o cache:
```bash
coati --force-refresh
```

## Contribuindo

Contribuições são bem-vindas! Por favor, siga estas diretrizes:

1. Fork o repositório
2. Crie uma branch de feature: `git checkout -b feature/sua-feature`
3. Faça suas alterações
4. Escreva testes para novas funcionalidades
5. Certifique-se de que todos os testes passam: `make test`
6. Formate seu código: `gofmt -w .`
7. Execute os linters: `go vet ./...`
8. Commit suas alterações: `git commit -m "feat: descrição"`
9. Push para a branch: `git push origin feature/sua-feature`
10. Abra um Pull Request

### Estilo de Código

- Siga as convenções padrão do Go
- Mantenha as funções focadas e pequenas
- Adicione documentação de pacote
- Escreva testes para todas as funções públicas
- Use logs estruturados com `slog`

## Licença

Este projeto está licenciado sob a Licença MIT - veja o arquivo [LICENSE](LICENSE) para detalhes.

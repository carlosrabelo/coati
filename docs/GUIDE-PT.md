# Guia de Uso do Coati

Este guia detalha o uso da CLI, comandos, opções e recursos avançados do Coati.

---

## Sumário
- [Fluxo Básico](#fluxo-básico)
- [Comandos da CLI](#comandos-da-cli)
  - [process](#process)
  - [import](#import)
  - [pull](#pull)
  - [push](#push)
  - [completion](#completion)
- [Recursos Avançados](#recursos-avançados)
  - [Backups Automáticos](#backups-automáticos)
  - [Hooks de Pós-Execução e Segurança](#hooks-de-pós-execução-e-segurança)
  - [Modo Merge](#modo-merge)
  - [Modo Check](#modo-check)
- [Sincronização com GitHub Gist](#sincronização-com-github-gist)

---

## Fluxo Básico

O Coati usa um único arquivo YAML (salvo localmente ou em um GitHub Gist privado) como fonte da verdade para gerar os arquivos `/etc/hosts` e `~/.ssh/config`.

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

## Comandos da CLI

### `process`
Gera os arquivos `/etc/hosts` e `~/.ssh/config` a partir da configuração YAML.

**Sintaxe**:
```bash
coati process [flags]
```

**Flags Comuns**:
*   `--hosts-list <caminho>`: Arquivo local de configuração YAML (padrão: `/etc/coati/hosts.yaml`).
*   `--output-hosts <caminho>`: Arquivo de hosts de destino (padrão: `data/gen/etc/hosts`). Para aplicar ao sistema, use `/etc/hosts` (requer sudo).
*   `--output-config <caminho>`: Arquivo SSH config de destino (padrão: `data/gen/ssh/config`). Para aplicar ao sistema, use `~/.ssh/config`.
*   `--hosts-template <caminho>`: Caminho para um template personalizado de hosts.
*   `--gist-id <id>`: ID do GitHub Gist contendo o YAML de configuração.
*   `--gist-file <nome>`: O nome do arquivo dentro do Gist se ele contiver múltiplos arquivos.
*   `--github-token <token>`: Token de Acesso Pessoal (PAT) do GitHub.
*   `--save-config`: Salva com segurança o `--gist-id` e `--github-token` em `/etc/coati/config.yaml` para uso posterior sem a necessidade de informar flags.
*   `--dry-run`: Exibe o conteúdo gerado no stdout sem modificar os arquivos.
*   `--check`: Mostra um diff unificado com as alterações propostas sem aplicar.
*   `--merge`: Preserva conteúdo existente nos arquivos (envolvendo-o entre delimitadores `# BEGIN ORIGINAL` / `# END ORIGINAL`) e atualiza apenas os blocos `# BEGIN COATI` / `# END COATI`.
*   `--backup`: Cria uma cópia de segurança `.bak` dos arquivos de destino antes de gravar (padrão: `true`).
*   `--allow-unsafe-hooks`: Ignora a validação do allowlist de comandos para hooks.
*   `--allowed-hooks <cmd1,cmd2>`: Lista separada por vírgulas de comandos adicionais permitidos para post-hooks.
*   `--force-refresh`, `-f`: Ignora o cache local e força o download do Gist.
*   `--verbose`, `-v`: Habilita logs detalhados de depuração.

---

### `import`
Importa seus arquivos `/etc/hosts` e `~/.ssh/config` existentes e os consolida em uma configuração `hosts.yaml` limpa.

**Sintaxe**:
```bash
coati import [flags]
```

**Flags**:
*   `--hosts-file <caminho>`: Caminho para o arquivo hosts existente (padrão: `/etc/hosts`).
*   `--ssh-file <caminho>`: Caminho para o arquivo SSH config existente (padrão: `~/.ssh/config`).
*   `--output <caminho>`: Caminho para salvar a configuração YAML gerada (padrão: `hosts.yaml`). Use `-` para exibir diretamente no terminal.

**Como funciona**:
1. Lê `/etc/hosts` para extrair endereços IP, hostnames, aliases e comentários de linha.
2. Descarta mapeamentos de loopback padrão do sistema (ex: `127.0.0.1 localhost`, `::1 localhost`, etc.) para evitar poluição.
3. Lê `~/.ssh/config` para extrair os blocos de hosts e suas propriedades (`HostName`, `User`, `Port`, `IdentityFile`, `Options`).
4. Mescla-os: se um host do SSH corresponder a um hostname ou alias no arquivo hosts, eles são agrupados na mesma entrada. Se não houver IP correspondente, a entrada é mantida como host apenas de SSH.
5. Exporta o YAML validado pronto para ser usado no Coati ou publicado no Gist.

---

### `pull`
Baixa a configuração do Gist remoto e a salva localmente.

**Sintaxe**:
```bash
coati pull [flags]
```

**Flags**:
*   `--gist-id <id>`: ID do GitHub Gist.
*   `--github-token <token>`: Token do GitHub.
*   `--output <caminho>`: Caminho do arquivo local (padrão: `data/src/gist.txt`).

---

### `push`
Envia o arquivo local de configuração de volta para o GitHub Gist.

**Sintaxe**:
```bash
coati push [flags]
```

**Flags**:
*   `--gist-id <id>`: ID do GitHub Gist.
*   `--github-token <token>`: Token do GitHub.
*   `--input <caminho>`: Caminho do arquivo local (padrão: `data/src/gist.txt`).

---

### `completion`
Gera scripts de autocompletar comandos para diferentes shells.

**Sintaxe**:
```bash
coati completion [bash|zsh|fish|powershell]
```

---

## Recursos Avançados

### Backups Automáticos
Por padrão, o comando `process` realiza uma cópia de segurança antes de sobrescrever os arquivos:
*   `/etc/hosts` é copiado para `/etc/hosts.bak`.
*   `~/.ssh/config` é copiado para `~/.ssh/config.bak`.
*   **Permissões**: A cópia preserva as permissões de acesso originais do arquivo (normalmente `0644` para hosts, `0600` para SSH config).
*   **Desativação**: Pode ser pulado passando `--backup=false`.

### Hooks de Pós-Execução e Segurança
Os hooks (`post_hooks` no YAML) permitem executar comandos (como recarregar um serviço DNS local) após a gravação bem-sucedida.

Por segurança, por padrão apenas comandos seguros são permitidos:
`systemctl`, `service`, `docker`, `kubectl`, `nginx`, `apache2`, `httpd`.

Caso precise rodar outros comandos:
1.  **Parâmetro de linha de comando**: Use `--allowed-hooks` para registrar novos comandos:
    ```bash
    coati process --allowed-hooks dnsmasq,unbound
    ```
2.  **Configuração local**: Adicione o campo `allowed_hooks` no arquivo `/etc/coati/config.yaml`:
    ```yaml
    gist_id: ...
    github_token: ...
    allowed_hooks:
      - dnsmasq
      - script-customizado
    ```
3.  **Liberação Total**: Use `--allow-unsafe-hooks` para ignorar a validação de executáveis. Caracteres de injeção de shell (como `;`, `&`, `|`) continuarão bloqueados nos parâmetros dos hooks.

### Modo Merge
Com a flag `--merge`, as linhas manuais originais são salvas em blocos próprios:
```
# BEGIN ORIGINAL
# Minhas entradas manuais...
# END ORIGINAL

# BEGIN COATI
# Gerenciado pelo Coati...
# END COATI
```
Isso garante segurança ao rodar o comando consecutivas vezes.

### Modo Check
Usando `--check` você confere as modificações propostas no formato de diff unificado tradicional:
```diff
--- /etc/hosts
+++ /etc/hosts
@@ -3,4 +3,5 @@
 127.0.0.1 localhost
+192.168.1.50 db-prod
```

---

## Sincronização com GitHub Gist

1.  Gere um Token de Acesso Pessoal (PAT) no GitHub com escopo de permissão `gist`.
2.  Crie um Gist secreto contendo um arquivo YAML (ex.: `hosts.yaml`).
3.  Execute o Coati uma vez salvando suas credenciais localmente:
    ```bash
    coati process --gist-id <seu_gist_id> --github-token <seu_pat> --save-config
    ```
4.  Rode o Coati normalmente. Ele baixará o Gist, validará o YAML, guardará no cache local e atualizará seus arquivos.

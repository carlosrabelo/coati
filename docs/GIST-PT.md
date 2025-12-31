# Referência de Configuração

O Coati usa uma estrutura YAML única para definir sua infraestrutura. Este conteúdo deve ser armazenado em seu GitHub Gist privado.

> **Nota**: O nome do arquivo dentro do Gist (ex.: `hosts.yaml`) não é relevante; o Coati lê o primeiro arquivo que encontrar no Gist. Recomenda-se usar a extensão `.yaml` ou `.yml` para realce de sintaxe.

## Visão Geral da Estrutura

A configuração é dividida em três seções principais:
1. **defaults**: Configurações globais para geração do SSH.
2. **hosts**: Lista de máquinas/servidores.
3. **cnames**: Aliases que apontam para outros hostnames.

## 1. Defaults
Essas configurações se aplicam ao bloco `Host *` no SSH config gerado, funcionando como valores padrão.

```yaml
defaults:
  user: root                     # Usuário SSH padrão
  port: 22                       # Porta SSH padrão
  identity_file: ~/.ssh/id_rsa   # Chave privada padrão
  options:                       # Opções SSH arbitrárias (Chave: Valor)
    StrictHostKeyChecking: "no"
    ServerAliveInterval: "120"
```

## 2. Hosts
Esta lista define seus endpoints reais. Cada entrada gera uma linha no `/etc/hosts` E um bloco `Host` no `~/.ssh/config`.

```yaml
hosts:
  - ip: "10.0.0.50"              # [Obrigatório] Endereço IP
    hostname: "db-prod"          # [Obrigatório] Hostname canônico
    
    # [Opcional] Lista de aliases.
    # Adicionados à linha do /etc/hosts e à linha 'Host' do SSH.
    aliases: ["db", "postgresql"]
    
    # [Opcional] Usuário SSH específico para este host
    user: "admin"
    
    # [Opcional] Porta SSH
    port: 2222
    
    # [Opcional] Chave específica
    identity_file: "~/.ssh/db_key"
    
    # [Opcional] Nome DNS para lógica PTR/DNS reverso (adicionado aos aliases)
    dns_name: "ec2-10-0-0-50.compute.amazonaws.com"
    
    # [Opcional] Opções SSH adicionais arbitrárias
    options:
      ForwardAgent: "yes"
      
    # [Opcional] Comentário adicionado à linha do /etc/hosts
    comment: "Banco de Dados Principal"
```

## 3. CNAMEs (Aliases)
Use para criar atalhos convenientes para hosts existentes sem duplicar IPs.

```yaml
cnames:
  - target: "db-prod"            # Deve corresponder a um 'hostname' definido acima
    alias:                       # Nota: chave é 'alias' (singular)
      - "writer.db"
      - "reader.db"
```

## 4. PTRs (DNS Reverso)
Mapeia um IP para um domínio. Útil quando você quer buscar um domínio pelo IP (adicionado aos aliases do host).

```yaml
ptrs:
  - ip: "200.0.35.148"
    domain: "p148.venkon.org"
```

## 5. Modo Simplificado
Lista de hostnames que não devem ter Aliases/CNAMEs/PTRs processados. Útil para hosts específicos onde você quer controle estrito.

```yaml
simplified_mode_hosts:
  - "m492721"
```

## Exemplo Completo

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

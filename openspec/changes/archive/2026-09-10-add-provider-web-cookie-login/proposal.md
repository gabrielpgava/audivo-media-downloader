## Why

O Audivo atualmente exige que o usuário exporte e selecione manualmente um arquivo de cookies do Apple Music, e não oferece nenhum fluxo equivalente para autenticar o YouTube. Isso torna a configuração difícil, deixa o YouTube sem uma forma explícita de reutilizar uma sessão autenticada e permite que o caminho de cookies de um provedor seja aplicado ao outro.

O change adiciona um fluxo de login iniciado em Configurações para Apple Music e YouTube: o usuário abre a página do provedor em uma janela de autenticação do Audivo, faz login se necessário e confirma a captura local da sessão. Os cookies permanecem no dispositivo e são associados ao provedor correto.

## What Changes

- Adicionar, em Configurações, ações separadas para conectar Apple Music e YouTube.
- Abrir uma janela de autenticação controlada pelo aplicativo no domínio oficial do provedor, reutilizando a sessão dessa janela quando já houver login e exibindo a tela de login quando não houver.
- Detectar uma sessão autenticada, capturar os cookies necessários localmente e mostrar somente o estado da conexão, nunca o conteúdo ou os valores dos cookies.
- Persistir cookies em arquivos locais protegidos, com caminhos separados para Apple Music e YouTube, e corrigir o roteamento para que cada engine receba apenas os cookies do próprio provedor.
- Manter a seleção manual de arquivo Netscape como fallback quando a captura automática não estiver disponível ou não puder ser validada.
- Exibir estados de conexão, carregamento, cancelamento, sessão inválida e erro acionável sem bloquear o restante das Configurações.
- Atualizar testes, bindings Wails e documentação de privacidade/uso para cobrir login, armazenamento local, sanitização de logs e reautenticação.

## Capabilities

### New Capabilities

- `provider-web-authentication`: fluxo de autenticação web por provedor, janela de login, detecção de sessão, captura local, cancelamento e estados de erro.
- `provider-cookie-management`: armazenamento local protegido, validação, expiração/reconexão, importação manual e entrega dos cookies corretos aos engines.

### Modified Capabilities

- Nenhuma. Não há especificações principais existentes em `openspec/specs/`; os contratos atuais pertencem a changes anteriores ainda ativos.

## Impact

- Backend Go: modelos de Settings e status de autenticação, serviço de configuração/download, validação de cookies, sanitização de logs e integração com a captura web por plataforma.
- Camada Wails: novos métodos para iniciar/cancelar/consultar autenticação e selecionar/importar cookies, além dos bindings TypeScript gerados.
- Frontend React/MUI: novas ações e estados no diálogo de Configurações, com identidade explícita do provedor e feedback responsivo.
- Engines: `yt-dlp` passará a receber o arquivo de cookies do YouTube quando configurado; `gamdl` continuará recebendo somente o arquivo do Apple Music.
- Plataforma e distribuição: suporte a uma janela web de autenticação e seus adaptadores nos sistemas suportados, sem backend remoto ou envio de cookies ao Audivo.
- Qualidade e documentação: testes unitários de contratos/segurança, testes de componente e aceitação manual com contas descartáveis; a captura real depende de rede, login e políticas dos provedores.

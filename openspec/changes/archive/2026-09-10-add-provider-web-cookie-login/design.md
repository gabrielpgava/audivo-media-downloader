## Context

O Audivo é um aplicativo Wails v2 local com frontend React/MUI, serviços Go e engines externos. A configuração atual persiste apenas `AppleMusicCookiesPath`, abre um seletor nativo para um arquivo Netscape e passa esse caminho ao `gamdl`. O `yt-dlp` ainda não recebe cookies configurados; o serviço inclusive consulta o campo do Apple Music durante o fluxo do YouTube, o que impede separar as sessões por provedor.

O runtime Wails disponível oferece `BrowserOpenURL`, que abre o navegador padrão do sistema, mas não fornece ao aplicativo os cookies criados nessa navegação. Um iframe também não é suficiente para capturar cookies protegidos ou de outra origem. A solução precisa, portanto, manter a autenticação local, funcionar sem backend do Audivo e evitar que conteúdo de cookies chegue ao React, aos logs ou a um serviço remoto.

## Goals / Non-Goals

**Goals:**

- Oferecer em Configurações uma ação independente para conectar Apple Music e YouTube.
- Reutilizar uma sessão persistente pertencente ao Audivo; detectar uma sessão já autenticada e capturar os cookies sem pedir credenciais novamente.
- Abrir uma janela web visível do provedor quando a sessão não estiver autenticada, permitindo login, MFA e consentimentos diretamente pelo usuário.
- Gerar/validar arquivos Netscape locais, privados e específicos por provedor.
- Entregar cookies do YouTube somente ao `yt-dlp` e cookies do Apple Music somente ao `gamdl`.
- Preservar a importação manual como fallback e expor estados de conexão, erro, cancelamento e navegador indisponível.
- Permitir desconectar/reconectar localmente sem criar conta Audivo, telemetria ou serviço intermediário.

**Non-Goals:**

- Ler ou copiar silenciosamente perfis/cookies existentes do Chrome, Safari, Edge ou Firefox do usuário.
- Automatizar credenciais, contornar MFA, CAPTCHA, DRM, assinatura, restrições regionais ou políticas dos provedores.
- Enviar cookies, URLs ou mídia para um backend próprio.
- Transformar o Audivo em um navegador geral ou suportar navegação para URLs arbitrárias na janela de autenticação.
- Garantir que uma sessão autenticada permita baixar qualquer conteúdo protegido; a autorização final continua sendo decidida pelo provedor e pelo engine.

## Decisions

### 1. Janela de autenticação controlada pelo aplicativo

O login será executado em uma sessão web visível e separada, iniciada pelo backend com um perfil persistente dentro do diretório privado de dados do Audivo. A sessão será controlada por um adaptador CDP para um navegador Chromium instalado na máquina, com um diretório de perfil distinto por provedor. O adaptador deve localizar apenas navegadores suportados, abrir a URL inicial oficial do provedor e expor a leitura da loja de cookies da própria sessão.

Essa escolha permite que o aplicativo saiba quando a sessão está pronta e capture cookies `HttpOnly` sem depender de `document.cookie`. Também evita alterar a janela principal Wails ou importar credenciais de um navegador pessoal. Se não houver um navegador compatível, a UI informa o motivo e mantém a importação manual.

Alternativas consideradas:

- `runtime.BrowserOpenURL`: rejeitada porque apenas delega ao navegador padrão e não retorna cookies ao processo Go.
- `iframe` no diálogo React: rejeitada por isolamento de origem, políticas de framing e cookies `HttpOnly`.
- Ler o banco de cookies do navegador padrão: rejeitada por locks, keychain/DPAPI, diferenças entre navegadores e risco de copiar sessões sem consentimento explícito.
- Navegador headless: rejeitado porque o login, MFA e CAPTCHA precisam ser visíveis e interativos.

### 2. Contrato de sessão e eventos

O backend terá uma interface `AuthBrowser` testável com operações de abrir perfil, navegar em uma URL permitida, observar a sessão, obter cookies, fechar e cancelar. Um `AuthManager` serializará uma sessão ativa por vez e manterá o estado fora do frontend:

```text
React Configurações
    -> StartProviderAuth(provider)
    -> AuthManager -> AuthBrowser (janela visível + perfil privado)
    <- provider-auth:updated { provider, state, message }
    -> GetProviderAuthStatus / CancelProviderAuth
    -> CookieStore -> Settings + engine
```

Os estados públicos serão, no mínimo, `idle`, `opening`, `waiting_login`, `connected`, `cancelled`, `unavailable` e `error`. O payload Wails poderá conter provedor, estado, mensagem acionável, identificador opaco da sessão e timestamp, mas nunca valor de cookie, header, conteúdo de arquivo ou comando completo.

O usuário poderá cancelar pela UI. O cancelamento encerra o processo/navegador e remove qualquer arquivo temporário; uma sessão válida anterior permanece intacta até uma captura nova ser validada e promovida atomicamente.

### 3. Armazenamento e serialização de cookies

O `CookieStore` criará um diretório privado de cookies derivado de `platform.Paths`, separado do JSON de preferências. A captura será convertida para Netscape com domínio, subdomínio, caminho, flag Secure, expiração, nome e valor. O arquivo será escrito com permissão restritiva, sincronizado e promovido por rename atômico; valores não serão persistidos no JSON de configurações.

Cada provedor terá um caminho independente (`AppleMusicCookiesPath` e `YouTubeCookiesPath`, ou estrutura equivalente compatível com o JSON atual). A validação exige formato Netscape, arquivo legível, pelo menos um registro e domínio pertencente à lista permitida do provedor. A importação manual deve copiar o arquivo para a área privada do Audivo para aplicar as mesmas permissões e validações.

O change preservará um `AppleMusicCookiesPath` legado válido durante a migração. Uma nova sessão ou importação o normalizará para o armazenamento gerenciado; configurações inválidas retornam aos defaults sem exibir conteúdo sensível.

### 4. Integração específica com engines

Os builders de argumentos receberão um caminho de cookies opcional e validado quando o serviço for YouTube. O `yt-dlp` usará `--cookies <caminho>` tanto na análise de metadata quanto no download quando houver uma sessão configurada; sem cookies, o fluxo público existente continua funcionando. O `gamdl` continuará recebendo `--cookies-path <caminho>` apenas para Apple Music, e Apple Music continuará exigindo cookies antes da análise/download.

O serviço deixará de consultar `AppleMusicCookiesPath` no ramo do YouTube. Logs serão sanitizados com todos os caminhos de cookies conhecidos e com as opções `--cookies`/`--cookies-path`; mensagens de erro usarão códigos por provedor, sem incluir valores ou conteúdo.

### 5. Experiência de Configurações

O diálogo terá uma seção `Contas e sessões` com duas linhas visualmente identificadas: Apple Music e YouTube. Cada linha exibirá `Não conectado`, `Conectando`, `Conectado` ou uma falha acionável, além de `Conectar`, `Importar arquivo` e `Desconectar` quando aplicável.

Ao clicar em `Conectar`, a janela web é aberta e o estado muda para `waiting_login`. Se a sessão já estiver autenticada, o backend captura e promove os cookies automaticamente. Caso contrário, a UI orienta o usuário a concluir o login na janela do provedor; a detecção pode concluir sozinha quando os cookies válidos aparecerem e também deve oferecer uma ação explícita de `Verificar sessão` para casos em que o provedor não permita um sinal confiável.

O diálogo continua utilizável durante a autenticação, mas impede iniciar uma segunda sessão. Operações não relacionadas, como tema e pasta de downloads, não devem ser bloqueadas por uma falha de login.

### 6. Segurança e privacidade

- URLs de navegação são constantes por provedor e passam por allowlist de domínios de autenticação; a janela não recebe uma URL digitada pelo usuário.
- O perfil web é app-owned, não sincroniza com contas do navegador do sistema e fica sob o diretório de dados privado do Audivo.
- O frontend recebe apenas estado/mensagem; o backend mantém os caminhos e arquivos.
- Capturas incompletas não substituem cookies válidos. Desconectar remove o arquivo gerenciado e limpa o caminho salvo; não promete fazer logout no provedor.
- Testes automatizados usam cookies sintéticos e perfis temporários. Contas reais só aparecem na aceitação manual, nunca no CI ou em logs.

## Risks / Trade-offs

- **Navegador Chromium não instalado ou incompatível** → detectar antes de abrir, informar a plataforma e manter importação manual Netscape.
- **Mudança de fluxo, domínio ou expiração de cookies do provedor** → manter allowlists/seletores isolados por provedor, validar a sessão com um request/metadata mínimo do engine e registrar smoke tests manuais; nunca depender de um único nome de cookie público.
- **Login com CAPTCHA/MFA ou bloqueio de automação** → usar janela visível e interação humana; se a sessão não puder ser capturada, preservar a opção de exportação manual.
- **Perfil persistente contém credenciais de sessão** → armazenar em diretório privado com permissões restritivas, oferecer `Desconectar`, não sincronizar e documentar que o arquivo local equivale a uma credencial.
- **Falha durante substituição do arquivo** → escrever em temporário, validar antes do rename e manter o arquivo anterior até a nova captura estar completa.
- **Compatibilidade multiplataforma do CDP/CGO** → encapsular o navegador atrás de uma interface, executar fakes no CI e marcar como não suportada a plataforma sem navegador detectável, em vez de simular sucesso.
- **Cookies do YouTube são opcionais, mas podem ser necessários para conteúdo restrito** → retornar erro de autenticação acionável do engine e oferecer Configurações, sem tornar todo download público dependente de login.

## Migration Plan

1. Adicionar os modelos de provedor/auth e os novos campos de configuração com defaults compatíveis. Ler o caminho legado do Apple Music e só migrá-lo quando ele passar pela validação/cópia gerenciada.
2. Implementar `CookieStore`, serialização Netscape, permissões, allowlists, importação manual, limpeza e testes sem alterar ainda a UI.
3. Implementar o adaptador `AuthBrowser`/`AuthManager` e uma implementação falsa para ciclo de vida, cancelamento, sessão existente e navegador indisponível.
4. Integrar serviço Go, eventos Wails, métodos de iniciar/cancelar/consultar, correção do roteamento YouTube e builders de argumentos; regenerar bindings.
5. Implementar as linhas de provedores no diálogo de Configurações, estados, feedback, importação e desconexão; adicionar testes de componente e acessibilidade.
6. Atualizar README, aviso de privacidade, smoke record e CI. Rodar testes determinísticos; executar a aceitação real apenas com contas descartáveis e registrar limitações por plataforma.

O rollback é local: remover o fluxo web deixa os métodos de seleção manual e os arquivos Netscape existentes como fallback. Não há migração de banco nem dados remotos; qualquer arquivo gerenciado pode ser removido sem afetar downloads já concluídos.

## Open Questions

- Quais executáveis Chromium e versões mínimas serão suportados em cada runner (Chrome, Edge, Chromium, Brave) e como serão localizados sem baixar um navegador adicional?
- O sinal automático de sessão autenticada será suficientemente estável por provedor, ou a primeira versão exigirá sempre o botão `Verificar sessão` após o login?
- Quais domínios de autenticação e registros Netscape devem compor a allowlist inicial após uma rodada manual em macOS, Windows e Linux?
- A política de distribuição aceita depender de um navegador instalado pelo usuário, ou será necessário empacotar/baixar um runtime dedicado em uma mudança posterior?

## ADDED Requirements

### Requirement: A área principal SHALL aceitar entradas universais sem duplicar pipeline

O sistema SHALL aceitar digitação, Ctrl+V/Cmd+V, botão `Colar` e drag-and-drop de URL/texto quando suportado, encaminhando todas as fontes para a mesma normalização, validação, detecção e análise. Arquivos locais SHALL ser ignorados neste ciclo.

#### Scenario: Botão Colar com URL válida
- **WHEN** o usuário clica em `Colar` e o clipboard contém URL suportada
- **THEN** o campo é preenchido e a análise pode ser solicitada sem iniciar download automático

#### Scenario: Clipboard sem URL compatível
- **WHEN** o clipboard não contém URL válida de provider suportado
- **THEN** a UI informa que nenhum link compatível foi encontrado e mantém o usuário no estado de entrada

#### Scenario: Texto arrastado
- **WHEN** o usuário arrasta texto contendo uma URL suportada para a zona de entrada
- **THEN** a URL é normalizada pela mesma pipeline do campo de texto

### Requirement: Detecção de clipboard ao recuperar foco SHALL ser sugestiva e não intrusiva

Quando habilitada, a detecção de foco SHALL sugerir uma URL compatível encontrada no clipboard, mas SHALL nunca baixar automaticamente, substituir texto já digitado ou abrir modal bloqueante.

#### Scenario: Campo vazio ao recuperar foco
- **WHEN** o Audivo recebe foco com URL compatível no clipboard e o campo está vazio
- **THEN** a UI oferece a ação discreta `Colar`

#### Scenario: Usuário já digitou
- **WHEN** o Audivo recebe foco e o campo contém texto do usuário
- **THEN** o clipboard não substitui o texto nem abre uma interrupção

### Requirement: Preferências SHALL ser lembradas e aplicar defaults seguros

O sistema SHALL persistir preferências recentes de formato, qualidade, destino e fluxo. Ao analisar conteúdo, SHALL selecionar a preferência compatível ou a melhor alternativa disponível e SHALL explicar a adaptação quando ela diferir da preferência.

#### Scenario: Preferência 1080p indisponível
- **WHEN** o usuário prefere vídeo 1080p e a mídia oferece no máximo 720p
- **THEN** o sistema seleciona 720p como melhor disponível e informa essa escolha sem erro técnico

#### Scenario: Preferência de áudio
- **WHEN** o usuário configurou áudio MP3 e analisa novo conteúdo compatível
- **THEN** MP3 aparece como default sem exigir reconfiguração

### Requirement: O modo rápido SHALL reduzir escolhas sem iniciar automaticamente

Configurações SHALL permitir `Sempre mostrar opções` ou `Usar minhas preferências`. Mesmo no modo rápido, o sistema SHALL analisar e mostrar preview antes do início, e SHALL exigir ação explícita para baixar.

#### Scenario: Modo rápido com single
- **WHEN** uma URL individual é analisada no modo de preferências
- **THEN** a UI mostra a decisão aplicada e uma ação clara `Baixar`, sem iniciar o processo sozinha

#### Scenario: Modo rápido com coleção
- **WHEN** uma playlist ou álbum é analisado no modo rápido
- **THEN** a UI mostra o preview resumido e aguarda a confirmação de baixar todos

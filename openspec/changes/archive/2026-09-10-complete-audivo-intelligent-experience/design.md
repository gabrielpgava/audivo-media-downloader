## Context

O repositório é um aplicativo Wails v2 pequeno, com backend Go e frontend React/TypeScript. Hoje `App` mantém apenas o contexto Wails, `DownloadYoutube` e `DownloadAppleMusic` montam comandos diretamente, e o frontend possui duas telas separadas que exibem logs de subprocesso como texto. Não há fila, análise prévia, persistência, contrato de progresso estruturado ou testes de domínio.

O complemento precisa evoluir essa base sem trocar a stack nem expor a complexidade de yt-dlp, gamdl e FFmpeg. A experiência deve continuar sendo uma entrada única de URL com decisão explícita do usuário para iniciar o download. O escopo inclui somente YouTube e Apple Music, com histórico e preferências locais, sem conta, servidor, analytics, sincronização ou plugin externo.

## Goals / Non-Goals

**Goals:**

- Criar uma pipeline única de entrada, análise e download, com providers internos substituíveis.
- Representar downloads individuais e coleções como jobs explícitos, com uma unidade de execução completa por vez.
- Emitir estados e progresso estruturados suficientes para preview, cancelamento, retry, histórico e UI simples.
- Preservar arquivos concluídos, evitar sobrescrita e tratar paths, espaço, temporários e metadata de forma segura.
- Manter histórico e preferências locais, limitados e apagáveis, sem indexar o sistema de arquivos.
- Dar mensagens acionáveis para falhas e diagnóstico copiável sem segredos.
- Entregar testes de domínio e integração com engines por fixtures/mocks, além de documentação operacional.

**Non-Goals:**

- Não adicionar Spotify, Deezer, Tidal, Instagram, TikTok ou outros providers neste ciclo.
- Não criar player, biblioteca de mídia, editor de tags, conversor genérico, cloud sync, login, API pública ou extensão de navegador.
- Não criar marketplace/plugin loader; providers são uma abstração interna.
- Não executar downloads automaticamente ao colar ou ao detectar algo no clipboard.
- Não implementar auto-update silencioso; a primeira versão apenas detecta e abre a release oficial.
- Não paralelizar downloads completos nem criar scheduler complexo.

## Decisions

### 1. Pipeline de domínio com providers internos

Introduzir `NormalizeInput`, `ValidateURL`, `ProviderRegistry`, `Analyze` e `DownloadManager` como fronteiras de domínio. Cada provider implementa detecção, análise, download e health check; `YouTubeProvider` encapsula yt-dlp e `AppleMusicProvider` encapsula gamdl. A aplicação conhece DTOs de mídia e jobs, não flags de CLI.

Alternativas consideradas:

- **Condicionais nas telas e no `App`:** rejeitada porque espalharia diferenças de provider e tornaria coleções, erro e health inconsistentes.
- **Sistema de plugins externos:** rejeitada porque aumenta superfície de segurança e configuração sem requisito de produto.

### 2. `DownloadManager` com fila FIFO e uma execução completa

Um job de coleção terá itens/subjobs e será processado em ordem, com no máximo um processo de download completo ativo. O provider pode continuar usando paralelismo interno que a própria engine considere adequado. O manager será dono de cancelamento, retry, transições, progresso, resultado e atualização do histórico.

Alternativas consideradas:

- **Vários downloads completos em paralelo:** rejeitada por consumo imprevisível, colisões mais difíceis e cancelamento menos confiável.
- **Um goroutine independente por botão:** rejeitada porque não oferece estado único, recuperação de coleção ou controle de fila.

### 3. Contrato Wails estruturado

Substituir logs crus como contrato primário por DTOs versionáveis (`MediaInfo`, `CollectionItem`, `DownloadRequest`, `DownloadProgress`, `DownloadResult`, `ProviderHealth`, `DiagnosticReport`). Eventos Wails transportarão estados e progresso serializáveis; mensagens brutas continuarão somente nos logs sanitizados. Métodos de binding terão respostas e erros explícitos.

Alternativas consideradas:

- **Continuar concatenando stdout no frontend:** rejeitada porque o formato das engines não é API estável e impede progresso confiável.
- **WebSocket/servidor local:** rejeitada porque Wails já fornece transporte suficiente e um servidor ampliaria o escopo.

### 4. Adaptadores de subprocesso usando capacidades nativas das engines

Cada provider resolverá o binário de forma explícita, executará com `exec.CommandContext`, capturará stdout/stderr de maneira controlada e converterá linhas/conclusões para o domínio. Retry, resume, metadata e artwork serão delegados às opções nativas quando suportadas; o Audivo classificará o resultado e não duplicará lógica de engine sem necessidade.

Alternativas consideradas:

- **Implementar clientes HTTP/parsadores próprios para cada serviço:** rejeitada por duplicação e maior risco de incompatibilidade.
- **Deixar o frontend interpretar saída CLI:** rejeitada por acoplamento e vazamento de detalhes técnicos.

### 5. Persistência local versionada e limitada

Usar um pequeno arquivo de dados no diretório de dados/configuração do usuário, com versão de schema, escrita atômica e limite de registros. O histórico guardará somente downloads conhecidos pelo Audivo; preferências serão armazenadas separadamente no mesmo repositório local. O frontend nunca será responsável pela persistência de negócio.

Alternativas consideradas:

- **SQLite:** não é necessário para dezenas de registros e adicionaria dependência/migração a um app local simples.
- **`localStorage` do navegador:** rejeitado porque não atende bem paths do sistema, limpeza coordenada e acesso aos dados pelo backend.
- **Varredura de Downloads:** rejeitada explicitamente; o histórico não é indexador.

### 6. Segurança de nomes, paths e espaço

Um serviço de filesystem receberá títulos e metadados, sanitizará componentes por plataforma, preservará Unicode quando possível, tratará nomes reservados e aplicará limites de comprimento. Resolução de colisão será não destrutiva. Para tamanhos conhecidos, a verificação de espaço incluirá margem para arquivos temporários, muxing e conversão.

Alternativas consideradas:

- **Confiar nos templates das engines:** rejeitada porque não garante a mesma política em todas as plataformas/providers.
- **Sobrescrever ou apagar o destino existente:** rejeitada por risco de perda de dados.

### 7. State machine única no backend e projeção simples no frontend

Jobs usarão estados enumerados (`idle`, `analyzing`, `ready`, `queued`, `preparing`, `downloading`, `retrying`, `processing`, `completed`, `cancelled`, `error`, além dos estados de coleção). O frontend derivará botões e mensagens dessa máquina, em vez de combinar booleanos independentes. A home exibirá somente o próximo passo relevante; detalhes técnicos ficarão em configurações/diagnóstico.

Alternativas consideradas:

- **Booleanos `isDownloading`, `hasError`, `isRetrying` etc.:** rejeitados por combinações inválidas e condições de corrida.
- **State machine sofisticada/event-sourcing:** rejeitada por custo desnecessário para um aplicativo single-user local.

### 8. Diagnóstico e atualização assíncronos e sem segredos

Health checks completos ocorrerão sob demanda ou em background controlado, não em toda abertura de tela. O diagnóstico agregará versão, OS, arquitetura, providers, engines e operação recente com sanitização centralizada. A verificação de release consultará GitHub Releases no máximo uma vez por período configurado, filtrará pre-releases e nunca interromperá um download.

Alternativas consideradas:

- **Auditoria completa síncrona no startup:** rejeitada porque degrada o primeiro contato e pode bloquear a UI.
- **Auto-update silencioso:** rejeitado por risco operacional e por não ser necessário para a primeira entrega.

## Risks / Trade-offs

- [Integração com gamdl pode variar por versão e ambiente] → manter adapter isolado, health check específico, fixtures de erro e mensagens de dependência ausente/credencial expirada.
- [A saída de yt-dlp/gamdl pode mudar] → parsear somente sinais necessários, preferir opções de progresso documentadas e cobrir o adapter com testes de contrato.
- [Estimativa de tamanho pode ser desconhecida ou imprecisa] → tratar a verificação como preventiva, usar margem conservadora quando houver estimativa e nunca bloquear sem evidência.
- [Persistência local pode ser corrompida ou alterada manualmente] → escrita atômica, schema versionado, fallback para estado vazio e log sanitizado; nunca apagar mídia como recuperação.
- [Cancelar um subprocesso pode deixar temporários] → separar cancelamento manual de falha transitória, limpar somente arquivos identificados como seguros e preservar itens concluídos.
- [Paths válidos variam entre Windows, macOS e Linux] → testar casos representativos (`AC/DC`, `CON`, Unicode, comprimento) em helpers puros e aplicar regras da plataforma alvo.
- [Notificações Wails podem não ser uniformes] → manter feature opcional, sem dependência pesada, e tratar a ausência de suporte como degradação silenciosa.
- [O escopo é grande para a base atual] → implementar na ordem de dependências do `tasks.md`, mantendo wrappers temporários e gates por capacidade antes de avançar.

## Migration Plan

1. Adicionar modelos, helpers puros e contratos sem remover imediatamente os bindings atuais; cobrir normalização, paths, classificação, histórico e estados com testes.
2. Implementar adapters e registry, depois mover os bindings de YouTube/Apple Music para o `DownloadManager`, preservando a funcionalidade de download simples.
3. Adicionar persistência local, coleção, progresso e cancelamento; iniciar com fixtures/mocks e só então validar subprocessos reais em ambiente disponível.
4. Migrar a tela React para a pipeline única, mantendo preferências default compatíveis e sem exigir migração de dados existentes.
5. Integrar diagnóstico, logs, updates e documentação; validar build, testes, smoke de single download e cenários de coleção.

Rollback: reverter o binário/código para a versão anterior. Os novos arquivos de dados ficam isolados, versionados e podem ser ignorados/limpos pela configuração; nenhuma mídia existente é removida ou renomeada automaticamente. Se uma capability não estiver pronta, o manager deve falhar de forma explícita sem iniciar download parcial invisível.

## Open Questions

- Quais flags e formato de progresso da versão distribuída de gamdl devem ser tratados como contrato mínimo para álbum/playlist e cookies?
- Onde os binários de yt-dlp, gamdl e FFmpeg serão empacotados ou descobertos em cada plataforma de distribuição?
- Qual repositório/tag e estratégia de versão do Audivo serão a fonte oficial para a verificação de GitHub Releases?
- Quais APIs de notificação são aceitáveis na matriz de plataformas alvo sem adicionar uma dependência pesada?

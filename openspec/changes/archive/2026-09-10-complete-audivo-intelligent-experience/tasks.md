## 1. Contratos e estrutura de domínio

- [x] 1.1 Mapear os bindings Wails e os comandos atuais de YouTube/Apple Music para definir a fronteira entre frontend, domínio e adapters sem alterar a stack.
- [x] 1.2 Criar os tipos de domínio para `MediaKind`, provider, `MediaInfo`, item de coleção, `DownloadRequest`, estados de job, `DownloadProgress`, `DownloadResult` e `ProviderHealth`.
- [x] 1.3 Implementar `NormalizeInput`, validação de URL e `ProviderRegistry` com resultado explícito para URL inválida, não suportada e provider detectado.
- [x] 1.4 Definir nomes de eventos Wails e payloads serializáveis para análise, estado, progresso, resultado e atualização, sem usar stdout bruto como contrato.
- [x] 1.5 Criar fixtures e doubles de provider/engine que permitam testar a aplicação sem depender da rede ou de credenciais reais.

## 2. Persistência local, histórico e preferências

- [x] 2.1 Implementar resolução do diretório de dados do usuário e schema versionado para histórico e preferências, sem criar arquivos dentro do diretório do repositório.
- [x] 2.2 Implementar leitura/escrita atômica do armazenamento local com recuperação para arquivo ausente, inválido ou de versão desconhecida.
- [x] 2.3 Implementar o repositório de histórico limitado a 50 entradas, preservando URL normalizada, provider, título, tipo, path, destino, estado e timestamps sem armazenar cookies.
- [x] 2.4 Implementar recentes limitados, consulta de duplicidade por URL e verificação sob demanda do path registrado.
- [x] 2.5 Implementar limpeza explícita do histórico sem apagar arquivos de mídia e sem varrer diretórios externos.
- [x] 2.6 Implementar persistência das preferências de formato, qualidade, destino e modo de fluxo.
- [x] 2.7 Adicionar testes do repositório para limite, escrita atômica, arquivo corrompido, limpeza, duplicidade e arquivo removido.

## 3. Segurança de nomes, destinos e espaço

- [x] 3.1 Criar o serviço central de sanitização de componentes de path com regras por plataforma, Unicode preservado, separadores removidos, nomes reservados e limite de comprimento.
- [x] 3.2 Criar a política de destinos para single, playlist, álbum e artista/álbum, mantendo uma raiz padrão previsível do Audivo.
- [x] 3.3 Implementar resolução não destrutiva de colisões com sufixos determinísticos e identificação segura de arquivo já produzido pela mesma operação.
- [x] 3.4 Implementar estimativa de espaço livre com margem para temporários, muxing e conversão, sem bloquear quando o tamanho for desconhecido.
- [x] 3.5 Implementar a política de temporários, resume e limpeza segura para sucesso, falha recuperável e cancelamento manual.
- [x] 3.6 Adicionar testes para `AC/DC - Thunderstruck`, `What's Up?`, `Artist: Song`, `<Official Video>`, `CON`, Unicode, paths longos e colisões.

## 4. Adapters e health dos providers

- [x] 4.1 Implementar resolução de executáveis/recursos de yt-dlp, gamdl e FFmpeg com mensagens de dependência ausente e sem abrir terminal para o usuário.
- [x] 4.2 Implementar `YouTubeProvider` para análise e download individual via yt-dlp, incluindo opções de áudio/vídeo compatíveis com as preferências.
- [x] 4.3 Implementar análise de playlist YouTube com metadados resumidos, itens identificáveis e estimativas disponíveis sem iniciar os downloads.
- [x] 4.4 Implementar `AppleMusicProvider` para música individual via gamdl, incluindo descoberta segura de cookies/configuração sem gravá-los no histórico ou diagnóstico.
- [x] 4.5 Implementar suporte de análise/download para álbum e playlist Apple Music quando suportado pela versão disponível do gamdl, com fallback claro quando não suportado.
- [x] 4.6 Solicitar metadata e artwork pelas capacidades nativas das engines, sem scraping adicional, e mapear campos opcionais ausentes sem falhar o download.
- [x] 4.7 Implementar health check leve por provider e testes de detecção, dependência ausente, saída conhecida e provider não suportado usando fixtures.

## 5. DownloadManager, fila e resiliência

- [x] 5.1 Implementar a state machine de job individual e de coleção, impedindo combinações inválidas de booleanos.
- [x] 5.2 Implementar `DownloadManager` com fila FIFO, job pai/subjobs e no máximo um download completo ativo por vez.
- [x] 5.3 Executar subprocessos com contexto cancelável, captura controlada de stdout/stderr e encerramento sem deixar o frontend preso em `downloading`.
- [x] 5.4 Agregar progresso do item e da coleção, incluindo índice/total, percentuais, etapa, velocidade e ETA somente quando confiáveis.
- [x] 5.5 Implementar classificação de falhas transitórias/definitivas e retry limitado a até três tentativas, preferindo as capacidades nativas do provider.
- [x] 5.6 Implementar feedback de `retrying`, mensagens acionáveis para conteúdo privado, cookies expirados, engine desatualizada, formato inexistente e espaço insuficiente.
- [x] 5.7 Implementar cancelamento que interrompe o item atual, impede itens seguintes, preserva concluídos e limpa temporários somente conforme a política de segurança.
- [x] 5.8 Implementar resultado parcial e retomada que pule subjobs concluídos identificáveis, sem rebaixar os itens já preservados.
- [x] 5.9 Registrar resultado concluído/parcial no histórico e emitir evento final único por job, mantendo detalhes técnicos nos logs.
- [x] 5.10 Adicionar testes unitários do manager para FIFO, um processo ativo, progressão, cancelamento, retry, erro definitivo, retomada e coleção parcial.

## 6. Frontend: entrada única e download individual

- [x] 6.1 Substituir a navegação de duas telas por uma home única orientada pela entrada de URL, mantendo Material UI e o princípio de invisibilidade da infraestrutura.
- [x] 6.2 Implementar digitação, botão `Colar`, Ctrl+V/Cmd+V e drag-and-drop de URL/texto encaminhados à mesma pipeline de análise.
- [x] 6.3 Implementar sugestão discreta de clipboard ao recuperar foco, sem substituir texto existente, abrir modal bloqueante ou iniciar download automático.
- [x] 6.4 Implementar estados de análise, mídia pronta, URL inválida, provider não suportado e health indisponível com mensagens não técnicas.
- [x] 6.5 Implementar preview de download individual com formato/qualidade/destino derivados das preferências e ação explícita `Baixar`.
- [x] 6.6 Implementar progresso, cancelamento, conclusão, retry e erro do download individual usando somente eventos estruturados.
- [x] 6.7 Regenerar/ajustar os bindings TypeScript do Wails e remover dependência do console textual como feedback principal.

## 7. Frontend: coleções, recentes e configurações

- [x] 7.1 Implementar preview compacto de playlist/álbum com provider, título, autor/álbum, contagem e ação principal `Baixar todos`.
- [x] 7.2 Implementar visualização secundária de itens sem obrigar seleção manual para o fluxo padrão.
- [x] 7.3 Implementar progresso geral e do item atual, fila aguardando, tentativa de retry e estados `collection-partial`/`collection-completed`.
- [x] 7.4 Implementar confirmação de cancelamento de coleção e resumo de itens concluídos, pendentes e falhos sem apagar concluídos.
- [x] 7.5 Implementar recentes compactos com ações `Abrir`, `Abrir pasta`, aviso de duplicidade e fluxo de arquivo não encontrado.
- [x] 7.6 Implementar configurações pequenas para destino, áudio, vídeo, fluxo, aparência, cookies Apple Music, engines, diagnóstico e limpeza de histórico.
- [x] 7.7 Implementar modo rápido e fallback automático de preset, mostrando quando a melhor alternativa disponível diferir da preferência.
- [x] 7.8 Verificar acessibilidade básica, foco, estados vazios, responsividade e ausência de controles técnicos desnecessários.

## 8. Diagnóstico, logs e notificações

- [x] 8.1 Implementar agregador de diagnóstico com versão do Audivo, OS, arquitetura, Wails, providers e estados das engines.
- [x] 8.2 Implementar sanitização centralizada para cookies, tokens, headers, conteúdo de credenciais e paths sensíveis antes de copiar ou persistir detalhes.
- [x] 8.3 Implementar logs locais rotativos com limite de tamanho/quantidade e redaction de dados sensíveis, sem armazenar mídia ou cookies.
- [x] 8.4 Implementar `Copiar diagnóstico` e abertura opcional do GitHub Issues sem envio automático.
- [x] 8.5 Implementar notificação nativa opcional somente para conclusão/erro relevante em background, sem notificar cada etapa ou item.
- [x] 8.6 Adicionar testes de sanitização, rotação, conteúdo mínimo do relatório, ação externa e degradação sem suporte a notificações.

## 9. Verificação de atualizações

- [x] 9.1 Definir fonte oficial, versão instalada e comparação semântica para filtrar releases alpha, beta, rc e nightly.
- [x] 9.2 Implementar cache local da última verificação, com periodicidade padrão de 24 horas e execução assíncrona.
- [x] 9.3 Implementar aviso não intrusivo de release estável e ação para abrir a página oficial, sem auto-update silencioso ou interrupção de download.
- [x] 9.4 Adicionar testes de comparação de versões, filtro de pre-release, janela de cache, falha de rede e download ativo.

## 10. Startup, engines e documentação

- [x] 10.1 Implementar startup rápido com preparação/descoberta leve e feedback de primeira execução quando ferramentas precisarem ser preparadas.
- [x] 10.2 Garantir que operações longas de metadata, download, hash, preparação e update rodem fora da thread da UI.
- [x] 10.3 Atualizar o README com recursos reais, pré-requisitos, execução, testes, providers, histórico, duplicidade, retry, diagnóstico e limites do produto.
- [x] 10.4 Criar screenshots representativos de idle, preview, downloading, coleção, completed e settings sem expor console técnico.
- [x] 10.5 Criar `CONTRIBUTING.md` com fluxo de desenvolvimento, testes, arquitetura de providers, issue e PR.
- [x] 10.6 Criar template de bug report com versão, OS, arquitetura, provider, tipo de URL, passos, diagnóstico e aviso para nunca enviar cookies/credenciais.

## 11. Testes de aceitação e integração

- [x] 11.1 Criar fixture de playlist de cinco itens com quatro sucessos e um erro definitivo e verificar `collection-partial`, progresso e arquivos preservados.
- [x] 11.2 Criar fixture de cancelamento com item 1 concluído e item 2 ativo e verificar que item 3 não inicia.
- [x] 11.3 Cobrir URL nunca baixada, URL com arquivo presente, arquivo removido e escolha explícita de baixar novamente.
- [x] 11.4 Executar cenários de provider detection, health, metadata/artwork, retry transitório e classificação de erro definitivo sem rede real.
- [x] 11.5 Executar smoke de download individual com yt-dlp/gamdl disponíveis e registrar claramente qualquer validação externa bloqueada por credencial/ambiente.
- [ ] 11.6 Validar comportamento de destino, colisão, espaço e nomes em matriz de sistemas/fixtures multiplataforma.

## 12. Gates finais

- [x] 12.1 Rodar `gofmt`, testes Go, `go vet` e testes com race detector para os módulos de domínio e manager.
- [x] 12.2 Rodar formatação/lint, typecheck, testes e build do frontend com os bindings atualizados.
- [x] 12.3 Rodar build/execução Wails e verificar startup, single download, preview de coleção, cancelamento e encerramento limpo.
- [x] 12.4 Rodar validação estrita do OpenSpec para a mudança e `git diff --check`.
- [x] 12.5 Revisar `git status --short`, manter fora do escopo artefatos/credenciais locais e registrar no handoff o que foi validado, o que depende de engines reais e qualquer backlog restante.

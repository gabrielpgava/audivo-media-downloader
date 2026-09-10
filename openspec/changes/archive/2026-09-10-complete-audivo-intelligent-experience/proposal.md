## Why

O Audivo atualmente expõe downloads diretos e específicos por tela, enquanto o produto precisa interpretar melhor o link, organizar operações longas e recuperar falhas sem transferir decisões técnicas ao usuário. Esta mudança transforma a base atual em uma experiência única e simples para downloads individuais e coleções de YouTube e Apple Music, preservando Wails v2, Go, React, TypeScript, Vite, Material UI, yt-dlp, gamdl e FFmpeg.

## What Changes

- Unificar a entrada de URLs em uma pipeline de normalização, validação, detecção de provider e análise de mídia.
- Introduzir providers internos para YouTube/yt-dlp e Apple Music/gamdl, com health checks leves e detalhes das engines ocultos da interface.
- Reconhecer downloads individuais, playlists, álbuns e outras coleções suportadas; mostrar preview resumido antes de iniciar uma coleção.
- Criar um `DownloadManager` com fila de um download completo por vez, progresso do item e da coleção, cancelamento seguro e retomada parcial.
- Centralizar geração de nomes, diretórios previsíveis, sanitização multiplataforma, colisões sem sobrescrita e verificações de espaço em disco.
- Adicionar retry somente para falhas transitórias, classificação de erros definitivos, preservação segura de temporários e tradução de erros para mensagens acionáveis.
- Persistir um histórico local limitado dos downloads, recentes discretos, detecção de duplicidade e tratamento de arquivos removidos, sem indexar o computador nem enviar dados externamente.
- Adicionar clipboard/colar, drop zone de URLs, preferências lembradas, preset automático e modo rápido sem iniciar downloads automaticamente por padrão.
- Adicionar configurações compactas, diagnóstico copiável e sanitizado, logs rotativos locais, notificações nativas discretas e verificação periódica de releases estáveis.
- Atualizar README, screenshots, `CONTRIBUTING.md` e template de issue com instruções de execução, testes, arquitetura e coleta segura de diagnóstico.

## Capabilities

### New Capabilities

- `provider-orchestration`: pipeline de entrada, registro de providers, análise, download encapsulado e health checks.
- `collection-downloads`: preview, fila, progresso, estados, cancelamento e conclusão parcial de playlists/álbuns.
- `download-safety`: paths, nomes, metadata, artwork, colisões, resume e pré-verificação de espaço.
- `download-resilience`: retry transitório, classificação de falhas, feedback de recuperação e mensagens de erro acionáveis.
- `local-history`: histórico local limitado, recentes, duplicidade, arquivo ausente e limpeza com privacidade local.
- `input-and-preferences`: entrada universal de URLs, clipboard, foco, preferências, presets e modo rápido.
- `diagnostics-and-health`: diagnóstico de ambiente, sanitização, logs rotativos e relatório pronto para issue.
- `application-updates`: verificação periódica de releases estáveis e aviso não intrusivo sem auto-update silencioso.
- `documentation-and-release-readiness`: README, screenshots, contribuição, issue template e critérios de validação do produto.

### Modified Capabilities

Nenhuma. O repositório ainda não possui specs de capacidade existentes; todas as capacidades deste complemento serão introduzidas como contratos novos.

## Impact

- Backend Go: substituição das chamadas diretas espalhadas por serviços de provider, manager de downloads, repositórios locais, classificação de erros, diagnóstico e eventos Wails estruturados.
- Frontend React/TypeScript: uma tela principal orientada por estado, preview de coleções, progresso, recentes, configurações e feedback de erro sem expor engines ou parâmetros CLI.
- Persistência local: novo armazenamento pequeno e versionável para histórico e preferências; nenhum servidor, conta, analytics ou sincronização.
- Execução externa: integração real com os binários já previstos (`yt-dlp`, `gamdl`, `FFmpeg`) e seus mecanismos nativos de retry/resume quando confiáveis.
- Distribuição e documentação: arquivos de suporte ao projeto, verificação de versão via GitHub Releases e instruções para diagnóstico seguro.
- Escopo explicitamente fora: novos providers como Spotify/Deezer/Tidal, player, biblioteca de mídia, conversor genérico, cloud sync, login, API pública, extensões e auto-update silencioso.

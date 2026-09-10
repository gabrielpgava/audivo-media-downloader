# Registro de autenticação de provedores — 2026-08-17 e 2026-09-10

## Verificado automaticamente

- O contrato Wails expõe apenas estado, mensagem, identificador opaco de sessão e timestamp; caminhos e valores de cookies não atravessam o backend.
- Fixtures Netscape sintéticas cobrem Apple Music, YouTube, rejeição por provedor incorreto, formato inválido, substituição atômica, permissões privadas e desconexão.
- Fakes de navegador cobrem sessão já autenticada, login aguardando, `Verificar sessão`, captura inválida com retry, cancelamento, navegador indisponível e bloqueio de segundo login.
- O host macOS arm64 detectou Google Chrome `152.0.7977.83` e Microsoft Edge `147.0.3912.98`; os argumentos e limitações estão em [`provider-auth-browser-support.md`](provider-auth-browser-support.md).

## Aceitação manual pendente

Ainda é necessário executar com contas descartáveis, fora do CI, para ambos os provedores:

- sessão existente e novo login;
- MFA/CAPTCHA e consentimento;
- verificação manual após login;
- importação inválida e importação do provedor errado;
- cancelamento, desconexão, expiração e navegador indisponível;

Esses casos dependem de credenciais reais e não devem ser registrados com valores, headers, conteúdo ou caminhos de cookies.

## Smoke de download concluído — 2026-09-10

- Sessões locais já configuradas foram usadas pelo app nativo para analisar e baixar um vídeo público do YouTube e uma faixa do Apple Music.
- O resultado confirmou os caminhos de engine separados: YouTube via `yt-dlp` e Apple Music via `gamdl`; os detalhes técnicos não exibiram cookies, headers ou argumentos de cookies.
- `ffprobe` confirmou streams de áudio/vídeo no MP4 do YouTube e áudio no MP3 do Apple Music; os artefatos foram movidos para uma pasta temporária fora do repositório.
- O smoke público do YouTube sem sessão permanece coberto pelo registro de validação automatizada anterior; a aceitação de login manual continua pendente.

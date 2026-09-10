# Audivo Media Downloader

Aplicativo desktop local para baixar mídia de links do YouTube e Apple Music. O Audivo mantém a experiência em uma única tela: cole o link, confira o preview, escolha o formato, acompanhe o progresso e abra o arquivo concluído.

O projeto usa Wails v2, Go, React, TypeScript e MUI. Não há conta Audivo, backend próprio, banco de dados ou telemetria obrigatória.

## Escopo suportado

- YouTube: áudio MP3/M4A e vídeo MP4 com qualidade disponível até 2160p.
- Apple Music: músicas, álbuns e vídeos compatíveis com o `gamdl`; áudio M4A/MP3 e vídeo MP4.
- Uma tarefa de download ativa por vez, com cancelamento explícito e limpeza de processos filhos.
- Fila FIFO local com um processo completo ativo por vez; coleções exibem progresso geral, retry limitado e resultado parcial sem apagar concluídos.
- Histórico local limitado a 50 entradas, recentes até cinco itens, aviso de duplicidade e limpeza sem tocar na mídia.
- Preview estruturado, progresso visual, logs técnicos recolhidos, seleção de pasta e ações nativas para abrir arquivo/pasta.
- Diagnóstico copiável sanitizado, logs locais rotativos e verificação periódica de releases estáveis sem auto-update.

URLs de outros serviços, hosts parecidos e URLs com credenciais embutidas são rejeitados antes de qualquer engine ser executado.

## Evidência visual

Capturas do fluxo nativo Wails, sem console técnico: [idle](docs/screenshots/01-idle.png), [preview](docs/screenshots/02-preview.png), [downloading](docs/screenshots/03-downloading.png), [coleção](docs/screenshots/04-collection.png), [completed](docs/screenshots/05-completed.png) e [settings](docs/screenshots/06-settings.png).

## Desenvolvimento

Requisitos:

- Go 1.25 ou superior.
- Node.js compatível com o lockfile e npm.
- Wails CLI v2.15.0 para desenvolvimento e empacotamento.
- Python 3.10 ou superior somente para preparar o ambiente isolado do `gamdl`.

Na raiz do repositório:

```bash
cd frontend
npm ci
npm run typecheck
npm test
npm run build
cd ..

go test ./...
go vet ./...
go install github.com/wailsapp/wails/v2/cmd/wails@v2.15.0
$(go env GOPATH)/bin/wails generate module
$(go env GOPATH)/bin/wails dev
```

O build de produção é:

```bash
cd frontend && npm ci && npm run build && cd ..
$(go env GOPATH)/bin/wails generate module
$(go env GOPATH)/bin/wails build -clean
```

O `frontend/dist` precisa existir antes de comandos Go que compilam `main.go`, pois os assets são embutidos pelo Wails. A pasta é gerada e ignorada pelo Git.

## Engines

O backend resolve executáveis a partir do diretório de engines da aplicação e, como fallback, do `PATH`. O diretório gerenciado é derivado de APIs do sistema, separado por sistema operacional e arquitetura. Um manifesto JSON opcional em `engines.json` dentro do diretório de configuração permite instalar artefatos com versão, URL, executável e SHA-256; arquivos inválidos são descartados antes de serem ativados. Um exemplo está em `docs/engine-manifest.example.json`.

O botão **Atualizar** em Configurações executa uma verificação explícita. A verificação automática é limitada a uma vez por 24 horas. A implementação não versiona binários no repositório e não inclui executáveis pessoais ou específicos de uma máquina.

O YouTube usa `yt-dlp`, FFmpeg, ffprobe e Deno/EJS. O Apple Music prepara uma virtualenv local para `gamdl`; não executa `pip install` global nem exige privilégios administrativos.

## Apple Music e cookies

Apple Music exige uma assinatura válida e uma sessão autenticada. Em **Configurações > Contas e sessões**, use **Conectar** para abrir uma janela visível do Apple Music. Se a sessão do perfil privado do Audivo já estiver autenticada, os cookies são capturados automaticamente; caso contrário, conclua o login, MFA ou CAPTCHA nessa janela e use **Verificar sessão** se necessário.

O YouTube oferece o mesmo fluxo de forma independente. As sessões usam perfis persistentes próprios do Audivo, separados por provedor; o aplicativo não lê nem copia o perfil pessoal do Chrome, Safari, Edge, Firefox ou Brave. Se não houver um navegador Chromium compatível, **Importar arquivo** aceita um export Netscape válido como fallback.

Os cookies capturados ou importados são validados por domínio, copiados atomicamente para uma pasta privada local com permissão restritiva e nunca são enviados ao frontend, a um backend Audivo ou aos logs. **Desconectar** remove apenas a sessão local; não faz logout no site do provedor.

O `gamdl` exige Python 3.10+. A primeira preparação pode baixar dependências e levar algum tempo. Falhas de Python, cookies, FFmpeg ou engine aparecem como erros acionáveis na tela.

## Histórico, diagnóstico e atualizações

O histórico é um arquivo JSON local, versionado e limitado a 50 registros. A
home mostra somente os cinco mais recentes; limpar o histórico remove registros
e não remove arquivos. O diagnóstico informa versões e health resumidos, troca
paths de usuário por `~` e remove cookies/tokens antes de copiar. Logs locais
rotacionam em tamanho limitado. A verificação de atualização consulta apenas
releases estáveis oficiais do GitHub, no máximo uma vez por 24 horas, e nunca
instala ou interrompe um download.

## Arquitetura

```text
React/MUI
  -> bindings Wails gerados
App (lifecycle, dialogs, open actions)
  -> internal/app.Service
     -> auth.AuthManager + Chromium/CDP (perfil privado por provedor)
     -> auth.CookieStore (Netscape, validação, importação, disconnect)
     -> download.JobManager (fila FIFO, um job ativo, cancelamento, eventos)
     -> process.Runner (execução direta, stdout/stderr, process tree)
     -> engines (yt-dlp, gamdl, FFmpeg, Deno, manifesto/checksum)
     -> config + platform (JSON atômico, paths, dialogs auxiliares)
```

Os contratos públicos estão em `internal/models`. Os engines recebem argumentos separados; nenhum valor digitado pelo usuário passa por shell.

## Testes e CI

Os comandos principais são:

```bash
cd frontend && npm ci && npm run typecheck && npm test && npm run build
cd .. && go test ./... && go vet ./...
```

Os testes cobrem detecção de serviço, paths, configuração corrompida/atômica, builders de argumentos sem shell injection, metadata, progresso, checksum, stderr, erro, cancelamento, processos filhos e transições do job. Os workflows em `.github/workflows` executam a validação e builds smoke nativos; download real de mídia e cookies reais ficam na aceitação manual.

O registro da rodada atual, incluindo o smoke real de YouTube e as limitações não verificadas, está em [`docs/smoke-test-record.md`](docs/smoke-test-record.md).

## Dados locais e privacidade

Configuração, caches, engines, temporários e logs ficam nos diretórios de usuário resolvidos pelas APIs da plataforma. A pasta de downloads é a pasta Downloads do sistema quando disponível. A configuração é escrita atomicamente, com permissão restritiva onde suportado.

O Audivo faz apenas as requisições diretas necessárias aos engines selecionados. Não há serviço intermediário para receber URLs, mídia ou cookies.

## Limitações conhecidas

- A disponibilidade de vídeos, formatos, DRM, região e autenticação depende do YouTube, Apple Music e dos engines upstream.
- A aceitação real de Apple Music exige cookies válidos e não é reproduzível em CI.
- A aceitação real dos logins Apple Music e YouTube, incluindo MFA/CAPTCHA e captura em janela Chromium, exige contas descartáveis e ainda deve ser executada manualmente no runner correspondente.
- O fluxo web depende de Google Chrome, Microsoft Edge, Brave ou Chromium instalado; a matriz observada no host macOS está registrada em [`docs/provider-auth-browser-support.md`](docs/provider-auth-browser-support.md).
- A matriz de release só deve ser considerada publicada após o smoke build nativo do runner correspondente passar.
- Um build fonte sem manifesto de artefatos e sem engines no `PATH` exibirá os engines como ausentes até que a distribuição forneça o manifesto ou o usuário configure os executáveis.

## Licença

O código do Audivo é distribuído sob GPL-3.0; veja `LICENSE`. Avisos e fontes das dependências e engines distribuíveis estão em [`THIRD_PARTY_NOTICES.md`](THIRD_PARTY_NOTICES.md).

# Registro de validação — 2026-08-17 e 2026-09-10

Este registro separa evidência automatizada, smoke real de engine e aceitação manual ainda pendente. Os downloads reais foram gravados fora do repositório; as mídias dos smokes de 2026-09-10 foram movidas para a Lixeira após a validação.

## Verificado

- `frontend/npm ci`: concluído sem vulnerabilidades reportadas.
- `frontend/npm run typecheck`: concluído.
- `frontend/npm test`: 12 testes concluídos.
- `frontend/npm run build`: concluído.
- `go test ./...`: concluído.
- `go vet ./...`: concluído.
- `go test -race ./...`: concluído.
- `wails generate module`: concluído.
- `wails dev`: iniciado com frontend Vite, bindings e servidor dev; encerrado após o smoke.
- `wails build -clean`: concluído localmente para `darwin/arm64` com Wails v2.15.0.
- `openspec validate complete-audivo-intelligent-experience --strict`: válido.

## Smoke real do YouTube

Engine local usado: `yt-dlp 2026.07.04`, Deno 2.9.5 e FFmpeg 9.0.1.

- URL pública: `https://www.youtube.com/watch?v=jNQXAC9IVRw`.
- Metadata estruturada: passou; extractor `Youtube`, título `Me at the zoo`, duração 19 s.
- Áudio: passou; `-x --audio-format m4a`, com arquivo `.m4a` produzido pelo FFmpeg.
- Vídeo: passou; streams de vídeo/áudio separados baixados e merge concluído em `.mp4` pelo FFmpeg.
- Uma fixture anterior (`BaW_jenozKc`) retornou “Video unavailable”; ela foi substituída por uma URL pública disponível e isso não alterou o código.

## Smoke da tela local

- Vite local: tela única carregada e inspecionada no navegador in-app.
- Formulário, labels, diálogo de configurações e estado de erro foram encontrados no snapshot de acessibilidade.
- Entrada de URL e submissão por `Enter` funcionaram; sem a ponte Wails, a análise exibiu a mensagem amigável “Não foi possível analisar este link.”.
- A tentativa de colar sem a ponte nativa exibiu “Não foi possível ler a área de transferência.”, confirmando o caminho de falha sem expor erro técnico.
- A captura visual confirmou a hierarquia desktop e o estado de erro. A capacidade de alterar viewport não está disponível neste navegador, então a janela pequena continua pendente.

## Smoke nativo Wails

- A aplicação Wails foi aberta, encerrada e reaberta localmente com `wails dev`.
- URL inválida exibiu o erro amigável de serviço suportado; o painel nativo carregou destino, cookies, tema, formato e status dos engines.
- Engines detectados como prontos: yt-dlp `2026.07.04`, gamdl `3.8.5`, FFmpeg, ffprobe e Deno `2.9.5`.
- Preview YouTube real: `Me at the zoo`, autor `jawed`, duração `0:19`.
- Download nativo de áudio MP3 concluído em diretório temporário; o resultado foi `Me at the zoo.mp3`.
- Download nativo de vídeo concluído em diretório temporário; o resultado foi `Me at the zoo.mp4`, com merge de áudio/vídeo pelo FFmpeg.
- Nesta rodada, o build empacotado foi reaberto; o preview real do álbum Apple Music `Lover` exibiu 18 itens, e o cancelamento em `preparando` exibiu `Download cancelado` sem iniciar os itens seguintes.
- A execução terminou com o processo nativo encerrado após fechar a janela; as telas representativas estão em [`docs/screenshots`](screenshots/).
- O diálogo de seleção de pasta abriu e aceitou diretórios temporários; após o smoke, `/Users/gabrielpgava/Downloads` foi restaurado.
- A ação de abrir pasta mostrou o diretório e o arquivo no Finder; a ação de abrir arquivo foi invocada sem erro visível.
- A janela foi reduzida até o mínimo configurado, com rolagem funcional e controles sem sobreposição aparente.
- Tab percorreu configurações, URL, Colar e Analisar; `Return` no botão Analisar iniciou uma nova análise.
- Clipboard nativo: a URL foi alterada, copiada com `Cmd+C`, substituída e recuperada pelo botão Colar.
- Preferências de tema/formato e destino foram salvas e carregadas novamente após reinício do processo; o estado final foi restaurado para tema do sistema, MP3 e Downloads.
- Tentativa de cancelamento do vídeo curto: o arquivo terminou antes da ação Cancelar; este caso não é contado como cancelamento validado.
- Capturas visuais do fluxo nativo: [`docs/screenshots`](screenshots/), cobrindo idle, preview, downloading, coleção, completed e settings sem console técnico.

## Smoke Apple Music — 2026-09-10

- Sessões locais já importadas foram reconhecidas como `Conectado` para Apple Music e YouTube; o conteúdo dos cookies não foi registrado. O arquivo Netscape do Apple Music autenticou a API do gamdl e retornou assinatura ativa da storefront brasileira.
- Engines preparados e reconhecidos pelo app: yt-dlp `2026.07.04`, gamdl `3.8.5`, Deno `2.9.5`, FFmpeg e ffprobe.
- O primeiro smoke revelou dois defeitos reais e foi corrigido: o diretório temporário não era criado no startup, e o helper de metadata recebia logs de debug no stdout junto com o JSON. O helper passou a manter stdout exclusivamente JSON.
- Preview nativo de song: [`Brota no Meu Setor`](https://music.apple.com/br/song/brota-no-meu-setor-feat-mc-rodrigo-do-cn/6805642156), 2:23, com opção MP3.
- Download nativo da song concluído com conversão FFmpeg para MP3; 1 arquivo foi produzido e movido para a Lixeira após a validação.
- Preview nativo de album: [`Lover`](https://music.apple.com/us/album/lover/1468058165), 18 itens.
- Download nativo do album concluído; 18 MP3 foram encontrados na pasta de saída e a pasta foi movida para a Lixeira após a validação.
- Cancelamento nativo de um segundo job do album exibiu `Download cancelado`; nenhum arquivo parcial permaneceu no `Downloads`.

## Smoke YouTube — 2026-09-10

- URL pública: `https://www.youtube.com/watch?v=jNQXAC9IVRw`; preview `Me at the zoo`, autor `jawed`, duração 19 s.
- Um job de vídeo foi cancelado ainda em `preparing` e o app exibiu `Download cancelado`.
- Após o cancelamento, a mesma URL foi analisada novamente e um novo job concluiu em `Me at the zoo.mp4`; `ffprobe` confirmou streams de vídeo e áudio no arquivo merged. O arquivo foi movido para a Lixeira após a validação.
- Um processo separado com PATH sem os executáveis locais reproduziu o cenário de engine ausente e exibiu a mensagem amigável de preparação do `yt-dlp`.

## Ainda não verificado

- Abertura do arquivo no aplicativo associado: a ação foi invocada sem erro visível, mas a confirmação observável ficou limitada ao Finder; repetir com uma associação de mídia disponível é recomendado.
- A matriz de fixtures de destino, colisão, espaço e nomes passou no macOS nativo e nos pacotes `internal/download`/`internal/platform` em Linux via container; Windows x64 e macOS Intel ainda não foram executados.
- Smoke builds dos runners Windows x64, macOS Intel e Linux x64: definidos no CI/release, não executados neste Mac Apple Silicon.
- Distribuição final: o manifesto de exemplo contém placeholder de checksum; um release precisa fornecer o manifesto completo por plataforma com URLs, versões e SHA-256 reais.

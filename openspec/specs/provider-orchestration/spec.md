# Provider Orchestration

## Purpose

Audivo routes one normalized media input through internal YouTube and Apple Music providers and exposes typed previews, health, and progress to the application.

## Requirements

### Requirement: O sistema SHALL normalizar e analisar uma entrada única de mídia

O backend SHALL processar qualquer entrada textual por uma sequência única de normalização, validação de URL, detecção de provider e análise, retornando um contrato de mídia que identifique `single` ou `collection` quando suportado.

#### Scenario: URL de vídeo do YouTube
- **WHEN** o usuário fornece uma URL válida de vídeo do YouTube
- **THEN** o sistema identifica o provider YouTube, classifica a mídia como `single` e retorna título, tipo e opções disponíveis sem iniciar o download

#### Scenario: URL de playlist do YouTube
- **WHEN** o usuário fornece uma URL válida de playlist do YouTube
- **THEN** o sistema classifica a mídia como `collection`, retorna metadados resumidos e a contagem/itens disponíveis sem iniciar os downloads

#### Scenario: Entrada inválida ou não suportada
- **WHEN** a entrada não é uma URL compatível com nenhum provider registrado
- **THEN** o sistema retorna um erro acionável e não cria job nem executa engine

### Requirement: O registry SHALL encapsular os providers iniciais

O sistema SHALL manter um registry interno com providers de YouTube via yt-dlp e Apple Music via gamdl. A camada de UI e o manager SHALL depender apenas da interface de provider, sem condicionais espalhadas ou configuração de plugins externos.

#### Scenario: Detecção de provider
- **WHEN** o registry recebe uma URL suportada por um provider
- **THEN** ele devolve exatamente um provider responsável pela análise e download

#### Scenario: URL sem provider
- **WHEN** nenhum provider aceita a URL
- **THEN** o registry retorna resultado de não suporte sem tentar cada engine de forma destrutiva

### Requirement: Cada provider SHALL expor health check leve e resultado estruturado

Cada provider SHALL informar disponibilidade, dependências ausentes e mensagem segura, e SHALL traduzir a execução da engine para `MediaInfo`, `DownloadResult` e eventos de progresso sem expor flags ou stdout bruto como contrato da aplicação.

#### Scenario: Dependências disponíveis
- **WHEN** o health check de um provider encontra sua engine e dependências mínimas
- **THEN** retorna `Available=true` e a UI pode exibir o provider como pronto

#### Scenario: Dependência ausente
- **WHEN** a engine necessária não está disponível
- **THEN** o provider retorna `Available=false` com instrução acionável e nenhuma execução de download é iniciada

### Requirement: A aplicação SHALL emitir progresso de download estruturado

O backend SHALL emitir eventos com job, estado/etapa, item atual, total, percentual do item, percentual geral, velocidade e ETA quando disponíveis. O frontend SHALL poder renderizar o progresso sem interpretar texto de CLI.

#### Scenario: Download com progresso conhecido
- **WHEN** a engine informa percentual ou velocidade do item atual
- **THEN** o evento inclui os valores conhecidos e o progresso geral correspondente

#### Scenario: Engine sem ETA
- **WHEN** a engine não fornece ETA ou velocidade confiável
- **THEN** os campos ficam ausentes/indeterminados sem inventar valores e o estado continua atualizável

# Collection Downloads

## Purpose

Audivo previews collections and processes their items predictably, with observable progress, safe cancellation, and resumable partial results.

## Requirements

### Requirement: Coleções SHALL apresentar preview resumido antes da fila

O sistema SHALL mostrar nome, provider, quantidade e, quando disponível, autor/álbum de uma playlist ou álbum antes de iniciar o download. A lista detalhada SHALL ser opcional e não SHALL exigir seleção item a item para o fluxo padrão.

#### Scenario: Playlist analisada
- **WHEN** a análise identifica uma playlist com 42 itens
- **THEN** a UI mostra a playlist e a contagem com uma ação principal para baixar todos, sem renderizar 42 controles obrigatórios

#### Scenario: Usuário abre detalhes
- **WHEN** o usuário escolhe ver os itens de uma coleção
- **THEN** a UI revela os itens como informação/seleção secundária sem alterar a regra de fila controlada

### Requirement: O DownloadManager SHALL processar uma coleção em fila previsível

O manager SHALL representar uma coleção como job pai com subjobs, processar inicialmente um download completo por vez em ordem estável e preservar a capacidade de usar paralelismo interno da engine quando aplicável.

#### Scenario: Coleção iniciada
- **WHEN** o usuário confirma o download de uma coleção
- **THEN** o primeiro item entra em `preparing/downloading` e os demais permanecem aguardando sem iniciar processos completos concorrentes

#### Scenario: Próximo item
- **WHEN** um item termina com sucesso
- **THEN** o manager marca o item como concluído e inicia o próximo item elegível da fila

### Requirement: O sistema SHALL expor progresso do item e da coleção

A UI SHALL mostrar item atual, índice/total e progresso geral, além do progresso do item, sem misturar percentuais ou apresentar como concluído o que ainda está aguardando.

#### Scenario: Coleção em andamento
- **WHEN** o item 17 de 42 está sendo baixado a 86%
- **THEN** a UI mostra o item atual e uma progressão geral coerente com 17/42 e o percentual do item

#### Scenario: Coleção sem progresso numérico
- **WHEN** o provider não consegue calcular percentual do item
- **THEN** a UI mostra estado de atividade e contagem da coleção sem exibir um percentual falso

### Requirement: Cancelamento SHALL preservar o que já foi concluído

Ao cancelar uma coleção, o sistema SHALL interromper o item atual, SHALL impedir o início de novos itens, SHALL preservar arquivos válidos concluídos e SHALL limpar temporários somente quando for seguro.

#### Scenario: Cancelamento durante item 2
- **WHEN** o item 1 foi concluído e o usuário cancela enquanto o item 2 está ativo
- **THEN** o item 1 permanece, o item 2 é cancelado e o item 3 não inicia

#### Scenario: Cancelamento após parte da coleção
- **WHEN** 17 de 42 itens foram concluídos antes do cancelamento
- **THEN** o job termina como cancelado/parcial com os 17 itens preservados e sem rebaixá-los automaticamente numa retomada posterior

### Requirement: Falha parcial SHALL ser recuperável sem rebaixar itens concluídos

O sistema SHALL registrar quais subjobs concluíram com segurança e SHALL permitir retry/retomada da coleção somente dos itens pendentes ou falhos quando conseguir identificá-los.

#### Scenario: Quatro sucessos e uma falha
- **WHEN** uma fixture de cinco itens produz quatro sucessos e um erro definitivo
- **THEN** o resultado da coleção é `collection-partial`, informa o item falho e mantém os quatro arquivos

#### Scenario: Retomada de coleção parcial
- **WHEN** o usuário solicita novamente uma coleção com itens previamente concluídos identificáveis
- **THEN** o manager pula os itens concluídos e processa somente os demais

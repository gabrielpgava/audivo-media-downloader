## ADDED Requirements

### Requirement: Configurações SHALL expor health e diagnóstico compactos

O sistema SHALL apresentar, sob demanda, versão do Audivo, OS, arquitetura, estado de yt-dlp, FFmpeg, gamdl e provider, sem executar uma auditoria pesada em toda abertura. Health checks completos SHALL ocorrer ao usar o provider ou por ação explícita.

#### Scenario: Diagnóstico solicitado
- **WHEN** o usuário abre `Configurações > Diagnóstico`
- **THEN** a UI exibe um relatório atualizado com os componentes detectáveis e seus estados

#### Scenario: Startup normal
- **WHEN** o Audivo é aberto
- **THEN** ele executa somente verificações leves e não bloqueia a tela inicial aguardando todas as engines

### Requirement: O relatório copiável SHALL ser útil e sanitizado

O sistema SHALL gerar texto copiável com versão, sistema, arquitetura, provider/operação/resultado e erro resumido, removendo cookies, tokens, headers, conteúdo de credenciais e paths sensíveis desnecessários.

#### Scenario: Falha com path sensível
- **WHEN** uma operação falha sob `/Users/gabriel/Downloads`
- **THEN** o relatório usa uma forma sanitizada como `~/Downloads` e não inclui conteúdo de cookies ou tokens

#### Scenario: Copiar diagnóstico
- **WHEN** o usuário clica em `Copiar diagnóstico`
- **THEN** o texto é colocado no clipboard e está pronto para uma issue sem envio automático

### Requirement: Logs SHALL ser locais, pequenos e rotativos

O backend SHALL manter logs de debugging com rotação e limite de tamanho, sem armazenar mídia, cookies ou crescimento ilimitado. Logs SHALL ser usados para detalhes técnicos que não pertencem à mensagem principal.

#### Scenario: Rotação de log
- **WHEN** o arquivo de log atinge o limite configurado
- **THEN** os arquivos anteriores são rotacionados dentro de um número limitado e o log atual continua gravável

#### Scenario: Evento com credencial
- **WHEN** uma engine imprime dados potencialmente sensíveis
- **THEN** o logger redige o valor antes de persistir ou exibir no diagnóstico

### Requirement: Erros SHALL oferecer ações externas somente por decisão do usuário

Em erros relevantes, a UI SHALL oferecer `Copiar diagnóstico` e, quando configurado, `Abrir GitHub Issues`, abrindo o endereço oficial sem criar envio automático.

#### Scenario: Usuário abre issue
- **WHEN** o usuário escolhe `Abrir GitHub Issues`
- **THEN** o Audivo abre o navegador no destino oficial e não transmite o relatório sem confirmação adicional

### Requirement: Notificações nativas SHALL ser discretas e opcionais

Quando houver suporte confiável da plataforma, o sistema SHALL notificar apenas conclusão/erro relevante enquanto o Audivo estiver em background, sem notificar análise, início, conversão ou cada item individual.

#### Scenario: Coleção concluída em background
- **WHEN** uma coleção termina enquanto a janela está em background
- **THEN** o usuário recebe no máximo uma notificação resumida da conclusão da coleção

#### Scenario: Download ativo em foreground
- **WHEN** o download termina com a janela em primeiro plano
- **THEN** nenhuma sequência redundante de notificações é emitida

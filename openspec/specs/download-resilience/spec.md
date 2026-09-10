# Download Resilience

## Purpose

Audivo classifies failures, retries recoverable operations within bounded limits, communicates recovery clearly, and keeps job state consistent.

## Requirements

### Requirement: O sistema SHALL classificar falhas transitórias e definitivas

Falhas de timeout, conexão interrompida, fragmento ou indisponibilidade temporária SHALL ser elegíveis a retry limitado. URL inválida, conteúdo removido/privado sem autenticação, cookies inválidos, indisponibilidade regional e formato inexistente SHALL retornar erro definitivo sem repetição automática.

#### Scenario: Timeout transitório
- **WHEN** uma engine retorna timeout ou conexão interrompida
- **THEN** o resultado é classificado como recuperável e o job pode entrar em retry

#### Scenario: URL inválida
- **WHEN** a engine ou a validação identifica URL inválida
- **THEN** o job falha imediatamente com mensagem clara e não consome tentativas de retry

### Requirement: Retry SHALL usar a capacidade nativa do provider antes de lógica duplicada

Cada provider SHALL preferir as opções nativas e confiáveis de yt-dlp/gamdl para retry/resume, com limite padrão de até três tentativas automáticas para falhas recuperáveis. O manager SHALL impedir loops indefinidos.

#### Scenario: Retry bem-sucedido
- **WHEN** a primeira tentativa falha por conexão e a segunda tentativa conclui
- **THEN** o job termina como concluído e o histórico não registra a falha transitória como erro final

#### Scenario: Tentativas esgotadas
- **WHEN** todas as tentativas permitidas falham por condição recuperável
- **THEN** o job termina como erro acionável e não inicia outra tentativa automaticamente

### Requirement: A UI SHALL comunicar recuperação sem expor detalhes de engine

Durante retry, o backend SHALL emitir estado `retrying` com tentativa atual/limite quando disponível, e a UI SHALL mostrar mensagens como `Reconectando` ou `Tentando novamente`, sem exigir que o usuário interprete logs CLI.

#### Scenario: Coleção em retry
- **WHEN** um subjob de coleção entra em retry
- **THEN** a UI identifica o item atual e a tentativa sem perder o progresso geral da coleção

### Requirement: Erros conhecidos SHALL produzir orientação acionável

O sistema SHALL traduzir erros de conteúdo privado, cookies expirados, engine desatualizada, dependência ausente e espaço insuficiente para mensagens claras, mantendo detalhes técnicos somente no diagnóstico/log sanitizado.

#### Scenario: Cookies expirados
- **WHEN** o provider Apple Music detecta sessão/cookies expirados
- **THEN** a UI informa que a sessão expirou e oferece a ação de atualizar cookies, sem exibir os cookies

#### Scenario: Engine desatualizada
- **WHEN** uma execução falha com sinal conhecido de incompatibilidade da engine
- **THEN** a UI informa que o mecanismo pode estar desatualizado e oferece verificação/atualização apropriada

### Requirement: O estado SHALL permanecer consistente durante erro e cancelamento

O job SHALL usar estados enumerados e SHALL garantir que retry, erro, cancelamento e conclusão sejam mutuamente consistentes, sem deixar a UI presa em `downloading` após o processo terminar.

#### Scenario: Cancelamento durante retry
- **WHEN** o usuário cancela um job enquanto ele aguarda nova tentativa
- **THEN** o job termina como `cancelled`, nenhuma tentativa futura inicia e o cancelamento não é convertido em erro

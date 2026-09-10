# Local History

## Purpose

Audivo keeps a small local history of known downloads to support recent items, duplicate warnings, and explicit re-downloads without indexing the computer or syncing data.

## Requirements

### Requirement: O histórico SHALL persistir somente downloads conhecidos e permanecer limitado

O backend SHALL armazenar localmente registros versionados dos downloads concluídos ou parciais relevantes, com URL normalizada, provider, título, tipo, path/destino e timestamps. O repositório SHALL manter no máximo 50 entradas ou limite equivalente definido pelo produto e SHALL descartar as mais antigas primeiro.

#### Scenario: Download concluído
- **WHEN** um download termina com sucesso
- **THEN** um registro é persistido localmente com os dados necessários para recentes, duplicidade e abertura do arquivo

#### Scenario: Limite atingido
- **WHEN** uma nova entrada faria o histórico ultrapassar o limite configurado
- **THEN** as entradas mais antigas são removidas do repositório sem tocar nos arquivos de mídia

### Requirement: A home SHALL mostrar recentes de forma discreta

O sistema SHALL expor no máximo alguns registros recentes, por padrão até cinco, com título, resultado e ação relevante como abrir arquivo ou pasta. Recentes SHALL permanecer secundários ao campo de entrada e não SHALL virar uma biblioteca indexada.

#### Scenario: Home com histórico
- **WHEN** existem downloads recentes
- **THEN** a home mostra uma seção compacta com ações de abertura sem substituir o fluxo principal de colar e baixar

#### Scenario: Home sem histórico
- **WHEN** nenhum download foi registrado
- **THEN** a home não exibe uma lista vazia dominante e mantém o estado de entrada simples

### Requirement: O sistema SHALL alertar duplicidade sem bloquear a decisão do usuário

Antes de enfileirar uma URL já concluída, o backend SHALL consultar o histórico pela URL normalizada e informar data/path se o arquivo ainda existir. O usuário SHALL poder abrir o arquivo ou escolher baixar novamente.

#### Scenario: URL já baixada e arquivo presente
- **WHEN** o usuário analisa uma URL registrada cujo arquivo ainda existe
- **THEN** a UI informa que o conteúdo já foi baixado e oferece abrir ou baixar novamente

#### Scenario: Usuário escolhe baixar novamente
- **WHEN** o usuário confirma explicitamente um novo download apesar do aviso
- **THEN** o manager cria o novo job respeitando a política de colisão

### Requirement: O histórico SHALL detectar arquivo removido sem escanear o computador

Ao tentar abrir um registro, o sistema SHALL verificar somente o path armazenado. Se o arquivo não existir, SHALL marcar o registro como ausente e oferecer download novamente, sem indexar outros diretórios.

#### Scenario: Arquivo removido manualmente
- **WHEN** o path registrado não existe mais
- **THEN** a UI mostra `Arquivo não encontrado` e oferece baixar novamente

#### Scenario: Arquivo registrado existe
- **WHEN** o path registrado ainda existe
- **THEN** a ação abrir arquivo/pasta é disponibilizada sem uma varredura adicional

### Requirement: O usuário SHALL poder limpar o histórico e os dados SHALL permanecer locais

Configurações SHALL oferecer limpeza explícita do histórico. O sistema SHALL não sincronizar, enviar para analytics/GitHub ou exigir conta, e SHALL manter cookies fora do histórico.

#### Scenario: Limpeza confirmada
- **WHEN** o usuário confirma `Limpar histórico`
- **THEN** os registros locais são removidos, os arquivos de mídia permanecem intactos e a home volta ao estado sem recentes

#### Scenario: Uso offline
- **WHEN** o Audivo opera sem serviço de sincronização configurado
- **THEN** histórico e preferências continuam disponíveis localmente e nenhuma chamada de analytics é criada

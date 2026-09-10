## ADDED Requirements

### Requirement: O sistema SHALL gerar paths e nomes seguros multiplataforma

Uma camada central SHALL sanitizar componentes derivados de título, artista e álbum para Windows, macOS e Linux, removendo separadores/caracteres incompatíveis, tratando nomes reservados do Windows, limitando comprimento e preservando Unicode válido quando possível.

#### Scenario: Título com caracteres reservados
- **WHEN** o título contém `AC/DC`, `What's Up?`, `Artist: Song` ou `<Official Video>`
- **THEN** o path gerado não contém componentes inválidos e permanece legível

#### Scenario: Nome reservado no Windows
- **WHEN** o título sanitizado seria `CON` ou outro nome reservado
- **THEN** o serviço altera o componente de modo determinístico para evitar falha no Windows

#### Scenario: Path excessivamente longo
- **WHEN** a combinação de diretório e nome excede o limite seguro da plataforma
- **THEN** o serviço encurta componentes de forma determinística sem perder a extensão nem gerar path inválido

### Requirement: Downloads SHALL ser organizados sem sobrescrever arquivos existentes

O destino padrão SHALL usar uma raiz previsível do Audivo e subdiretórios de coleção/álbum quando apropriado. A resolução de colisões SHALL preservar o arquivo existente e gerar sufixo ou reconhecer a mesma operação sem apagar conteúdo.

#### Scenario: Primeira coleção
- **WHEN** o usuário baixa uma playlist chamada `Summer 2026`
- **THEN** os itens são salvos sob um diretório próprio e previsível da coleção

#### Scenario: Colisão de nome
- **WHEN** `Musica.mp3` já existe no destino
- **THEN** o novo download não sobrescreve silenciosamente e recebe um nome seguro como `Musica (2).mp3` ou é identificado como o mesmo arquivo

### Requirement: O sistema SHALL verificar espaço disponível antes de operações grandes quando houver estimativa

Quando o provider fornecer tamanho estimado, o backend SHALL comparar o tamanho com o espaço disponível e incluir margem para temporários, muxing e conversão. Sem estimativa confiável, SHALL continuar sem inventar bloqueio.

#### Scenario: Espaço insuficiente
- **WHEN** a estimativa com margem excede o espaço livre do destino
- **THEN** o download não inicia e a UI informa os valores conhecidos e oferece escolha de outra pasta

#### Scenario: Espaço suficiente
- **WHEN** há espaço livre acima da estimativa com margem
- **THEN** o job pode entrar na fila e a verificação não bloqueia o usuário

### Requirement: Metadata e artwork disponíveis SHALL ser preservados sem scraping adicional

Os providers SHALL solicitar título, artista, álbum, número da faixa, ano e capa quando a engine oferecer suporte confiável, sem criar uma etapa de scraping independente apenas para completar metadata.

#### Scenario: Áudio com metadata
- **WHEN** gamdl/yt-dlp retorna metadata e artwork válidos para um download de áudio
- **THEN** o resultado final inclui essas informações conforme suportado pelo formato

#### Scenario: Metadata indisponível
- **WHEN** o provider não fornece um campo opcional
- **THEN** o download continua com os dados disponíveis e não falha apenas por ausência de metadata não essencial

### Requirement: Resume SHALL respeitar o estado seguro dos temporários

O adapter SHALL preservar a possibilidade de resume quando a engine suportar isso com segurança, não apagando temporários em falhas recuperáveis. Cancelamento manual SHALL limpar somente temporários que não possam ser retomados sem risco.

#### Scenario: Falha transitória após arquivo parcial
- **WHEN** a conexão cai depois de criar um arquivo parcial retomável
- **THEN** o temporário permanece disponível para a tentativa seguinte da engine

#### Scenario: Download concluído
- **WHEN** o processamento finaliza com sucesso
- **THEN** o resultado aponta para o arquivo final e não deixa temporário ativo como se fosse download pendente

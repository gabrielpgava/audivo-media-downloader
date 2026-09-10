## ADDED Requirements

### Requirement: O Audivo SHALL verificar releases estáveis de forma periódica

O sistema SHALL consultar a fonte oficial de GitHub Releases no máximo uma vez por janela de periodicidade, por padrão 24 horas, em operação assíncrona e sem bloquear startup, análise ou download. Releases alpha, beta, rc e nightly SHALL ser ignoradas por padrão.

#### Scenario: Release estável mais nova
- **WHEN** a verificação encontra uma release estável maior que a versão instalada
- **THEN** o sistema registra a atualização disponível e pode exibir um aviso não intrusivo

#### Scenario: Somente pre-release
- **WHEN** a API retorna apenas versões alpha, beta, rc ou nightly
- **THEN** nenhuma atualização é apresentada ao usuário padrão

#### Scenario: Verificação já realizada
- **WHEN** a janela de periodicidade ainda não expirou
- **THEN** o Audivo usa o resultado local ou não repete a requisição a cada abertura de tela

### Requirement: O aviso de atualização SHALL abrir a release oficial sem auto-update silencioso

O aviso SHALL informar que há nova versão e oferecer abertura da página oficial. O Audivo SHALL nunca baixar ou instalar uma atualização silenciosamente neste ciclo e SHALL nunca interromper download ativo.

#### Scenario: Usuário aceita ver atualização
- **WHEN** o usuário clica em `Ver atualização`
- **THEN** o navegador abre a página oficial da release

#### Scenario: Download ativo
- **WHEN** uma atualização é detectada durante um download
- **THEN** o aviso permanece secundário e o job de download continua sem interrupção

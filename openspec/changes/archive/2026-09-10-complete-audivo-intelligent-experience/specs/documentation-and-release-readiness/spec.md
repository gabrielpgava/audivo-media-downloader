## ADDED Requirements

### Requirement: O README SHALL documentar o produto real e seus limites

O README SHALL explicar downloads individuais, playlists/álbuns suportados, histórico, duplicidade, retry, diagnóstico, gerenciamento/verificação de engines, execução local e limitações conhecidas, sem prometer providers ou comportamentos fora do escopo.

#### Scenario: Novo colaborador lê o README
- **WHEN** uma pessoa segue o README em um checkout limpo
- **THEN** encontra pré-requisitos, comando de execução, testes e descrição da arquitetura de providers

#### Scenario: Usuário procura suporte
- **WHEN** o usuário consulta os tipos de mídia suportados
- **THEN** o README diferencia YouTube/Apple Music e não lista providers futuros como disponíveis

### Requirement: A documentação visual SHALL cobrir somente estados essenciais

O projeto SHALL incluir screenshots atualizados de idle, preview, download, coleção, conclusão e configurações, em quantidade suficiente para explicar o fluxo sem poluir a documentação.

#### Scenario: Fluxo principal documentado
- **WHEN** alguém consulta os screenshots
- **THEN** consegue reconhecer a entrada, preview, progresso e conclusão sem precisar ver telas técnicas de log

### Requirement: O projeto SHALL possuir contribuição e issue template seguros

`CONTRIBUTING.md` SHALL explicar rodar, testar, arquitetura de providers, issue e PR. O template de bug SHALL pedir versão, OS, arquitetura, provider, tipo de URL, passos e diagnóstico, e SHALL alertar para nunca colar cookies ou credenciais.

#### Scenario: Abertura de bug
- **WHEN** um usuário cria uma issue pelo template
- **THEN** recebe campos para reproduzir e anexar diagnóstico sanitizado, além do aviso explícito contra cookies/credenciais

#### Scenario: Pull request
- **WHEN** um colaborador prepara uma mudança
- **THEN** encontra no guia as verificações esperadas e a orientação para manter providers internos e a UX simples

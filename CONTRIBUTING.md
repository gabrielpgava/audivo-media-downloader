# Contribuindo

## Fluxo local

1. Crie uma branch curta a partir da versão atual.
2. Rode `cd frontend && npm ci && npm run typecheck && npm test && npm run build`.
3. Rode `go test ./...`, `go vet ./...` e `git diff --check`.
4. Descreva no PR o provider, engine e cenário validado; mantenha smoke real e credenciais fora do CI.

## Arquitetura

O frontend fala apenas com bindings Wails e DTOs de `internal/models`. A
detecção fica centralizada em `internal/download`; adapters de yt-dlp e gamdl
montam argumentos sem shell. Cookies são armazenados fora do repositório e
nunca devem aparecer em logs, issues, fixtures ou diagnósticos.

Mudanças que alterem um contrato de análise, job, evento ou configuração devem
atualizar os testes e os bindings gerados com `wails generate module`.

## Issues e PRs

Use o template de bug quando aplicável. Não anexe cookies, tokens, headers,
perfis de navegador, URLs privadas ou arquivos de mídia. Para uma alteração
maior, inclua limites conhecidos e o que não foi possível validar sem uma
conta/engine real.

# Workflow

## Conventional Commits

Format:
```
<type>(<scope>): <short description>

[optional body]

[optional footer(s)]
```

Types: `feat` · `fix` · `refactor` · `test` · `chore` · `docs` · `perf` · `style` · `revert`

Scopes (used in this repo): `customer` · `vehicle` · `auth` · `bootstrap` · `internal` · `infra` · `db` · `api` · `integration` · `web` · `middleware` · `clients` · `swagger` · `seed` · `k8s` · `ci`

Rules:
- Subject max 72 chars, lowercase, no trailing period.
- Imperative mood: "add feature" not "added feature".
- Breaking changes: `feat(auth)!: ...` or `BREAKING CHANGE:` footer.
- Reference issues in footer: `Closes #123`.

## Pre-commit Hooks
Configured in `.pre-commit-config.yaml`:
- `trailing-whitespace`, `end-of-file-fixer`, `check-yaml`, `check-added-large-files`, `check-merge-conflict`
- `golangci-lint` (v2.9.0)
- `go fmt ./...`
- `go build`
- `go-unit-tests`
- `go vet ./...`
- `check-ai-rules`: enforces `.ai/` as the only source of truth for rule files

## Development Commands
```bash
make run          # start server (PORT, default 8081; needs a reachable Postgres)
make test         # all tests with race detector (LOG_LEVEL=PANIC)
make coverage     # coverage.out + HTML report
make up           # docker compose up --build (app + PostgreSQL + migrate)
make down         # docker compose down -v
make migrate-up   # apply pending migrations (needs the golang-migrate CLI)
make mockgen      # regenerate all mocks via go generate
```

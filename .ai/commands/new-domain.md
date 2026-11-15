# Playbook: Scaffold New Domain

Use this playbook when adding a fully new domain to this service (Customer & Identity). Other business domains (work, supply, service orders...) belong to other services: do not add them here.

Model the new folder after `internal/vehicle/` (smallest domain with repository, service, controller and internal route).

## 1. Domain Model — `internal/<name>/domain/<name>.go`
- Define the struct with a `VARCHAR(36)` UUID `ID` and `CreatedAt`/`UpdatedAt` timestamps.
- Add `Validate() error` and sentinel errors in `<name>_errors.go` (`var ErrXxx = errors.New(...)`).
- Add the error to `validationErrors` / `notFoundErrors` / `conflictDomainErrors` in `internal/pkg/web/errors.go` if the API must map it.

## 2. Ports — `internal/<name>/interfaces/`
- `<name>_repository.go`: `<Name>Repository` interface (CRUD).
- `<name>_service.go` and `<name>_controller.go`: service and HTTP controller interfaces.
- Add `//go:generate go run github.com/vektra/mockery/v2@latest --name=<Name>Repository --with-expecter` on each interface file.

## 3. Repositories — `internal/<name>/repository/`
- `<name>_in_memory.go`: `NewInMemory()`, backed by a map. Used only by unit tests.
- `<name>_postgres.go`: `NewPostgres()`, raw SQL, no ORM. Connection comes from the context:
  `uow.GetTransaction(ctx)` (mutating, inside a UoW step) or `uow.GetOneTimeTransaction(ctx)` (read-only).
- Dynamic filters: `db.QueryBuilder(base)` from `internal/pkg/db`.

## 4. Service — `internal/<name>/service.go`
- Constructor: `NewService(uow uow.Executor, repo interfaces.<Name>Repository, ...) interfaces.<Name>Service`, wired in `di.go`.
- All business logic here; wrap mutations in `uow.Execute(ctx, step...)`. No SQL.

## 5. Controller — `internal/<name>/controller.go`
- Parse and validate HTTP input (`*http.Request`); call the service; return the response DTO from `adapters/`.
- Zero business logic.

## 6. Wire — `di.go` and routing
- `di.go`: `repo → service → controller`; expose `NewHTTPController(db)`.
- Add the field to `HandlersWrapper` in `internal/app/container/handlers.go` and the constructor in `httpContainer` (`internal/app/bootstrap/bootstrap.go`).
- Register routes in `internal/app/routing/routing.go` with `middleware.Auth(<role group>...)` and update `docs/swagger.yaml` and `.ai/rules/api.md`.

## 7. Migration
Follow `add-migration.md`. Add table, indexes, and FK constraints.

## 8. Tests
- Unit: service with in-memory repo + `uow.NewInMemoryUoW()`; controller with the mocks from `interfaces/mocks/`.
- Repository: `*_postgres_test.go` with `go-sqlmock`.
- HTTP flow: `internal/integration/` (build tag `integration`).
- Coverage: the CI gate is 80% without mocks.

## 9. Regenerate Mocks
```bash
make mockgen
```

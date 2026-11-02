# Architecture

## Layer Map

```
HTTP → Routing (internal/app/routing) → Controller (internal/<domain>/controller*.go)
     → Service (internal/<domain>/service*.go) → Repository (internal/<domain>/repository/*) → PostgreSQL
                                ↕
                      Domain (internal/<domain>/domain/*)
```

`<domain>` is `auth`, `customer` or `vehicle`. Each domain folder has `adapters/` (DTOs), `interfaces/` (ports, mocks in `interfaces/mocks/`), `domain/` (entities and errors), `repository/` and `di.go`. Shared infrastructure lives in `internal/pkg/*`; `internal/clients/serviceorder` is the HTTP client for the Service Order service.

## Layer Contracts
- **Controller** (the HTTP handler): parse and validate HTTP input, call the service, return the response — zero business logic. Gin types stay in `internal/app/`; controllers receive `*http.Request`.
- **Service**: all business logic (customer + user creation in one unit of work, role checks, document validation, service-order checks before delete).
- **Repository**: all SQL via `database/sql` + `lib/pq` — no queries outside repositories.
- **Domain**: structs + validation methods + sentinel errors. No Gin, no `lib/pq`.

## DI Root
- `internal/app/bootstrap/bootstrap.go` — builds the router, middlewares and HTTP server, and calls `httpContainer(db)`.
- `internal/<domain>/di.go` — constructors (`NewHTTPController(db)`, `NewInternalHTTPController(db)`, `NewService(...)`).
- `internal/app/container/handlers.go` — `HandlersWrapper` (controllers) and `Middlewares`.

## Repository DB Connection Pattern
Repositories hold no `*sql.DB`. The connection comes from the context, set by the UoW (`internal/pkg/uow`):
- **Mutating ops** (inside a UoW step): `uow.GetTransaction(ctx)` → `*sql.Tx` stored by the `OnStart` hook.
- **Read-only ops**: `uow.GetOneTimeTransaction(ctx)` → returns the shared `*sql.DB` pool (`app.DBConnect()`).

## Unit of Work
```go
uow.Execute(ctx, step1, step2, ...)
```
`uow.NewTransactionalUoW(db)` begins a transaction, injects it into the context, runs all steps, then commits or rolls back. `uow.NewInMemoryUoW()` is the test double. Services wrap all mutating repository calls in `uow.Execute`.

## Dynamic SQL
Use `db.QueryBuilder(baseQuery)` from `internal/pkg/db/`:
```go
qb := db.QueryBuilder(base)
qb.Add("column =", value)
qb.OrderBy(field, db.ASC)
qb.AddPagination(limit, offset)
query, args := qb.Build()
```
Used by `internal/vehicle/repository` (list and count with filters).

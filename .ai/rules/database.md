# Database

Database `customer` (PostgreSQL 15). Tables: `user`, `role`, `user_role`, `customer`, `vehicle`.
No foreign keys to other services' databases; no messaging tables (this service publishes and consumes nothing).

## PostgreSQL 15 Conventions
- IDs: `VARCHAR(36)` UUID.
- Timestamps: `TIMESTAMPTZ`, UTC, ISO 8601.
- Documents: `customer.cpf` / `customer.cnpj` are always stored without formatting (CPF 11 digits, CNPJ 14 characters), by the API and by the seed. See "Documentos" in `README.md`.

## Migration Strategy
All schema changes via versioned migration files only:
```
migrations/NNNNNN_description.up.sql
migrations/NNNNNN_description.down.sql
```
Never alter the schema manually. Never drop or modify existing columns — add new, deprecate old.

Commands:
```bash
make migrate-up      # apply pending
make migrate-down    # roll back last
make migrate-status  # current version
```

## Transaction Context Pattern
- `postgres.GetTransaction(ctx)` — gets `*sql.Tx` for mutating ops (called inside a UoW step).
- `postgres.GetOneTimeTransaction(ctx)` — gets a `*sql.DB` connection for read-only ops.

See `architecture.md` for the full UoW pattern.

## Required Indexes
`customer.user_id`, `vehicle.customer_id`, `user_role.user_id`, `user_role.role_id`, and the unique columns `user.email`, `customer.cpf`, `customer.cnpj`, `vehicle.license_plate`.

## Required Foreign Keys
`vehicle→customer`, `customer→user`, `user_role→user`, `user_role→role`.

## Read-only role for the login lambda
`lambda_auth_ro` only has `SELECT` on this database (see `README.md`, section "Login lambda").

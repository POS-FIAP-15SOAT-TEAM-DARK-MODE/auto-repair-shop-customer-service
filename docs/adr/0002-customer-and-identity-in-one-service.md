# ADR 0002 — Customer and Identity in one service

- Status: Accepted
- Date: 2026-11-01
- Deciders: Team DARK MODE

## Context

The monolith keeps users, roles, authentication, customers and vehicles together. The system is being split into services (see `docs/fase4/SPEC.md` of the infrastructure repository). Two constraints pull the identity data in two directions:

1. **Transactional coupling.** Creating a customer creates its login user (role `CUSTOMER`, default password = document) and deleting a customer deletes that user. Both happen in one unit of work (`uow`). Splitting users from customers would turn this into a distributed operation without a saga, for a flow that is local to one aggregate.
2. **Login lambda.** The login lambda authenticates against the database with two literal queries over `customer`, `user_role` and `role`. Keeping those tables in the same database, under a read-only role (`lambda_auth_ro`), avoids a second source of truth for credentials and keeps the lambda code unchanged.

## Decision

`customer-service` (port 8081, database `customer`) owns:

- `user`, `role`, `user_role`, authentication (`POST /v1/auth/login`, JWT issuance and validation) and user administration;
- `customer` and `vehicle`.

Other services reach this data through `/internal/customers/{id}` and `/internal/vehicles/{id}`, protected by `X-Internal-Token`, never routed by the gateway. The login lambda reads the database as `lambda_auth_ro` (`SELECT` only).

Service order, work, supply and stock live in other services and keep no copy of these tables; they store snapshots (name, document, unit price) at the time of the operation.

## Consequences

- Customer deletion no longer checks for service orders (they are in another database). Whether a customer with open service orders may be deleted is now a rule the caller must enforce, or a future internal query must answer. Reintroduced in F6 through the internal REST of Service Order (`TODO(F6)` in the code). Until then, deleting a customer with open orders is not blocked by this service; see README, "Limitações conhecidas".
- Customer and user changes stay atomic without distributed transactions.
- The lambda compatibility is covered by an integration test (`internal/integration/lambda_compat_test.go`) that runs the two literal queries as `lambda_auth_ro` and checks that writes are denied.
- The `JWT_SECRET` is shared by every service that validates tokens; rotation requires a coordinated change (same value in all secrets).

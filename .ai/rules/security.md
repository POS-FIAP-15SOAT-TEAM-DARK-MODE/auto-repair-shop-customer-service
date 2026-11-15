# Security

## JWT
- All protected endpoints require `Authorization: Bearer <JWT>` — return 401 if missing/invalid, 403 if the roles do not match.
- Public endpoints (no JWT): `POST /v1/auth/login`, `GET /ping`, `GET /metrics`, the Swagger routes.
- `/internal/*` is not JWT-protected: it requires `X-Internal-Token` (constant-time compare; an empty `INTERNAL_TOKEN` rejects everything; `internal.auth_failed` is logged without the token).
- JWT claims (HS256, `internal/pkg/auth/jwt.go`): `user_id` and `roles` (array). The same claims and `JWT_SECRET` are used by the login lambda.
- The auth middleware makes no DB calls — it relies entirely on the JWT payload.
- Expiry configurable via `JWT_EXPIRES_IN` env var (default `24h`).

## RBAC Enforcement
- Map roles from the JWT payload; reject out-of-scope access with 403.
- `ADMIN` is set alone: a role change replaces all current roles with one (`ChangeRole`).
- Role constants and groups live in `internal/auth/domain/auth.go`.
- Groups: `AttendantRoles`, `MechanicRoles`, `CustomerRoles`, `AttendantAndMechanicRoles` (each includes ADMIN).

## Passwords
- bcrypt. Cost comes from `BCRYPT_COST`; when unset the code uses `bcrypt.DefaultCost` (10). Manifests set `BCRYPT_COST=12` (base, stg, prd); the seed defaults to 12. The always-applied rule "min cost 12" is not enforced in code yet (open point, see `README.md`).
- Default customer password = their CPF or CNPJ.
- Never log or return password hashes.

## PII — Never Expose or Log
CPF, CNPJ, passwords, and JWT tokens must never appear in:
- HTTP response bodies or headers.
- Log entries (centralized redaction handled in `internal/pkg/logger/logger.go`).
- Error messages returned to clients.

Avoid `zap.Any` for domain/request objects that may contain PII — use explicit safe fields instead.

## Input Sanitization
- Validate all user input at handler layer before passing to service.
- Use domain validation methods (`customer.Validate()`, etc.) — never trust raw strings.
- Never expose stack traces or internal error details in production responses.

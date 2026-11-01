# Observability

## Logger
Structured JSON via Zap. Initialized in `internal/pkg/logger/logger.go` with centralized sensitive-field redaction (Issue #223).

## Log Levels
| Level | Use |
|---|---|
| `INFO` | Business events (customer created, role changed) |
| `WARN` | Business rule violations (invalid credentials, invalid document) |
| `ERROR` | Unexpected failures (DB errors, unhandled panics) |

## Canonical Event Names (Issue #236)
Reuse these — do not invent new names for the same operations:
- `customer.created`
- `customer.deleted`
- `user.role_changed`
- `auth.login_failed` — never logs the submitted credentials, CPF/CNPJ or tokens
- `internal.auth_failed` — missing or wrong `X-Internal-Token`

## Sensitive Field Redaction
The logger automatically redacts keys matching: CPF, CNPJ, password, token, dsn, database_url.
- Prefer explicit fields (`zap.String("customer_id", id)`) over `zap.Any("customer", obj)`.
- Never add logging that passes raw domain objects or request bodies with PII.

## Read Operations
- Keep success logs minimal (debug or omit).
- Always log errors from DB calls at `ERROR` level with operation context.

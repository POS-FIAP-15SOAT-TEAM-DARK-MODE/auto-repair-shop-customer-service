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
- `internal.auth_failed` — missing or wrong `X-Internal-Token` (logs the route, never the token)
- `vehicle.created` / `vehicle.deleted`
- `request.started` / `request.completed` — emitted by the Logger middleware
- `server.started` / `server.stopped`, `app.shutdown_signal`, `app.shutting_down`
- `<entity>.<action>_failed` for handler/repository failures (e.g. `customer.create_failed`, `vehicle.delete_failed`), `<entity>.<action>_payload_invalid` for bind errors
- `database.error`, `database.conflict`, `database.unhandled_error`

## Common Fields
Every line carries `service` (`SERVICE_NAME`, default `customer-service`) and `request_id` (from `X-Request-Id`, or generated). `saga_id`, `service_order_id` and `message_id` are added only by services that run the saga or consume messages; customer-service does neither.

## Sensitive Field Redaction
The logger automatically redacts keys matching: CPF, CNPJ, password, token, jwt, access_token, x-internal-token, authorization, dsn, database_url.
Query strings are logged through `logger.RedactQuery` (e.g. `GET /v1/customers?document=` never logs the document).
- Prefer explicit fields (`zap.String("customer_id", id)`) over `zap.Any("customer", obj)`.
- Never add logging that passes raw domain objects or request bodies with PII.

## Read Operations
- Keep success logs minimal (debug or omit).
- Always log errors from DB calls at `ERROR` level with operation context.

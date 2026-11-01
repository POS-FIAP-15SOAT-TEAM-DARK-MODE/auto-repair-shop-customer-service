# API

Base path `/v1`. Port `8081`. Errors: `{"code": <http>, "errors": [...]}` or `{"code": <http>, "error": "..."}`.

| Method | Path | Roles |
|---|---|---|
| POST | /auth/register | ADMIN |
| POST | /auth/login | public |
| PATCH | /users/:id/role | ADMIN |
| POST | /customers | ATTENDANT |
| GET | /customers?document= | ATTENDANT |
| GET | /customers/:id | ATTENDANT |
| PUT | /customers/:id | ATTENDANT |
| DELETE | /customers/:id | ATTENDANT |
| GET | /vehicles | ATTENDANT, MECHANIC |
| POST | /vehicles | ATTENDANT |
| PUT | /vehicles/:id | ATTENDANT |
| DELETE | /vehicles/:id | ATTENDANT |

Internal (service-to-service, header `X-Internal-Token`, never through the gateway):

| Method | Path |
|---|---|
| GET | /internal/customers/:id |
| GET | /internal/vehicles/:id |

Conventions:
- IDs: UUID strings.
- Public JSON fields are camelCase (`licensePlate`, `customerId`); internal ones are snake_case per `docs/fase4/contracts/internal-rest.md`.
- Request validation at the handler layer; services never trust raw strings.

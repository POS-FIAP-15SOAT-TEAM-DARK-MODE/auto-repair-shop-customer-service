# auto-repair-shop-customer-service

Customer & Identity service of the Auto Repair Shop system: users and authentication (JWT), roles, customers and vehicles. Port **8081**, PostgreSQL database **`customer`**.

Decision record: [`docs/adr/0002-customer-and-identity-in-one-service.md`](docs/adr/0002-customer-and-identity-in-one-service.md).

## Endpoints

Public API (through the gateway), base `/v1`:

| Method | Path | Roles |
|---|---|---|
| POST | `/auth/register` | ADMIN |
| POST | `/auth/login` | public |
| PATCH | `/users/:id/role` | ADMIN |
| GET / POST | `/customers` (`?document=` to search) | ATTENDANT |
| GET / PUT / DELETE | `/customers/:id` | ATTENDANT |
| GET | `/vehicles` (`?customerId=` or `?plate=`) | ATTENDANT, MECHANIC |
| POST | `/vehicles` | ATTENDANT |
| PUT / DELETE | `/vehicles/:id` | ATTENDANT |

Internal (service-to-service, header `X-Internal-Token: <INTERNAL_TOKEN>`, never routed by the gateway):

| Method | Path | Response |
|---|---|---|
| GET | `/internal/customers/:id` | `200 {id, user_id, type, name, email, document, company_name, phone}`; `404` |
| GET | `/internal/vehicles/:id` | `200 {id, customer_id, license_plate, brand, model, year}`; `404` |

Missing or wrong token: `401`.

Example `200` (individual customer; for a `COMPANY` customer `company_name` is the company name and `document` the CNPJ):

```json
{"id":"44444444-4444-4444-8444-000000000002","user_id":"44444444-4444-4444-8444-000000000001","type":"INDIVIDUAL","name":"João Silva","email":"joao.silva@email.com","document":"52998224725","company_name":null,"phone":"11999990001"}
```

- `name` is always the user's name, for both types.
- `company_name` is a string for `COMPANY` and `null` for `INDIVIDUAL`.
- `document` is always unformatted (CPF 11 digits, CNPJ 14 characters).
- Contract: `docs/fase4/contracts/internal-rest.md`.

Other: `GET /ping`, `GET /metrics`, `GET /swagger/index.html` (spec in `docs/swagger.yaml`).

## Running locally

```bash
cp .env.example .env          # set INTERNAL_TOKEN and, for the seed, SEED_ENABLED=true
docker compose up -d --build  # db, migrate (applies migrations/), app
curl -fsS http://localhost:8081/ping
docker compose down -v
```

`make up` / `make down` are the same commands. Without Docker: `make run` with a reachable Postgres.

The seed (`SEED_ENABLED=true`, only when the `customer` table is empty) creates the staff users `attendant@autorepairshop.com` / `mechanic@autorepairshop.com`, customers A (João Silva, CPF `52998224725`) and B (Transportes Silva Ltda, CNPJ `11222333000181`) and their vehicles, with the fixed IDs of `docs/fase4/contracts/resources.md`. Their initial passwords are the digits-only document. The admin user `admin@email.com` comes from migration `000002`.

## Database

- `migrations/000001_init_schema` — `user`, `role`, `user_role`, `customer`, `vehicle`, indexes and the four default roles.
- `migrations/000002_create_user_admin` — the initial admin.

### Login lambda (`lambda_auth_ro`)

The login lambda connects as `lambda_auth_ro` (`SELECT` only) to database `customer`.

- Local: `local/postgres-init/01-lambda-ro.sql` creates the role on the first start of the `db` container and sets default privileges for tables of the owner.
- After migrations, the grant on existing tables is still needed (default privileges only apply to objects created afterwards by the owner):
  ```bash
  docker compose exec db psql -U postgres -d customer -c "GRANT SELECT ON ALL TABLES IN SCHEMA public TO lambda_auth_ro;"
  ```
  In the Lab this is done by the bootstrap Job of the infrastructure repository.
- The integration test `lambda_compat_test.go` creates the role and grants itself, so it also runs in CI.

#### Documentos

CPF and CNPJ are **always stored without formatting**: CPF as 11 digits (`52998224725`), CNPJ as 14 characters (`11222333000181`, alphanumeric CNPJs included). The API sanitizes input before saving and searching (`GET /v1/customers?document=…` accepts formatted or plain values), and the seed writes the plain form too. The login lambda therefore receives and compares the plain form, as in `{"cpf":"52998224725"}`.

Databases seeded before this change contain formatted documents. Recreate them with `docker compose down -v` (or drop the `customer` table data) before using the seed again.

## Configuration

| Variable | Default | Notes |
|---|---|---|
| `PORT` | `8081` | |
| `POSTGRES_HOST`, `POSTGRES_PORT`, `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB` | `localhost`, `5432`, `postgres`, `postgres`, `customer` | |
| `POSTGRES_SSLMODE` | `disable` | `require` on RDS |
| `JWT_SECRET`, `JWT_EXPIRES_IN` | `your_jwt_secret`, `24h` | The same `JWT_SECRET` in every service |
| `BCRYPT_COST` | `10` in code when unset (`bcrypt.DefaultCost`); `12` in the k8s manifests and in the seed | Use `4` only in tests. Min. 12 in production is the team rule, not enforced by code yet |
| `INTERNAL_TOKEN` | — | Empty value rejects every `/internal/*` call |
| `SERVICE_ORDER_URL` | `http://service-order-service:8086` | Checked before deleting a customer or vehicle |
| `SERVICE_NAME` | `customer-service` | Value of the `service` field in every log line |
| `SEED_ENABLED` | `false` | Never `true` in prd |
| `LOG_LEVEL` | `INFO` | |

## Tests and coverage

```bash
make test                 # go test ./... --race (LOG_LEVEL=PANIC)
make coverage             # coverage.out + HTML
make test-integration     # needs Postgres with the migrations applied (CI does it with a service container)
```

Unit tests run without a database. Integration tests (`//go:build integration`) cover auth, customers, vehicles, the internal routes and the lambda compatibility.

Coverage gate: the CI `coverage` job requires at least **80%** total coverage **excluding generated mocks** (`/mocks/` paths, mockery output), the same exclusion as `sonar.exclusions=**/mocks/**` in `sonar-project.properties`.

Measured with `LOG_LEVEL=PANIC go test ./... --race -coverprofile=coverage.out` on the `chore/fa-cleanup` branch:

| Measure | Result |
|---|---|
| Total, without `/mocks/` (gate) | **90.9%** |
| Total, with mocks | 57.4% |

Per package (without mocks): `internal/auth/repository` 70.2% (below the 90% target for authentication), `internal/pkg/db` 84.2%, `internal/vehicle/adapters` 83.3%, `internal/app` 76.2%; the others are 90% or more.

## Limitações conhecidas

- **Exclusão de cliente e veículo depende do Service Order.** Antes de excluir, `DELETE /v1/customers/:id` e `DELETE /v1/vehicles/:id` consultam `GET /internal/service-orders/exists` (`SERVICE_ORDER_URL`, timeout 3 s, sem retry). Com ordens, respondem `409`; se o Service Order estiver indisponível, respondem `502` e não excluem. Ver `contracts/internal-rest.md`, seção Service Order.
- **Login do admin não verificado.** A senha original do `admin@email.com` não está documentada no repositório. O hash bcrypt foi copiado literalmente da migration `000002` do monólito, sem alteração.
- **Lint local.** `golangci-lint` v2.15 aponta 12 itens (errcheck 1, govet 3, staticcheck 8) em `internal/pkg/db/*`, `internal/pkg/logger/logger.go`, `internal/customer/*` e `internal/clients/serviceorder/client.go`. Não foram alterados na limpeza A7 para não mudar comportamento. O pre-commit do CI usa a versão fixada no `.pre-commit-config.yaml`.


## Kubernetes

`k8s/manifests` — base and overlays `lab`, `stg`, `prd`. Objects are prefixed with `customer-service` (ConfigMap `customer-service-config`, Secret `customer-service-secrets`, Job `customer-service-db-migrate`). Build with `kubectl kustomize k8s/manifests/overlays/lab`.

## Layout

```
cmd/service/            entrypoint
internal/app/           bootstrap, routing, container, middlewares, DB connection, seed
internal/auth/          users, roles, login
internal/customer/      customers (+ internal route)
internal/vehicle/       vehicles (+ internal route)
internal/clients/       HTTP client of the Service Order service
internal/pkg/           shared infrastructure (jwt, uow, db, logger, metrics, web)
internal/integration/   integration tests (build tag integration)
migrations/             golang-migrate SQL
k8s/manifests/          Kustomize base and overlays
```

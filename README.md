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
| GET | `/internal/customers/:id` | `200 {id, user_id, type, name, document}`; `404` |
| GET | `/internal/vehicles/:id` | `200 {id, customer_id, license_plate, brand, model, year}`; `404` |

Missing or wrong token: `401`. `name` is the company name for `COMPANY` customers; `document` is returned without formatting.

Other: `GET /ping`, `GET /metrics`, `GET /swagger/index.html` (spec in `docs/swagger.yaml`).

## Running locally

```bash
cp .env.example .env          # set INTERNAL_TOKEN and, for the seed, SEED_ENABLED=true
docker compose up -d --build  # db, migrate (applies migrations/), app
curl -fsS http://localhost:8081/ping
docker compose down -v
```

`make up` / `make down` are the same commands. Without Docker: `make run` with a reachable Postgres.

The seed (`SEED_ENABLED=true`, only when the `customer` table is empty) creates the staff users `attendant@autorepairshop.com` / `mechanic@autorepairshop.com`, customers A (João Silva, CPF `529.982.247-25`) and B (Transportes Silva Ltda, CNPJ `11.222.333/0001-81`) and their vehicles, with the fixed IDs of `docs/fase4/contracts/resources.md`. The admin user `admin@email.com` comes from migration `000002`.

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

#### Formato do CPF

Verified in the code and the seed:

- The API stores `customer.cpf` **sanitized (digits only)**: `ParseDocument`/`NewCustomer` strip formatting before saving.
- The **seed** inserts the CPF **formatted** (`529.982.247-25`) because it writes SQL directly.

Therefore the lambda must receive the CPF **exactly as stored**. For customers created through the API that is the 11 digits; for the seeded customer A it is `529.982.247-25`. This mismatch is an open point (see the PR): the recommended fix is to make the seed store digits only.

## Configuration

| Variable | Default | Notes |
|---|---|---|
| `PORT` | `8081` | |
| `POSTGRES_HOST`, `POSTGRES_PORT`, `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB` | `localhost`, `5432`, `postgres`, `postgres`, `customer` | |
| `POSTGRES_SSLMODE` | `disable` | `require` on RDS |
| `JWT_SECRET`, `JWT_EXPIRES_IN` | `your_jwt_secret`, `24h` | The same `JWT_SECRET` in every service |
| `BCRYPT_COST` | `12` (min. 12 in production) | Use `4` only in tests |
| `INTERNAL_TOKEN` | — | Empty value rejects every `/internal/*` call |
| `SEED_ENABLED` | `false` | Never `true` in prd |
| `LOG_LEVEL` | `INFO` | |

## Tests and coverage

```bash
make test                 # go test ./... --race (LOG_LEVEL=PANIC)
make coverage             # coverage.out + HTML
make test-integration     # needs Postgres with the migrations applied (CI does it with a service container)
```

Unit tests run without a database. Integration tests (`//go:build integration`) cover auth, customers, vehicles, the internal routes and the lambda compatibility. CI enforces at least 80% total coverage (`coverage` job).

## Kubernetes

`k8s/manifests` — base and overlays `lab`, `stg`, `prd`. Objects are prefixed with `customer-service` (ConfigMap `customer-service-config`, Secret `customer-service-secrets`, Job `customer-service-db-migrate`). Build with `kubectl kustomize k8s/manifests/overlays/lab`.

## Layout

```
cmd/service/            entrypoint
internal/app/           bootstrap, routing, container, middlewares, DB connection, seed
internal/auth/          users, roles, login
internal/customer/      customers (+ internal route)
internal/vehicle/       vehicles (+ internal route)
internal/pkg/           shared infrastructure (jwt, uow, db, logger, metrics, web)
internal/integration/   integration tests (build tag integration)
migrations/             golang-migrate SQL
k8s/manifests/          Kustomize base and overlays
```

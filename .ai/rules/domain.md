# Domain

Scope of this service: **Customer & Identity** (users, roles, authentication, customers, vehicles).
Work, supply, stock and service orders belong to other services of the system.

## Ubiquitous Language

| Code name | Business concept |
|---|---|
| `Customer` | Identified by CPF (individual) or CNPJ (company); linked 1:1 to a User |
| `User` | System account: name, email, bcrypt-hashed password, roles |
| `Vehicle` | License plate + brand/model/year, always linked to a Customer |

## RBAC

Roles: `ADMIN`, `ATTENDANT`, `MECHANIC`, `CUSTOMER`

| Role | Scope in this service |
|---|---|
| `ADMIN` | User registration and role changes (exclusive role) |
| `ATTENDANT` | Customers and vehicles (write); customers and vehicles (read) |
| `MECHANIC` | Read vehicles |
| `CUSTOMER` | Not used by any route here (customers are created by ATTENDANT) |

Role groups in `internal/auth/domain`: `AttendantRoles`, `MechanicRoles`, `CustomerRoles`, `AttendantAndMechanicRoles`.

## User & Customer Split
- Creating a Customer automatically creates a linked User with role `CUSTOMER` (same transaction, `uow`).
- Default password = CPF (individual) or CNPJ (company), bcrypt-hashed (min cost 12).
- Deleting a Customer also deletes the associated User (same transaction).
- Customer deletion does not check service orders: they live in another service and reference the customer by id only.

## Internal routes (`/internal/*`)
- Service-to-service only; never routed by the gateway. Guarded by `X-Internal-Token` (`INTERNAL_TOKEN`, constant-time compare, fails closed when empty).
- `GET /internal/customers/{id}` returns `{id, user_id, type, name, document}`; `document` is unformatted.
- `GET /internal/vehicles/{id}` returns `{id, customer_id, license_plate, brand, model, year}`.

## Input Validation
- CPF: format `###.###.###-##` + check digit algorithm.
- CNPJ: format `##.###.###/####-##` + check digit algorithm (alphanumeric CNPJ supported).
- License plate: old `ABC-1234` and Mercosul `ABC1D23`.

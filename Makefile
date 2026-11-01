-include .env
export

# Detect if running on Windows
ifeq ($(OS),Windows_NT)
    USE_WINDOWS := 1
else
    USE_WINDOWS :=
endif

run:
	go run cmd/service/main.go

test:
ifeq ($(USE_WINDOWS),1)
	set LOG_LEVEL=PANIC && go test ./... -v
else
	LOG_LEVEL=PANIC go test ./... --race -v
endif

test-integration:
ifeq ($(USE_WINDOWS),1)
	set BCRYPT_COST=4 && go test -tags=integration ./internal/integration/... -v
else
	BCRYPT_COST=4 go test -tags=integration ./internal/integration/... -v
endif

coverage:
ifeq ($(USE_WINDOWS),1)
	set LOG_LEVEL=PANIC && go test ./... --coverprofile=coverage.out
	go tool cover -html=coverage.out
else
	LOG_LEVEL=PANIC go test ./... --coverprofile=coverage.out
	go tool cover -html=coverage.out
endif

up:
	docker compose up --build

down:
	docker compose down -v

mockgen:
	go generate ./...

migrate-install:
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

migrate-up:
	@type migrate >/dev/null 2>&1 || { $(MAKE) migrate-install; }
	migrate -path migrations -database "postgresql://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:5432/customer?sslmode=disable" up

migrate-down:
	@type migrate >/dev/null 2>&1 || { $(MAKE) migrate-install; }
	migrate -path migrations -database "postgresql://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:5432/customer?sslmode=disable" down

migrate-status:
	@type migrate >/dev/null 2>&1 || { $(MAKE) migrate-install; }
	migrate -path migrations -database "postgresql://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:5432/customer?sslmode=disable" version

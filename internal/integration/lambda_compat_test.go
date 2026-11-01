//go:build integration

package integration_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/app"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

// seededCustomerACPF is the value the seed writes into customer.cpf: digits only, as the API
// stores it and as the login lambda receives it (see README, "Documentos").
const seededCustomerACPF = "52998224725"

// lambdaAuthRole is the read-only role used by the login lambda.
const lambdaAuthRole = "lambda_auth_ro"

func pgDSN(user, password string) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		user, password,
		envOr("POSTGRES_HOST", "localhost"), envOr("POSTGRES_PORT", "5432"),
		envOr("POSTGRES_DB", "customer"), envOr("POSTGRES_SSLMODE", "disable"),
	)
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// TestLambdaAuthCompat runs the two literal queries of the login lambda as lambda_auth_ro and
// checks that the role cannot write. The role and its grants are created here so the test also
// works in CI, where the Postgres service does not run the local init script.
func TestLambdaAuthCompat(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	owner, err := sql.Open("postgres", pgDSN(envOr("POSTGRES_USER", "postgres"), envOr("POSTGRES_PASSWORD", "postgres")))
	require.NoError(t, err)
	defer func() { _ = owner.Close() }()
	require.NoError(t, owner.PingContext(ctx))

	// Migrations must already be applied (CI and `make up` run them before this test).
	var tables int
	require.NoError(t, owner.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name IN ('customer','user','user_role','role')`,
	).Scan(&tables))
	require.Equal(t, 4, tables, "customer schema missing: run the migrations first")

	// Seed customer A (SEED_ENABLED=true only inserts when the customer table is empty).
	t.Setenv("SEED_ENABLED", "true")
	app.RunSeed(ctx, owner)

	lambdaPassword := envOr("LAMBDA_AUTH_PASSWORD", "lambda")
	_, err = owner.ExecContext(ctx, fmt.Sprintf(`DO $$ BEGIN
		IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = '%[1]s') THEN
			CREATE ROLE %[1]s LOGIN PASSWORD '%[2]s';
		END IF;
	END $$`, lambdaAuthRole, lambdaPassword))
	require.NoError(t, err)
	_, err = owner.ExecContext(ctx, fmt.Sprintf(`GRANT CONNECT ON DATABASE %s TO %s`, envOr("POSTGRES_DB", "customer"), lambdaAuthRole))
	require.NoError(t, err)
	_, err = owner.ExecContext(ctx, fmt.Sprintf(`GRANT USAGE ON SCHEMA public TO %s`, lambdaAuthRole))
	require.NoError(t, err)
	_, err = owner.ExecContext(ctx, fmt.Sprintf(`GRANT SELECT ON ALL TABLES IN SCHEMA public TO %s`, lambdaAuthRole))
	require.NoError(t, err)

	lambda, err := sql.Open("postgres", pgDSN(lambdaAuthRole, lambdaPassword))
	require.NoError(t, err)
	defer func() { _ = lambda.Close() }()
	require.NoError(t, lambda.PingContext(ctx))

	// Query 1 of the lambda (literal).
	var userID string
	err = lambda.QueryRowContext(ctx, `SELECT c.user_id FROM customer c WHERE c.cpf = $1`, seededCustomerACPF).Scan(&userID)
	require.NoError(t, err, "lambda query 1 must find the seeded customer A by its stored CPF")
	require.NotEmpty(t, userID)

	// Query 2 of the lambda (literal).
	rows, err := lambda.QueryContext(ctx, `SELECT r.name FROM user_role ur JOIN role r ON r.id = ur.role_id WHERE ur.user_id = $1`, userID)
	require.NoError(t, err)
	defer func() { _ = rows.Close() }()
	var roles []string
	for rows.Next() {
		var name string
		require.NoError(t, rows.Scan(&name))
		roles = append(roles, name)
	}
	require.NoError(t, rows.Err())
	require.Equal(t, []string{"CUSTOMER"}, roles)

	// The role must not be able to write.
	_, err = lambda.ExecContext(ctx, `INSERT INTO "role" (id, name) VALUES ('99', 'HACKER')`)
	require.Error(t, err, "lambda_auth_ro must not INSERT")
	_, err = lambda.ExecContext(ctx, `UPDATE customer SET phone = 'x'`)
	require.Error(t, err, "lambda_auth_ro must not UPDATE")
}

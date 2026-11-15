-- Local-only: read-only role used by the login lambda (lambda_auth_ro).
-- Runs once, on the first start of the db container (docker-entrypoint-initdb.d).
-- Connected as the owner (POSTGRES_USER) of database "customer".
-- The GRANT on tables that exist is repeated after the migrations (see README).
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'lambda_auth_ro') THEN
    CREATE ROLE lambda_auth_ro LOGIN PASSWORD 'lambda';
  END IF;
END
$$;

GRANT CONNECT ON DATABASE customer TO lambda_auth_ro;
GRANT USAGE ON SCHEMA public TO lambda_auth_ro;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON TABLES TO lambda_auth_ro;

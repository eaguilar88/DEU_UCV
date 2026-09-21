-- Runs once via docker-entrypoint-initdb.d on a fresh Postgres data directory.
-- golang-migrate's schema_migrations tracking table lives inside "deu" (see Makefile's
-- x-migrations-table DSN param), so the schema must exist before the very first `migrate up`
-- creates it — migration 000001 can't do this itself since it needs somewhere to record that
-- it ran.
CREATE SCHEMA IF NOT EXISTS deu;

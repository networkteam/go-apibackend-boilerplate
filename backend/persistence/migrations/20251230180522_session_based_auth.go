package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upSessionBasedAuth, downSessionBasedAuth)
}

func upSessionBasedAuth(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `		
		CREATE TABLE auth_sessions(
			auth_session_id uuid                     NOT NULL PRIMARY KEY,
			account_id      uuid REFERENCES accounts ON DELETE CASCADE,
			device_info     jsonb DEFAULT '{}'::jsonb NOT NULL,
			created_at      timestamp with time zone DEFAULT now()       NOT NULL
		);
		CREATE INDEX auth_sessions_account_id_idx ON auth_sessions (account_id);

		CREATE TABLE auth_access_tokens(
			auth_access_token_id uuid NOT NULL PRIMARY KEY,
			auth_session_id      uuid REFERENCES auth_sessions ON DELETE CASCADE,
			expires_at           timestamp with time zone               NOT NULL,
			created_at           timestamp with time zone DEFAULT now() NOT NULL
		);
		CREATE INDEX auth_access_tokens_auth_session_id_idx ON auth_access_tokens (auth_session_id);

		ALTER TABLE accounts DROP COLUMN secret;
	`)
	return err
}

func downSessionBasedAuth(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `
		DROP TABLE auth_access_tokens;
	`)
	return err
}

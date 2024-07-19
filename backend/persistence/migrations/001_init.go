package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upInit, downInit)
}

func upInit(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `
		CREATE TABLE accounts (
			account_id      uuid  NOT NULL PRIMARY KEY,
			role_identifier text  NOT NULL,
			secret          bytea NOT NULL,
			email_address   text  NOT NULL,
			password_hash   bytea NOT NULL
		);

		CREATE UNIQUE INDEX accounts_email_address_idx ON accounts (LOWER(email_address));

		CREATE OR REPLACE FUNCTION trigger_set_timestamp()
		RETURNS TRIGGER AS $$
		BEGIN
			NEW.updated_at = NOW();
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql;

		CREATE TABLE organisations
		(
			organisation_id uuid NOT NULL PRIMARY KEY,
			name text NOT NULL DEFAULT '',
			created_at timestamptz NOT NULL DEFAULT NOW(),
			updated_at timestamptz NOT NULL DEFAULT NOW()
		);

		CREATE TRIGGER set_timestamp
			BEFORE UPDATE ON organisations
			FOR EACH ROW
			EXECUTE PROCEDURE trigger_set_timestamp();

		ALTER TABLE accounts ADD COLUMN organisation_id uuid REFERENCES organisations (organisation_id) ON DELETE CASCADE;

		ALTER TABLE accounts
			ADD COLUMN created_at timestamptz NOT NULL DEFAULT NOW(),
			ADD COLUMN updated_at timestamptz NOT NULL DEFAULT NOW(),
			ADD COLUMN last_login timestamptz;

		CREATE TRIGGER set_timestamp
			BEFORE UPDATE ON accounts
			FOR EACH ROW
			EXECUTE PROCEDURE trigger_set_timestamp();
	`)
	return err
}

func downInit(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `
		DROP TRIGGER set_timestamp ON accounts;

		ALTER TABLE accounts
			DROP COLUMN created_at,
			DROP COLUMN updated_at,
			DROP COLUMN last_login;

		ALTER TABLE accounts DROP COLUMN organisation_id;

		DROP TABLE organisations;

		DROP FUNCTION trigger_set_timestamp;

		DROP TABLE accounts;
	`)
	return err
}

package migrations

import (
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigration(upAccountTimestamps, downAccountTimestamps)
}

func upAccountTimestamps(tx *sql.Tx) error {
	_, err := tx.Exec(`
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

func downAccountTimestamps(tx *sql.Tx) error {
	_, err := tx.Exec(`
		DROP TRIGGER set_timestamp ON accounts;

		ALTER TABLE accounts
			DROP COLUMN created_at,
			DROP COLUMN updated_at,
			DROP COLUMN last_login;
	`)
	return err
}

package migrations

import (
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigration(upConfirmationTokens, downConfirmationTokens)
}

func upConfirmationTokens(tx *sql.Tx) error {
	_, err := tx.Exec(`
		CREATE TABLE confirmation_tokens
		(
			token text NOT NULL PRIMARY KEY,
			account_id uuid NOT NULL REFERENCES accounts (account_id) ON DELETE CASCADE,
			expires timestamptz NOT NULL
		);
	`)
	return err
}

func downConfirmationTokens(tx *sql.Tx) error {
	_, err := tx.Exec(`
		DROP TABLE confirmation_tokens;
	`)
	return err
}

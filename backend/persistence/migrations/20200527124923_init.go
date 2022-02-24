package migrations

import (
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigration(upInit, downInit)
}

func upInit(tx *sql.Tx) error {
	_, err := tx.Exec(`
		CREATE TABLE accounts (
			account_id      uuid  NOT NULL PRIMARY KEY,
			role_identifier text  NOT NULL,
			secret          bytea NOT NULL,
			email_address   text  NOT NULL,
			password_hash   bytea NOT NULL
		);
		
		CREATE UNIQUE INDEX accounts_email_address_idx ON accounts (LOWER(email_address));
	`)
	return err
}

func downInit(tx *sql.Tx) error {
	_, err := tx.Exec(`
		DROP TABLE accounts;
	`)
	return err
}

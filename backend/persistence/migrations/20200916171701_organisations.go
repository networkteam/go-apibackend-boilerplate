package migrations

import (
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigration(upOrganisations, downOrganisations)
}

func upOrganisations(tx *sql.Tx) error {
	_, err := tx.Exec(`
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
	`)
	return err
}

func downOrganisations(tx *sql.Tx) error {
	_, err := tx.Exec(`
		ALTER TABLE accounts DROP COLUMN organisation_id;

		DROP TABLE organisations;
	`)
	return err
}

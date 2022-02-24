package migrations

import (
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigration(upTriggerSetTimestamp, downTriggerSetTimestamp)
}

func upTriggerSetTimestamp(tx *sql.Tx) error {
	_, err := tx.Exec(`
		CREATE OR REPLACE FUNCTION trigger_set_timestamp()
		RETURNS TRIGGER AS $$
		BEGIN
			NEW.updated_at = NOW();
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql;
	`)
	return err
}

func downTriggerSetTimestamp(tx *sql.Tx) error {
	_, err := tx.Exec(`
		DROP FUNCTION trigger_set_timestamp;
	`)
	return err
}

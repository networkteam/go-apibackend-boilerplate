package db

import (
	"database/sql"
	"os"
	"path/filepath"

	"github.com/friendsofgo/errors"
)

func ExecMigrations(db *sql.DB, sqlGlob string) error {
	files, err := filepath.Glob(sqlGlob)
	if err != nil {
		return errors.Wrapf(err, "could not list migrations in %s", sqlGlob)
	}

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return errors.Wrapf(err, "could not read migration %s", file)
		}

		_, err = db.Exec(string(data))
		if err != nil {
			return errors.Wrapf(err, "could not execute migration %s", file)
		}
	}

	return nil
}

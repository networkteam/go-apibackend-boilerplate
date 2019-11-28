package db

import (
	"database/sql"
	"io/ioutil"
	"path/filepath"

	"github.com/pkg/errors"
)

func ExecMigrations(db *sql.DB, sqlGlob string) error {
	files, err := filepath.Glob(sqlGlob)
	if err != nil {
		return errors.Wrapf(err, "could not list migrations in %s", sqlGlob)
	}

	for _, file := range files {
		data, err := ioutil.ReadFile(file)
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

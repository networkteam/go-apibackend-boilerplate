package db

import (
	"database/sql"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"testing"
)

func ExecFixtures(t *testing.T, db *sql.DB, fixtureFilenames ...string) {
	t.Helper()

	for _, file := range fixtureFilenames {
		fixtureSource := FixtureSourcePath()
		data, err := ioutil.ReadFile(fixtureSource + "/" + file + ".sql")
		if err != nil {
			t.Fatalf("could not read fixture %s: %v", file, err)
		}
		_, err = db.Exec(string(data))
		if err != nil {
			t.Fatalf("could not execute migration %s: %v", file, err)
		}
	}
}

func FixtureSourcePath() string {
	_, filename, _, _ := runtime.Caller(0)
	fixtureSource := filepath.Dir(filename + "/../../fixtures/")
	return fixtureSource
}

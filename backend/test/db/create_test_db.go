package db

import (
	"database/sql"
	"fmt"
	"log/slog"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/friendsofgo/errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jackc/pgx/v5/tracelog"
	"github.com/networkteam/slogutils"
	slogutils_tracelog "github.com/networkteam/slogutils/adapter/pgx/v5/tracelog"
	"github.com/pressly/goose/v3"

	// Import migrations with side-effect
	_ "myvendor.mytld/myproject/backend/persistence/migrations"
	"myvendor.mytld/myproject/backend/security/helper"
)

const (
	dbPort = 5432
	dbName = "myproject-test"
)

// PrepareTestDatabase prepares the test database (e.g. it creates extensions)
//
// This might be needed to be done outside of migrations because of concurrency issues when running package tests
// in parallel.
func PrepareTestDatabase() error {
	postgresDSN := fmt.Sprintf("host=localhost port=%d dbname=%s sslmode=disable", dbPort, dbName)

	connConfig, err := pgx.ParseConfig(postgresDSN)
	if err != nil {
		return errors.Wrap(err, "parsing PostgreSQL connection string")
	}
	connConfig.Tracer = &tracelog.TraceLog{
		Logger: slogutils_tracelog.NewLogger(
			slog.Default().With("component", "db.driver"),
			slogutils_tracelog.WithRemapLevel(tracelog.LogLevelInfo, slogutils.LevelTrace),
		),
		// Increase to LogLevelTrace to see all queries
		LogLevel: tracelog.LogLevelDebug,
	}
	connStr := stdlib.RegisterConnConfig(connConfig)
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return errors.Wrap(err, "open database")
	}

	_, err = db.Exec("CREATE EXTENSION IF NOT EXISTS btree_gist")
	if err != nil {
		return errors.Wrap(err, "creating extensions")
	}

	return nil
}

func CreateTestDatabase(t *testing.T) *sql.DB {
	t.Helper()

	randomSuffix, err := helper.GenerateRandomString(12)
	if err != nil {
		t.Fatalf("Failed to generate random string: %v", err)
	}
	schemaName := "test-" + strings.ToLower(randomSuffix)

	postgresDSN := fmt.Sprintf("host=localhost port=%d dbname=%s sslmode=disable search_path=%s", dbPort, dbName, schemaName)

	connConfig, err := pgx.ParseConfig(postgresDSN)
	if err != nil {
		t.Fatalf("Failed to parse PostgreSQL connection string: %v", err)
	}
	connConfig.Tracer = &tracelog.TraceLog{
		Logger: slogutils_tracelog.NewLogger(
			slog.Default().With("component", "db.driver"),
			slogutils_tracelog.WithRemapLevel(tracelog.LogLevelInfo, slogutils.LevelTrace),
			slogutils_tracelog.WithRemapErrorLevel(func(_ tracelog.LogLevel, err error) slog.Level {
				// Ignore the goose_db_version table not found error
				if err.Error() == "ERROR: relation \"goose_db_version\" does not exist (SQLSTATE 42P01)" {
					return slogutils.LevelTrace
				}
				return slog.LevelError
			}),
		),
		// Increase to LogLevelTrace to see all queries
		LogLevel: tracelog.LogLevelDebug,
	}
	connStr := stdlib.RegisterConnConfig(connConfig)
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		t.Fatalf("Failed to open database connection: %v", err)
	}

	_, err = db.Exec("CREATE SCHEMA \"" + schemaName + "\"")
	if err != nil {
		t.Fatalf("Failed to create schema: %v", err)
	}

	t.Cleanup(func() {
		_, err := db.Exec("DROP SCHEMA \"" + schemaName + "\" CASCADE")
		if err != nil {
			t.Fatalf("Failed to drop schema: %v", err)
		}
		err = db.Close()
		if err != nil {
			t.Logf("Error closing test DB: %v", err)
		}
		stdlib.UnregisterConnConfig(connStr)
	})

	_, filename, _, _ := runtime.Caller(0)
	migrationSource := filepath.Dir(filename + "/../../../persistence/migrations/")

	goose.SetLogger(testGooseLogger{t})
	err = goose.Up(db, migrationSource)
	if err != nil {
		t.Fatalf("Failed to execute migrations: %v", err)
	}

	return db
}

type testGooseLogger struct {
	t *testing.T
}

func (t testGooseLogger) Fatal(v ...interface{}) {
	t.t.Fatal(v...)
}

func (t testGooseLogger) Fatalf(format string, v ...interface{}) {
	t.t.Fatalf(format, v...)
}

func (t testGooseLogger) Print(_ ...interface{}) {
}

func (t testGooseLogger) Println(_ ...interface{}) {
}

func (t testGooseLogger) Printf(_ string, _ ...interface{}) {
}

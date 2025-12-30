package db

import (
	"context"
	"database/sql"
	std_errors "errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/friendsofgo/errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jackc/pgx/v5/tracelog"
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
func PrepareTestDatabase(ctx context.Context) error {
	postgresDSN := fmt.Sprintf("host=localhost port=%d dbname=%s sslmode=disable", dbPort, dbName)
	postgresUser := os.Getenv("POSTGRES_USER")
	if postgresUser != "" {
		postgresDSN += fmt.Sprintf(" user=%s", postgresUser)
	}

	connConfig, err := pgx.ParseConfig(postgresDSN)
	if err != nil {
		return errors.Wrap(err, "parsing PostgreSQL connection string")
	}
	connConfig.Tracer = &tracelog.TraceLog{
		Logger:   slogutils_tracelog.NewLogger(slog.Default().With("component", "db.driver")),
		LogLevel: tracelog.LogLevelWarn,
	}
	connStr := stdlib.RegisterConnConfig(connConfig)
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return errors.Wrap(err, "open database")
	}
	defer func() {
		closeErr := db.Close()
		err = std_errors.Join(err, closeErr)
	}()

	_, err = db.ExecContext(ctx, "CREATE EXTENSION IF NOT EXISTS btree_gist")
	if err != nil {
		return errors.Wrap(err, "creating extensions")
	}

	return nil
}

func CreateTestDatabase(t *testing.T) *sql.DB {
	t.Helper()

	schemaName := generateTestDatabaseSchemaName(t)

	db := openTestDatabase(t, schemaName, nil)
	createTestDatabaseSchema(t, db, schemaName)
	executeTestDatabaseMigrations(t, db)

	return db
}

func generateTestDatabaseSchemaName(t *testing.T) string {
	t.Helper()

	randomSuffix, err := helper.GenerateRandomString(12)
	if err != nil {
		t.Fatalf("Failed to generate random string: %v", err)
	}
	schemaName := "test-" + strings.ToLower(randomSuffix)
	return schemaName
}

func openTestDatabase(t *testing.T, schemaName string, overrideRuntimeParams map[string]string) *sql.DB {
	t.Helper()

	postgresDSN := fmt.Sprintf("host=localhost port=%d dbname=%s sslmode=disable search_path=%s,public", dbPort, dbName, schemaName)
	postgresUser := os.Getenv("POSTGRES_USER")
	if postgresUser != "" {
		postgresDSN += fmt.Sprintf(" user=%s", postgresUser)
	}

	// Increase to LogLevelInfo to log all queries
	logVerbosity := tracelog.LogLevelWarn
	if verbosityStr := os.Getenv("BACKEND_VERBOSITY"); verbosityStr != "" {
		verbosityInt, err := strconv.Atoi(verbosityStr)
		if err == nil {
			logVerbosity = tracelog.LogLevel(verbosityInt)
		}
	}

	connConfig, err := pgx.ParseConfig(postgresDSN)
	if err != nil {
		t.Fatalf("Failed to parse PostgreSQL connection string: %v", err)
	}
	connConfig.Tracer = &tracelog.TraceLog{
		Logger: slogutils_tracelog.NewLogger(
			slog.Default(),
			slogutils_tracelog.WithIgnoreErrors(func(err error) bool {
				return err.Error() == "ERROR: relation \"goose_db_version\" does not exist (SQLSTATE 42P01)"
			}),
		),
		LogLevel: logVerbosity,
	}
	for k, v := range overrideRuntimeParams {
		connConfig.RuntimeParams[k] = v
	}
	connStr := stdlib.RegisterConnConfig(connConfig)
	t.Cleanup(func() {
		stdlib.UnregisterConnConfig(connStr)
	})
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		t.Fatalf("Failed to open database connection: %v", err)
	}
	t.Cleanup(func() {
		err = db.Close()
		if err != nil {
			t.Logf("Error closing test DB: %v", err)
		}
	})
	return db
}

func createTestDatabaseSchema(t *testing.T, db *sql.DB, schemaName string) {
	t.Helper()

	_, err := db.ExecContext(t.Context(), "CREATE SCHEMA \""+schemaName+"\"")
	if err != nil {
		t.Fatalf("Failed to create schema: %v", err)
	}

	t.Cleanup(func() {
		ctx := context.Background() //
		_, err := db.ExecContext(ctx, "DROP SCHEMA \""+schemaName+"\" CASCADE")
		if err != nil {
			t.Fatalf("Failed to drop schema: %v", err)
		}
	})
}

func executeTestDatabaseMigrations(t *testing.T, db *sql.DB) {
	t.Helper()

	_, filename, _, _ := runtime.Caller(0)
	migrationSource := filepath.Dir(filename + "/../../../persistence/migrations/")

	goose.SetLogger(testGooseLogger{t})
	err := goose.Up(db, migrationSource)
	if err != nil {
		t.Fatalf("Failed to execute migrations: %v", err)
	}
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

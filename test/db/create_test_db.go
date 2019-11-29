package db

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
	"time"

	"github.com/lib/pq"
	"github.com/networkteam/go-sqllogger"
)

const (
	dbPort = 5432
	dbName = "myproject-test"
)

func CreateTestDatabase(t *testing.T) (*sql.DB, func()) {
	t.Helper()

	// TODO Better randomized name
	rand.Seed(time.Now().UnixNano())
	schemaName := "test" + strconv.FormatInt(rand.Int63(), 10)

	connectionString := fmt.Sprintf("port=%d dbname=%s sslmode=disable search_path=%s", dbPort, dbName, schemaName)

	pqConnector, err := pq.NewConnector(connectionString)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	logger := sqllogger.NewDefaultSQLLogger(log.New(os.Stdout, "SQL: ", 0))
	connector := sqllogger.LoggingConnector(logger, pqConnector)

	db := sql.OpenDB(connector)

	// Do not log schema creation and migrations
	logger.Enabled = false

	_, err = db.Exec("CREATE SCHEMA " + schemaName)
	if err != nil {
		t.Fatalf("Failed to create schema: %v", err)
	}

	_, filename, _, _ := runtime.Caller(0)
	migrationSource := filepath.Dir(filename + "/../../../persistence/migrations/")
	err = ExecMigrations(db, migrationSource+"/*.up.sql")
	if err != nil {
		t.Fatalf("Failed to execute migrations: %v", err)
	}

	// TODO Enable if needed
	// logger.Enabled = true

	return db, func() {
		_, err := db.Exec("DROP SCHEMA " + schemaName + " CASCADE")
		if err != nil {
			t.Fatalf("Failed to drop database: %v", err)
		}
		db.Close()
	}
}

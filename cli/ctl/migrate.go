package main

import (
	"fmt"
	"strings"

	"github.com/apex/log"
	_ "github.com/golang-migrate/migrate/database/postgres"
	_ "github.com/golang-migrate/migrate/source/file"
	"github.com/spf13/cobra"
)

var migrateFlags struct {
	directory string
}

type migrationLogger struct{}

func (l *migrationLogger) Printf(format string, v ...interface{}) {
	log.Infof(strings.TrimRight(format, "\n"), v...)
}

func (l *migrationLogger) Verbose() bool {
	return rootFlags.verbosity > 2
}

func init() {
	rootCmd.AddCommand(migrateCmd)

	migrateCmd.PersistentFlags().StringVar(&migrateFlags.directory, "dir", "./migrations", "Directory where migrations are stored")
}

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate the database",
}

func parseDsn(postgresDSN string) (string, error) {
	var migrationDsn string
	dsnParts := strings.Split(postgresDSN, " ")
	if len(dsnParts) < 2 {
		return "", fmt.Errorf("cannot split postgres dsn: %s", postgresDSN)
	}

	postgresDnsParts := map[string]string{}

	for _, part := range dsnParts {
		parts := strings.Split(part, "=")
		if len(parts) != 2 {
			return "", fmt.Errorf("cannot split part of postgresDSN: %v", part)
		}
		postgresDnsParts[parts[0]] = parts[1]
	}

	if v, ok := postgresDnsParts["user"]; ok {
		migrationDsn = v
		if v, ok := postgresDnsParts["password"]; ok {
			migrationDsn = migrationDsn + ":" + v
			delete(postgresDnsParts, "password")
		}
		migrationDsn = migrationDsn + "@"
		delete(postgresDnsParts, "user")
	}
	if v, ok := postgresDnsParts["host"]; ok {
		migrationDsn = migrationDsn + v
		delete(postgresDnsParts, "host")
	}
	if v, ok := postgresDnsParts["port"]; ok {
		migrationDsn = migrationDsn + ":" + v
		delete(postgresDnsParts, "port")
	}
	if v, ok := postgresDnsParts["dbname"]; ok {
		migrationDsn = migrationDsn + "/" + v
		delete(postgresDnsParts, "dbname")
	}
	var keys []string
	for k := range postgresDnsParts {
		keys = append(keys, k)
	}
	dsnParts = nil
	for _, k := range keys {
		dsnParts = append(dsnParts, k+"="+postgresDnsParts[k])
	}

	return migrationDsn + "?" + strings.Join(dsnParts, "&"), nil
}

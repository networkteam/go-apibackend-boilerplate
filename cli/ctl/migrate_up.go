package main

import (
	"fmt"
	"path/filepath"

	"github.com/golang-migrate/migrate"
	_ "github.com/golang-migrate/migrate/database/postgres"
	_ "github.com/golang-migrate/migrate/source/file"
	"github.com/spf13/cobra"
)

func init() {
	migrateCmd.AddCommand(migrateUpCmd)
}

var migrateUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Migrate the database up",
	RunE: func(cmd *cobra.Command, args []string) error {
		ok, err := isDirectory(migrateFlags.directory)
		if err != nil {
			return fmt.Errorf("cannot check if `dir` is a directory: %s", err)
		}

		if !ok {
			return fmt.Errorf("argument `dir` must be a valid directory")
		}

		migrateFlags.directory, err = filepath.Abs(migrateFlags.directory)
		if err != nil {
			return fmt.Errorf("cannot get absolute path of `dir`: %s", err)
		}

		migrationDsn, err := parseDsn(rootFlags.postgresDSN)
		if err != nil {
			return fmt.Errorf("error parsing DSN: %s", err)
		}

		// Do not display usage on errors after arguments are validated
		// See https://github.com/spf13/cobra/issues/340
		cmd.SilenceUsage = true

		m, err := migrate.New(pathToFileURL(migrateFlags.directory), fmt.Sprintf("postgres://%s", migrationDsn))
		if err != nil {
			return fmt.Errorf("cannot create migration: %s", err)
		}

		m.Log = new(migrationLogger)

		version, dirty, err := m.Version()
		if err != nil && err != migrate.ErrNilVersion {
			return fmt.Errorf("cannot get current version: %s", err)
		}
		if dirty {
			return fmt.Errorf("migration state is dirty")
		}
		fmt.Printf("Migrating from version %d\n", version)

		err = m.Up()
		if err == migrate.ErrNoChange {
			fmt.Println("No migrations needed")
		} else if err != nil {
			return fmt.Errorf("migration failed: %s", err)
		}

		return nil
	},
}

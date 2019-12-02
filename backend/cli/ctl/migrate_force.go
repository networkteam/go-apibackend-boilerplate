package main

import (
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/spf13/cobra"
)

func init() {
	migrateCmd.AddCommand(migrateForceCmd)
}

var migrateForceCmd = &cobra.Command{
	Use:   "force [version]",
	Short: "Force a migration version (and reset dirty flag)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		forceVersion, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("could not parse version: %s", err)
		}

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

		err = m.Force(forceVersion)
		if err != nil {
			return fmt.Errorf("forcing migration version failed: %s", err)
		}

		return nil
	},
}

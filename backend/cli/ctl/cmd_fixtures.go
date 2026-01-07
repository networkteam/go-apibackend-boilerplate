package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"syscall"

	"github.com/friendsofgo/errors"
	"github.com/urfave/cli/v2"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/term"

	"myvendor.mytld/myproject/backend/persistence/repository"
	security_helper "myvendor.mytld/myproject/backend/security/helper"
	"myvendor.mytld/myproject/backend/test/fixtures"
)

func newFixturesCmd() *cli.Command {
	return &cli.Command{
		Name:  "fixtures",
		Usage: "Set up fixtures",

		Subcommands: []*cli.Command{
			{
				Name:  "import",
				Usage: "Set up fixtures from static files for testing, will truncate the DB",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "force",
						Usage: "Force truncating and re-import of fixture data. Otherwise the import will be skipped if data (any account) already exists.",
					},
					&cli.BoolFlag{
						Name:  "extended",
						Usage: "Import extended fixtures for testing of different scenarios (e.g. lists with articles).",
					},
				},
				Action: fixturesImportAction,
			},
			{
				Name:  "generate-password-hash",
				Usage: "Generate a password hash for a given password",
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:  "cost",
						Usage: "Hash cost: defaults to min cost for bcrypt - only use for development and testing.",
						Value: bcrypt.MinCost,
					},
				},
				Action: func(_ *cli.Context) error {
					fmt.Print("Enter password:") //nolint:forbidigo
					line, err := term.ReadPassword(syscall.Stdin)
					if err != nil {
						return err
					}
					fmt.Println() //nolint:forbidigo
					password := strings.TrimSpace(string(line))

					hash, err := security_helper.GenerateHashFromPassword([]byte(password), bcrypt.MinCost)
					if err != nil {
						return errors.Wrap(err, "generating password hash")
					}

					fmt.Println(string(hash)) //nolint:forbidigo

					return nil
				},
			},
		},
	}
}

func fixturesImportAction(c *cli.Context) error {
	force := c.Bool("force")

	db, err := connectDatabase(c, nil)
	if err != nil {
		return err
	}

	accountCount, err := repository.CountAccounts(c.Context, db, repository.AccountsFilter{})
	if err != nil {
		return errors.Wrap(err, "counting accounts")
	}
	if accountCount > 0 && !force {
		slog.Info("Skipping fixtures import because there are already accounts in the database and --force was not set")
		return nil
	}

	if force {
		err = truncateDB(c.Context, db, nil)
		if err != nil {
			return err
		}
	}

	slog.Info("Creating fixture data")

	// Load the following SQL fixtures
	fixtureSQLFilenames := []string{
		"base",
	}

	if c.Bool("extended") {
		// TODO Add additional extended fixture files here (must be added to embed.go as well)
	}

	for _, file := range fixtureSQLFilenames {
		slog.Info("Importing SQL", "file", file)

		data, err := fixtures.FS.ReadFile(fmt.Sprintf("%s.sql", file))
		if err != nil {
			return errors.Wrapf(err, "could not read fixture %q", file)
		}

		_, err = db.ExecContext(c.Context, string(data))
		if err != nil {
			return errors.Wrapf(err, "could not execute fixture %s", file)
		}
	}

	return nil
}

func truncateDB(ctx context.Context, db *sql.DB, skipTables []string) error {
	tableNames, err := getTableNames(ctx, db, skipTables)
	if err != nil {
		return errors.Wrap(err, "getting table names")
	}

	slog.Info("Truncating tables", "tables", tableNames, "skippedTables", skipTables)

	// nosemgrep: go.lang.security.audit.database.string-formatted-query.string-formatted-query, go.lang.security.audit.sqli.gosql-sqli.gosql-sqli
	_, err = db.ExecContext(ctx, fmt.Sprintf("TRUNCATE TABLE %s CASCADE", strings.Join(tableNames, ", ")))
	if err != nil {
		return errors.Wrap(err, "truncating tables")
	}
	return nil
}

func getTableNames(ctx context.Context, db *sql.DB, skipTables []string) ([]string, error) {
	rows, err := db.QueryContext(ctx, "SELECT tablename FROM pg_catalog.pg_tables WHERE schemaname = 'public' AND tablename != 'goose_db_version'")
	if err != nil {
		return nil, errors.Wrap(err, "querying tables")
	}
	defer rows.Close()

	var tableNames []string

	var tableName string
	for rows.Next() {
		err = rows.Scan(&tableName)
		if err != nil {
			return nil, errors.Wrap(err, "scanning result")
		}

		if !slices.Contains(skipTables, tableName) {
			tableNames = append(tableNames, tableName)
		}
	}
	return tableNames, nil
}

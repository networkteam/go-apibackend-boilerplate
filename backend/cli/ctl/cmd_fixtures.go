package main

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/apex/log"
	"github.com/friendsofgo/errors"
	"github.com/urfave/cli/v2"

	"myvendor.mytld/myproject/backend/persistence/repository"
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
				},
				Action: fixturesImportAction,
			},
		},
	}
}

func fixturesImportAction(c *cli.Context) error {
	force := c.Bool("force")

	db, err := connectDatabase(c)
	if err != nil {
		return err
	}

	accountCount, err := repository.CountAccounts(c.Context, db, repository.AccountsFilter{})
	if err != nil {
		return errors.Wrap(err, "counting accounts")
	}
	if accountCount > 0 && !force {
		log.Info("Skipping fixtures import because there are already accounts in the database and --force was not set")
		return nil
	}

	if force {
		err = truncateDB(db)
		if err != nil {
			return err
		}
	}

	log.Info("Creating fixture data")

	// Load the following SQL fixtures
	fixtureSQLFilenames := []string{
		"base",
	}

	for _, file := range fixtureSQLFilenames {
		log.Infof("Importing SQL %q", file)

		data, err := fixtures.FS.ReadFile(fmt.Sprintf("%s.sql", file))
		if err != nil {
			return errors.Wrapf(err, "could not read fixture %q", file)
		}

		_, err = db.Exec(string(data))
		if err != nil {
			return errors.Wrapf(err, "could not execute fixture %s", file)
		}
	}

	return nil
}

func truncateDB(db *sql.DB) error {
	tableNames, err := getTableNames(db)
	if err != nil {
		return errors.Wrap(err, "getting table names")
	}

	log.WithField("tables", tableNames).Info("Truncating tables")

	// nosemgrep: go.lang.security.audit.database.string-formatted-query.string-formatted-query, go.lang.security.audit.sqli.gosql-sqli.gosql-sqli
	_, err = db.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", strings.Join(tableNames, ", ")))
	if err != nil {
		return errors.Wrap(err, "truncating tables")
	}
	return nil
}

func getTableNames(db *sql.DB) ([]string, error) {
	rows, err := db.Query("SELECT tablename FROM pg_catalog.pg_tables WHERE schemaname = 'public' AND tablename != 'goose_db_version'")
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
		tableNames = append(tableNames, tableName)
	}
	return tableNames, nil
}

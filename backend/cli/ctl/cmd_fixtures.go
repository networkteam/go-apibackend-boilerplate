package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/apex/log"
	"github.com/friendsofgo/errors"
	"github.com/urfave/cli/v2"
)

func newFixturesCmd() *cli.Command {
	return &cli.Command{
		Name:  "fixtures",
		Usage: "Set up fixtures",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:     "confirm",
				Required: true,
			},
		},

		Action: func(c *cli.Context) error {
			if !c.Bool("confirm") {
				return cli.Exit("must pass --confirm to truncate database and import data", 1)
			}

			db, err := connectDatabase(c)
			if err != nil {
				return err
			}

			err = truncateDB(db)
			if err != nil {
				return err
			}

			log.Info("Creating fixture data")

			// Load the following SQL fixtures
			fixtureFilenames := []string{
				"base",
			}

			for _, file := range fixtureFilenames {
				log.Infof("Importing %q", file)
				data, err := ioutil.ReadFile(fmt.Sprintf("./test/fixtures/%s.sql", file))
				if err != nil {
					return errors.Wrapf(err, "could not read fixture %q", file)
				}
				_, err = db.Exec(string(data))
				if err != nil {
					return errors.Wrapf(err, "could not execute fixture %s", file)
				}
			}

			return nil
		},
	}
}

func truncateDB(db *sql.DB) error {
	tableNames, err := getTableNames(db)
	if err != nil {
		return errors.Wrap(err, "getting table names")
	}

	log.WithField("tables", tableNames).Info("Truncating tables")

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

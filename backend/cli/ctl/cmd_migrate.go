package main

import (
	"github.com/friendsofgo/errors"
	"github.com/pressly/goose/v3"
	"github.com/urfave/cli/v2"

	_ "myvendor.mytld/myproject/backend/persistence/migrations"
)

func newMigrateCmd() *cli.Command {
	return &cli.Command{
		Name:  "migrate",
		Usage: "Manage database migrations",
		Subcommands: []*cli.Command{
			{
				Name:  "up",
				Usage: "Migrate up",
				Flags: []cli.Flag{
					&cli.Int64Flag{
						Name: "version",
					},
				},
				Action: func(c *cli.Context) error {
					db, err := connectDatabase(c)
					if err != nil {
						return err
					}

					if c.IsSet("version") {
						err = goose.UpTo(db, "./persistence/migrations", c.Int64("version"))
						if err != nil {
							return errors.Wrap(err, "applying migrations")
						}
					} else {
						err = goose.Up(db, "./persistence/migrations")
						if err != nil {
							return errors.Wrap(err, "applying migrations")
						}
					}

					return nil
				},
			},
			{
				Name:  "down",
				Usage: "Migrate down",
				Flags: []cli.Flag{
					&cli.Int64Flag{
						Name: "version",
					},
				},
				Action: func(c *cli.Context) error {
					db, err := connectDatabase(c)
					if err != nil {
						return err
					}

					if c.IsSet("version") {
						err = goose.DownTo(db, "./persistence/migrations", c.Int64("version"))
						if err != nil {
							return errors.Wrap(err, "applying migrations")
						}
					} else {
						err = goose.Down(db, "./persistence/migrations")
						if err != nil {
							return errors.Wrap(err, "applying migrations")
						}
					}

					return nil
				},
			},
		},
	}
}

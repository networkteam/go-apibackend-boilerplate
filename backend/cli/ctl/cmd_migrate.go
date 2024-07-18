package main

import (
	"github.com/friendsofgo/errors"
	"github.com/pressly/goose/v3"
	"github.com/urfave/cli/v2"

	"myvendor.mytld/myproject/backend/persistence/migrations"
)

func newMigrateCmd() *cli.Command {
	return &cli.Command{
		Name:  "migrate",
		Usage: "Manage database migrations",
		Before: func(c *cli.Context) error {
			goose.SetBaseFS(migrations.FS)
			return nil
		},
		Subcommands: []*cli.Command{
			{
				Name:  "up",
				Usage: "Migrate up",
				Flags: []cli.Flag{
					&cli.Int64Flag{
						Name: "version",
					},
					&cli.BoolFlag{
						Name:  "allow-missing",
						Usage: "Allow migration with missing previous migrations (not recommended for production)",
					},
				},
				Action: func(c *cli.Context) error {
					db, err := connectDatabase(c)
					if err != nil {
						return err
					}

					var opts []goose.OptionsFunc
					if c.Bool("allow-missing") {
						opts = append(opts, goose.WithAllowMissing())
					}

					if c.IsSet("version") {
						err = goose.UpTo(db, ".", c.Int64("version"), opts...)
						if err != nil {
							return errors.Wrap(err, "applying migrations")
						}
					} else {
						err = goose.Up(db, ".", opts...)
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
						err = goose.DownTo(db, ".", c.Int64("version"))
						if err != nil {
							return errors.Wrap(err, "applying migrations")
						}
					} else {
						err = goose.Down(db, ".")
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

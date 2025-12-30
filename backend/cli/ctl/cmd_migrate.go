package main

import (
	"text/template"

	"github.com/friendsofgo/errors"
	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/database"
	"github.com/urfave/cli/v2"

	"myvendor.mytld/myproject/backend/persistence/migrations"
)

func newMigrateCmd() *cli.Command {
	return &cli.Command{
		Name:  "migrate",
		Usage: "Manage database migrations",
		Before: func(_ *cli.Context) error {
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
					db, err := connectDatabase(c, nil)
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
					db, err := connectDatabase(c, nil)
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
			{
				Name:      "create",
				Flags:     []cli.Flag{},
				ArgsUsage: "<migration-name>",
				Before: func(c *cli.Context) error {
					if c.Args().First() == "" {
						return errors.New("missing migration name")
					}
					return nil
				},
				Action: func(c *cli.Context) error {
					db, err := connectDatabase(c, nil)
					if err != nil {
						return err
					}

					err = goose.CreateWithTemplate(db, "persistence/migrations", goSQLMigrationTemplate, c.Args().First(), "go")
					if err != nil {
						return errors.Wrap(err, "creating migration")
					}

					return nil
				},
			},
			{
				Name:        "status",
				Description: "Show detailed status of database migrations, if --ensure-applied is set, will return an error if there are pending migrations",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "ensure-applied",
						Usage: "Return an error if there are pending migrations",
					},
				},
				Action: func(c *cli.Context) error {
					db, err := connectDatabase(c, nil)
					if err != nil {
						return err
					}

					if c.Bool("ensure-applied") {
						expectedMigrations, err := goose.CollectMigrations(".", 0, goose.MaxVersion)
						if err != nil {
							return errors.Wrap(err, "collecting migrations")
						}

						store, err := database.NewStore(goose.DialectPostgres, goose.DefaultTablename)
						if err != nil {
							return errors.Wrap(err, "creating migration store")
						}

						existingMigrations, err := store.ListMigrations(c.Context, db)
						if err != nil {
							return errors.Wrap(err, "listing existing migrations")
						}

						existingMigrationsByVersion := make(map[int64]*database.ListMigrationsResult, len(existingMigrations))
						for _, m := range existingMigrations {
							existingMigrationsByVersion[m.Version] = m
						}

						for _, m := range expectedMigrations {
							if existingMigration, ok := existingMigrationsByVersion[m.Version]; !ok || !existingMigration.IsApplied {
								return errors.Errorf("pending migration found: %d", m.Version)
							}
						}

						return nil
					}

					err = goose.Status(db, ".")
					if err != nil {
						return errors.Wrap(err, "getting migration status")
					}

					return nil
				},
			},
		},
	}
}

//nolint:gochecknoglobals
var goSQLMigrationTemplate = template.Must(template.New("goose.go-migration").Parse(`package migrations

import (
	"context"
	"database/sql"
	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(up{{.CamelName}}, down{{.CamelName}})
}

func up{{.CamelName}}(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, ` + "`" + `
		-- CREATE TABLE ...
	` + "`" + `)
	return err
}

func down{{.CamelName}}(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, ` + "`" + `
		-- DROP TABLE ...
	` + "`" + `)
	return err
}
`))

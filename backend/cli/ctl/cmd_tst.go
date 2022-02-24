package main

import (
	"github.com/urfave/cli/v2"

	test_db "myvendor.mytld/myproject/backend/test/db"
)

func newTestCmd() *cli.Command {
	return &cli.Command{
		Name:  "test",
		Usage: "Test utilities",
		Subcommands: []*cli.Command{
			{
				Name:  "preparedb",
				Usage: "Prepare test database (e.g. install extensions)",
				Action: func(c *cli.Context) error {
					return test_db.PrepareTestDatabase()
				},
			},
		},
	}
}

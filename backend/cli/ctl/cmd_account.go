package main

import (
	"fmt"
	"strings"
	"syscall"

	"github.com/friendsofgo/errors"
	"github.com/gofrs/uuid"
	"github.com/urfave/cli/v2"
	"golang.org/x/term"

	"myvendor.mytld/myproject/backend/domain/command"
	"myvendor.mytld/myproject/backend/domain/types"
	"myvendor.mytld/myproject/backend/handler"
)

func newAccountCmd() *cli.Command {
	return &cli.Command{
		Name:  "account",
		Usage: "Manage accounts",
		Subcommands: []*cli.Command{
			{
				Name:  "create",
				Usage: "Create account",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "role",
						Value: string(types.RoleSystemAdministrator),
					},
					&cli.StringFlag{
						Name:     "email",
						Required: true,
					},
					&cli.StringFlag{
						Name: "organisationId",
					},
				},
				Action: func(c *cli.Context) error {
					fmt.Print("Enter password:") //nolint:forbidigo
					line, err := term.ReadPassword(syscall.Stdin)
					if err != nil {
						return err
					}
					fmt.Println() //nolint:forbidigo
					password := strings.TrimSpace(string(line))

					cmd, err := command.NewAccountCreateCmd(c.String("email"), types.Role(c.String("role")), password)
					if err != nil {
						return err
					}

					organisationIDStr := c.String("organisationId")
					if organisationIDStr != "" {
						organisationID, err := uuid.FromString(organisationIDStr)
						if err != nil {
							return errors.Wrap(err, "parsing organisation id")
						}
						cmd.OrganisationID = uuid.NullUUID{Valid: true, UUID: organisationID}
					}

					db, err := connectDatabase(c)
					if err != nil {
						return err
					}

					timeSource, err := newCurrentTimeSource(c)
					if err != nil {
						return err
					}

					config, err := getConfig(c)
					if err != nil {
						return err
					}

					h := handler.NewHandler(db, config, handler.Deps{
						TimeSource: timeSource,
					})
					err = h.AccountCreate(c.Context, cmd)
					if err != nil {
						return err
					}

					return nil
				},
			},
			newAccountListCmd(),
		},
	}
}

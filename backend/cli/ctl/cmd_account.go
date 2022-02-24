package main

import (
	"fmt"
	"strings"
	"syscall"

	"github.com/friendsofgo/errors"
	"github.com/gofrs/uuid"
	"github.com/urfave/cli/v2"
	"golang.org/x/term"

	"myvendor.mytld/myproject/backend/domain"
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
						Value: string(domain.RoleSystemAdministrator),
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
					fmt.Print("Enter password:")
					line, err := term.ReadPassword(syscall.Stdin)
					if err != nil {
						return err
					}
					fmt.Println()
					password := strings.TrimSpace(string(line))

					cmd, err := domain.NewAccountCreateCmd(c.String("email"), domain.Role(c.String("role")), password)
					if err != nil {
						return err
					}

					organisationIdStr := c.String("organisationId")
					if organisationIdStr != "" {
						organisationId, err := uuid.FromString(organisationIdStr)
						if err != nil {
							return errors.Wrap(err, "parsing organisationId")
						}
						cmd.OrganisationID = uuid.NullUUID{Valid: true, UUID: organisationId}
					}

					db, err := connectDatabase(c)
					if err != nil {
						return err
					}

					timeSource, err := newCurrentTimeSource(c)
					if err != nil {
						return err
					}

					h := handler.NewHandler(db, timeSource, nil, getConfig(c))
					err = h.AccountCreate(c.Context, cmd)
					if err != nil {
						return err
					}

					return nil
				},
			},
		},
	}
}

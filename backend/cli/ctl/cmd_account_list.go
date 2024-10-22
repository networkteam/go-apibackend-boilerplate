package main

import (
	"fmt"

	"github.com/friendsofgo/errors"
	"github.com/gofrs/uuid"
	"github.com/urfave/cli/v2"

	"myvendor.mytld/myproject/backend/domain/types"
	"myvendor.mytld/myproject/backend/persistence/repository"
)

func newAccountListCmd() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List accounts",
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:  "role",
				Usage: "Filter by role(s)",
			},
			&cli.IntFlag{
				Name:  "per-page",
				Value: 100,
			},
			&cli.IntFlag{
				Name:  "page",
				Value: 0,
			},
		},
		Action: func(c *cli.Context) error {
			db, err := connectDatabase(c)
			if err != nil {
				return err
			}

			roleIdentifiers := c.StringSlice("role")

			roles := make([]types.Role, len(roleIdentifiers))
			for i, roleIdentifier := range roleIdentifiers {
				role := types.Role(roleIdentifier)
				if !role.IsValid() {
					return errors.Errorf("invalid role %q", roleIdentifier)
				}
				roles[i] = role
			}

			page := c.Int("page")
			perPage := c.Int("per-page")
			accounts, err := repository.FindAllAccounts(c.Context, db,
				repository.AccountsFilter{
					Roles: roles,
				},
				repository.WithLimit(perPage),
				repository.WithOffset(page*perPage),
				repository.WithSort("emailAddress", repository.SortOrderAsc),
			)
			if err != nil {
				return errors.Wrap(err, "finding accounts")
			}

			for _, account := range accounts {
				var relatedID *uuid.UUID
				orgID := account.OrganisationID.UUID
				if account.OrganisationID.Valid {
					relatedID = &orgID
				}
				fmt.Printf("%s\t%s\t%s\t%s\n", account.ID, account.EmailAddress, account.Role, relatedID) //nolint:forbidigo
			}

			return nil
		},
	}
}

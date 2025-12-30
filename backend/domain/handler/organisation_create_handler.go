package handler

import (
	"context"
	"database/sql"

	"github.com/friendsofgo/errors"
	"github.com/networkteam/slogutils"

	"myvendor.mytld/myproject/backend/domain/command"
	"myvendor.mytld/myproject/backend/persistence/repository"
	"myvendor.mytld/myproject/backend/security/authentication"
	"myvendor.mytld/myproject/backend/security/authorization"
)

func (h *Handler) OrganisationCreate(ctx context.Context, cmd command.OrganisationCreateCmd) error {
	logger := slogutils.FromContext(ctx).
		With(
			"component", "handler",
			"handler", "OrganisationCreate",
		)

	logger.DebugContext(ctx, "Handling organisation create command", "cmd", cmd)

	if err := cmd.Validate(); err != nil {
		return err
	}

	authCtx := authentication.GetAuthContext(ctx)
	if err := authorization.NewAuthorizer(authCtx).AllowsOrganisationCreateCmd(cmd); err != nil {
		return err
	}

	err := repository.Transactional(ctx, h.db, func(tx *sql.Tx) error {
		changeSet := repository.OrganisationChangeSet{
			ID:   &cmd.OrganisationID,
			Name: &cmd.Name,
		}

		err := repository.InsertOrganisation(ctx, tx, changeSet)
		if err != nil {
			if constraintErr := repository.OrganisationConstraintErr(err); constraintErr != nil {
				return constraintErr
			}
			return errors.Wrap(err, "insert organisation")
		}

		return nil
	})
	if err != nil {
		return errors.Wrap(err, "running transaction")
	}

	logger.InfoContext(ctx, "Created organisation",
		"organisationID", cmd.OrganisationID,
		"organisationName", cmd.Name,
	)

	return nil
}

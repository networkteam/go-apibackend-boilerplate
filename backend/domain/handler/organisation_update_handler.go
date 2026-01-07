package handler

import (
	"context"
	"database/sql"

	"github.com/friendsofgo/errors"
	"github.com/networkteam/slogutils"

	"myvendor.mytld/myproject/backend/domain/command"
	"myvendor.mytld/myproject/backend/domain/types"
	"myvendor.mytld/myproject/backend/persistence/repository"
	"myvendor.mytld/myproject/backend/security/authentication"
	"myvendor.mytld/myproject/backend/security/authorization"
)

func (h *Handler) OrganisationUpdate(ctx context.Context, cmd command.OrganisationUpdateCmd) error {
	logger := slogutils.FromContext(ctx).
		With(
			"component", "handler",
			"handler", "OrganisationUpdate",
		)

	logger.DebugContext(ctx, "Handling organisation update command", "cmd", cmd)

	if err := cmd.Validate(); err != nil {
		return err
	}

	authCtx := authentication.GetAuthContext(ctx)
	if err := authorization.NewAuthorizer(authCtx).AllowsOrganisationUpdateCmd(cmd); err != nil {
		return err
	}

	var prevOrganisationName string
	err := repository.Transactional(ctx, h.db, func(tx *sql.Tx) error {
		prevRecord, err := repository.FindOrganisationByID(ctx, tx, cmd.OrganisationID, nil)
		if errors.Is(err, repository.ErrNotFound) {
			return types.FieldError{
				Field: "organisationId",
				Code:  types.ErrorCodeNotExists,
			}
		} else if err != nil {
			return errors.Wrap(err, "finding organisation")
		}
		prevOrganisationName = prevRecord.Name

		changeSet := repository.OrganisationChangeSet{
			Name: &cmd.Name,
		}

		err = repository.UpdateOrganisation(ctx, tx, cmd.OrganisationID, changeSet)
		if err != nil {
			if constraintErr := repository.OrganisationConstraintErr(err); constraintErr != nil {
				return constraintErr
			}
			return errors.Wrap(err, "update organisation")
		}

		return nil
	})
	if err != nil {
		return errors.Wrap(err, "running transaction")
	}

	logger.InfoContext(ctx, "Updated organisation",
		"organisationID", cmd.OrganisationID,
		"organisationName", cmd.Name,
		"prevOrganisationName", prevOrganisationName,
	)

	return nil
}

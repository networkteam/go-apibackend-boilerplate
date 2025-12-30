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

func (h *Handler) OrganisationDelete(ctx context.Context, cmd command.OrganisationDeleteCmd) error {
	logger := slogutils.FromContext(ctx).
		With(
			"component", "handler",
			"handler", "OrganisationDelete",
		)

	logger.DebugContext(ctx, "Handling organisation delete command", "cmd", cmd)

	authCtx := authentication.GetAuthContext(ctx)
	if err := authorization.NewAuthorizer(authCtx).AllowsOrganisationDeleteCmd(cmd); err != nil {
		return err
	}

	var organisationName string
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
		organisationName = prevRecord.Name

		err = repository.DeleteOrganisation(ctx, tx, cmd.OrganisationID)
		if err != nil {
			return errors.Wrap(err, "deleting organisation")
		}
		return nil
	})
	if err != nil {
		return errors.Wrap(err, "running transaction")
	}

	logger.InfoContext(ctx, "Deleted organisation",
		"organisationID", cmd.OrganisationID,
		"organisationName", organisationName,
	)

	return nil
}

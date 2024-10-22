package handler

import (
	"context"
	"database/sql"

	logger "github.com/apex/log"
	"github.com/friendsofgo/errors"

	"myvendor.mytld/myproject/backend/domain/command"
	"myvendor.mytld/myproject/backend/domain/types"
	"myvendor.mytld/myproject/backend/persistence/repository"
	"myvendor.mytld/myproject/backend/security/authentication"
	"myvendor.mytld/myproject/backend/security/authorization"
)

func (h *Handler) OrganisationDelete(ctx context.Context, cmd command.OrganisationDeleteCmd) error {
	log := logger.FromContext(ctx).
		WithField("component", "handler").
		WithField("handler", "OrganisationDelete")

	log.
		WithField("cmd", cmd).
		Debug("Handling organisation delete command")

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

	log.
		WithField("organisationID", cmd.OrganisationID).
		WithField("organisationName", organisationName).
		Info("Deleted organisation")

	return nil
}

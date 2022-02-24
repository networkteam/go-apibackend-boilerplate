package handler

import (
	"context"
	"database/sql"

	logger "github.com/apex/log"
	"github.com/friendsofgo/errors"

	"myvendor.mytld/myproject/backend/domain"
	"myvendor.mytld/myproject/backend/persistence/repository"
	"myvendor.mytld/myproject/backend/security/authentication"
	"myvendor.mytld/myproject/backend/security/authorization"
)

func (h *Handler) OrganisationUpdate(ctx context.Context, cmd domain.OrganisationUpdateCmd) error {
	log := logger.FromContext(ctx).
		WithField("component", "handler").
		WithField("handler", "OrganisationUpdate")

	log.
		WithField("cmd", cmd).
		Debug("Handling organisation update command")

	if err := cmd.Validate(); err != nil {
		return err
	}

	authCtx := authentication.GetAuthContext(ctx)
	if err := authorization.NewAuthorizer(authCtx).AllowsOrganisationUpdateCmd(cmd); err != nil {
		return err
	}

	var prevOrganisationName string
	err := repository.Transactional(ctx, h.db, func(tx *sql.Tx) error {
		prevRecord, err := repository.FindOrganisationByID(ctx, tx, cmd.OrganisationID)
		if err == repository.ErrNotFound {
			return domain.FieldError{
				Field: "organisationId",
				Code:  domain.ErrorCodeNotExists,
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

	log.
		WithField("organisationID", cmd.OrganisationID).
		WithField("organisationName", cmd.Name).
		WithField("prevOrganisationName", prevOrganisationName).
		Info("Updated organisation")

	return nil
}

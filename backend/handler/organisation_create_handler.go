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

func (h *Handler) OrganisationCreate(ctx context.Context, cmd domain.OrganisationCreateCmd) error {
	log := logger.FromContext(ctx).
		WithField("component", "handler").
		WithField("handler", "OrganisationCreate")

	log.
		WithField("cmd", cmd).
		Debug("Handling organisation create command")

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

	log.
		WithField("organisationID", cmd.OrganisationID).
		WithField("organisationName", cmd.Name).
		Info("Created organisation")

	return nil
}

package handler

import (
	"context"
	"database/sql"

	"github.com/apex/log"
	"github.com/pkg/errors"
	"github.com/zbyte/go-kallax"

	"myvendor/myproject/backend/domain"
	"myvendor/myproject/backend/persistence/records"
	"myvendor/myproject/backend/security/authentication"
	"myvendor/myproject/backend/security/authorization"
)

type OrganisationUpdateHandler struct {
	organisationStore *records.OrganisationStore
}

func NewOrganisationUpdateHandler(db *sql.DB) *OrganisationUpdateHandler {
	return &OrganisationUpdateHandler{
		organisationStore: records.NewOrganisationStore(db),
	}
}

func (h *OrganisationUpdateHandler) Handle(ctx context.Context, cmd domain.OrganisationUpdateCmd) error {
	log.
		WithField("handler", "OrganisationUpdateHandler").
		WithField("cmd", cmd).
		Debug("Handling organisation update command")

	err := cmd.Validate()
	if err != nil {
		return err
	}

	authCtx := authentication.GetAuthContext(ctx)
	if err := authorization.NewAuthorizer(authCtx).AllowsOrganisationUpdateCmd(cmd); err != nil {
		return err
	}

	q := records.NewOrganisationQuery().
		FindByID(kallax.UUID(cmd.OrganisationID))

	organisation, err := h.organisationStore.FindOne(q)
	if err == kallax.ErrNotFound {
		return domain.FieldError{
			// Field named as in GraphQL schema
			Field:     "id",
			Code:      domain.ErrorCodeNotExists,
			Arguments: []string{cmd.OrganisationID.String()},
		}
	} else if err != nil {
		return errors.Wrap(err, "could not query organisation")
	}

	q = records.NewOrganisationQuery().
		Where(
			kallax.And(
				records.LowerCaseEqual(records.Schema.Organisation.Name, cmd.Name),
				kallax.Neq(records.Schema.Organisation.ID, cmd.OrganisationID),
			),
		)
	if n, err := h.organisationStore.Count(q); err != nil {
		return errors.Wrap(err, "could not query organisation")
	} else if n > 0 {
		return domain.FieldError{
			Field:     "name",
			Code:      domain.ErrorCodeAlreadyExists,
			Arguments: []string{cmd.Name},
		}
	}

	organisation.Name = cmd.Name

	_, err = h.organisationStore.Update(organisation)
	if err != nil {
		return errors.Wrap(err, "could not update organisation")
	}

	log.
		WithField("handler", "OrganisationUpdateHandler").
		WithField("organisationID", cmd.OrganisationID).
		Info("Updated organisation")

	return nil
}

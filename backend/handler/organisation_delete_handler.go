package handler

import (
	"context"
	"database/sql"

	"github.com/apex/log"
	"github.com/friendsofgo/errors"
	"github.com/zbyte/go-kallax"

	"myvendor.mytld/myproject/backend/domain"
	"myvendor.mytld/myproject/backend/persistence/records"
	"myvendor.mytld/myproject/backend/security/authentication"
	"myvendor.mytld/myproject/backend/security/authorization"
)

type OrganisationDeleteHandler struct {
	organisationStore *records.OrganisationStore
}

func NewOrganisationDeleteHandler(db *sql.DB) *OrganisationDeleteHandler {
	return &OrganisationDeleteHandler{
		organisationStore: records.NewOrganisationStore(db),
	}
}

func (h *OrganisationDeleteHandler) Handle(ctx context.Context, cmd domain.OrganisationDeleteCmd) error {
	log.
		WithField("handler", "OrganisationDeleteHandler").
		WithField("cmd", cmd).
		Debug("Handling organisation delete command")

	authCtx := authentication.GetAuthContext(ctx)
	if err := authorization.NewAuthorizer(authCtx).AllowsOrganisationDeleteCmd(cmd); err != nil {
		return err
	}

	q := records.NewOrganisationQuery().FindByID(kallax.UUID(cmd.OrganisationID))
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

	err = h.organisationStore.Delete(organisation)
	if err != nil {
		return errors.Wrap(err, "could not delete organisation")
	}

	log.
		WithField("handler", "OrganisationDeleteHandler").
		WithField("organisationID", cmd.OrganisationID).
		Info("Deleted organisation")

	return nil
}

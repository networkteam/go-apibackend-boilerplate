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

type OrganisationCreateHandler struct {
	organisationStore *records.OrganisationStore
}

func NewOrganisationCreateHandler(db *sql.DB) *OrganisationCreateHandler {
	return &OrganisationCreateHandler{
		organisationStore: records.NewOrganisationStore(db),
	}
}

func (h *OrganisationCreateHandler) Handle(ctx context.Context, cmd domain.OrganisationCreateCmd) error {
	log.
		WithField("handler", "OrganisationCreateHandler").
		WithField("cmd", cmd).
		Debug("Handling organisation create command")

	err := cmd.Validate()
	if err != nil {
		return err
	}

	authCtx := authentication.GetAuthContext(ctx)
	if err := authorization.NewAuthorizer(authCtx).AllowsOrganisationCreateCmd(cmd); err != nil {
		return err
	}

	q := records.NewOrganisationQuery().Where(
		records.LowerCaseEqual(records.Schema.Organisation.Name, cmd.Name),
	)
	if n, err := h.organisationStore.Count(q); err != nil {
		return errors.Wrap(err, "could not query organisations")
	} else if n > 0 {
		return domain.FieldError{
			Field:     "name",
			Code:      domain.ErrorCodeAlreadyExists,
			Arguments: []string{cmd.Name},
		}
	}

	organisation := records.NewOrganisation()
	organisation.ID = kallax.UUID(cmd.OrganisationID)
	organisation.Name = cmd.Name

	err = h.organisationStore.Insert(organisation)
	if err != nil {
		return errors.Wrap(err, "could not insert organisation")
	}

	log.
		WithField("handler", "OrganisationCreateHandler").
		WithField("organisationID", cmd.OrganisationID).
		Info("Created organisation")

	return nil
}

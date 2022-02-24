package domain

import (
	"time"

	"github.com/gofrs/uuid"
)

type Employee struct {
	ID             uuid.UUID `read_col:"employees.employee_id" write_col:"employee_id"`
	OrganisationID uuid.UUID `read_col:"employees.organisation_id" write_col:"organisation_id"`
	Firstname      string    `read_col:"employees.firstname,sortable" write_col:"firstname"`
	Lastname       string    `read_col:"employees.lastname,sortable" write_col:"lastname"`

	CreatedAt time.Time `read_col:"employees.created_at,sortable"`
	UpdatedAt time.Time `read_col:"employees.updated_at,sortable"`

	// Joined from account
	Username string `read_col:"accounts.username"`
}

func (e Employee) FullName() string {
	return e.Firstname + " " + e.Lastname
}

type EmployeeFilter struct {
	IDs            []uuid.UUID
	Q              *string
	OrganisationID *uuid.UUID
}

// SetOrganisationID implements OrganisationIDSetter
func (f *EmployeeFilter) SetOrganisationID(organisationID *uuid.UUID) {
	f.OrganisationID = organisationID
}

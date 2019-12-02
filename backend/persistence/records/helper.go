package records

import "github.com/zbyte/go-kallax"

// LowerCaseEqual is a kallax operator to use like
// query.Where(persistence.LowerCaseEqual(SchemaField, value))
// both values are lower cased by this operator, no need to pass the
// argument lower cased
var LowerCaseEqual = kallax.NewOperator("LOWER(:col:) = LOWER(:arg:)")

func valueToUUID(value interface{}, err error) *kallax.UUID {
	if err != nil {
		return nil
	}
	if id, ok := value.(*kallax.UUID); ok {
		return id
	}
	return nil
}

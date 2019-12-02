package api

import (
	"io"
	"strconv"

	"github.com/99designs/gqlgen/graphql"
	"github.com/friendsofgo/errors"
	"github.com/gofrs/uuid"
)

func MarshalUUIDScalar(value uuid.UUID) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		// If uuid is empty it should be marshalled to the empty string
		if value == uuid.Nil {
			_, _ = w.Write([]byte(strconv.Quote("")))
		} else {
			_, _ = w.Write([]byte(strconv.Quote(value.String())))
		}
	})
}

func UnmarshalUUIDScalar(v interface{}) (uuid.UUID, error) {
	switch v := v.(type) {
	case string:
		return uuid.FromString(v)
	default:
		return uuid.Nil, errors.Errorf("%T is not a string", v)
	}
}

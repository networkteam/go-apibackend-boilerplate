package model

import (
	"io"
	"strconv"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/friendsofgo/errors"
)

func MarshalDateTimeScalar(value time.Time) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		_, _ = w.Write([]byte(strconv.Quote(value.Format(time.RFC3339)))) //nolint:errcheck
	})
}

func UnmarshalDateTimeScalar(v any) (time.Time, error) {
	switch v := v.(type) {
	case string:
		return time.Parse(time.RFC3339, v)
	default:
		return time.Time{}, errors.Errorf("%T is not a string", v)
	}
}

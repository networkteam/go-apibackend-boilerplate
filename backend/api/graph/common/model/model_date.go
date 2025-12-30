package model

import (
	"io"
	"strconv"

	"github.com/99designs/gqlgen/graphql"
	"github.com/friendsofgo/errors"

	"myvendor.mytld/myproject/backend/domain/types"
)

func MarshalDateScalar(value types.Date) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		_, _ = w.Write([]byte(strconv.Quote(value.String()))) //nolint:errcheck
	})
}

func UnmarshalDateScalar(v any) (types.Date, error) {
	dateString, ok := v.(string)
	if !ok {
		return types.Date{}, errors.Errorf("%T is not a string", v)
	}

	d, err := types.ParseDate(dateString)
	if err != nil {
		return d, err
	}
	return d, nil
}

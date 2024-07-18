package model

import (
	"io"
	"strconv"

	"github.com/99designs/gqlgen/graphql"
	"github.com/friendsofgo/errors"

	"myvendor.mytld/myproject/backend/domain"
)

func MarshalDateScalar(value domain.Date) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		_, _ = w.Write([]byte(strconv.Quote(value.String()))) //nolint:errcheck
	})
}

func UnmarshalDateScalar(v any) (domain.Date, error) {
	dateString, ok := v.(string)
	if !ok {
		return domain.Date{}, errors.Errorf("%T is not a string", v)
	}

	d, err := domain.ParseDate(dateString)
	if err != nil {
		return d, err
	}
	return d, nil
}

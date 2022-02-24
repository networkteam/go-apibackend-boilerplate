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
		_, _ = w.Write([]byte(strconv.Quote(value.String())))
	})
}

func UnmarshalDateScalar(v interface{}) (domain.Date, error) {
	switch v := v.(type) {
	case string:
		d, err := domain.ParseDate(v)
		if err != nil {
			return d, err
		}
		return d, nil
	default:
		return domain.Date{}, errors.Errorf("%T is not a string", v)
	}
}

package handler

import (
	"context"
	"errors"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/gqlerror"

	"myvendor.mytld/myproject/backend/domain"
)

type extendedError interface {
	Extensions() map[string]any
}

func ErrorPresenter(ctx context.Context, err error) *gqlerror.Error {
	graphqlErr := graphql.DefaultErrorPresenter(ctx, err)

	// A field resolvable error should be unwrapped before presenting
	var fieldErr domain.FieldResolvableError
	if errors.As(err, &fieldErr) {
		graphqlErr.Message = fieldErr.Error()
	}

	// Extended errors can add structured information (extensions) to the GraphQL error
	var extendedErr extendedError
	if errors.As(err, &extendedErr) {
		graphqlErr.Extensions = extendedErr.Extensions()
	}

	return graphqlErr
}

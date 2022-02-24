package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"myvendor.mytld/myproject/backend/api/graph/generated"
)

func (r *queryResolver) Echo(ctx context.Context, hello string) (string, error) {
	return fmt.Sprintf("Hello, %s", hello), nil
}

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type queryResolver struct{ *Resolver }

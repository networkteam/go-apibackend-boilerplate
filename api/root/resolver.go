package root

import (
	"myvendor/myproject/backend/api"
	"myvendor/myproject/backend/api/authentication"
)

// Resolver implements the GraphQL resolver for queries, mutations and field specific resolvers
type Resolver struct {
	api.ResolverDependencies
}

var _ api.ResolverRoot = new(Resolver)

func (r *Resolver) Mutation() api.MutationResolver {
	return &mutationResolver{
		ResolverDependencies:          r.ResolverDependencies,
		authenticationMutationResolver: &authentication.MutationResolver{r.ResolverDependencies},
	}
}

func (r *Resolver) Query() api.QueryResolver {
	return &queryResolver{
		ResolverDependencies: r.ResolverDependencies,
		authenticationQueryResolver:   &authentication.QueryResolver{r.ResolverDependencies},
	}
}

// Use type aliasing to prevent issues with duplicate embedded struct field names
type authenticationMutationResolver = authentication.MutationResolver

type mutationResolver struct {
	api.ResolverDependencies
	*authenticationMutationResolver
}

type authenticationQueryResolver = authentication.QueryResolver

type queryResolver struct {
	api.ResolverDependencies
	*authenticationQueryResolver
}

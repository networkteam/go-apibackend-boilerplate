package root

import (
	"myvendor.mytld/myproject/backend/api"
	"myvendor.mytld/myproject/backend/api/authentication"
)

// Resolver implements the GraphQL resolver for queries, mutations and field specific resolvers
type Resolver struct {
	api.ResolverDependencies
}

var _ api.ResolverRoot = new(Resolver)

func (r *Resolver) Mutation() api.MutationResolver {
	return &mutationResolver{
		ResolverDependencies:           r.ResolverDependencies,
		authenticationMutationResolver: &authentication.MutationResolver{r.ResolverDependencies},
	}
}

func (r *Resolver) Query() api.QueryResolver {
	return &queryResolver{
		ResolverDependencies:        r.ResolverDependencies,
		authenticationQueryResolver: &authentication.QueryResolver{r.ResolverDependencies},
	}
}

// Sub-resolver for specific types

func (r *Resolver) AppAccount() api.AppAccountResolver {
	return &authentication.AppAccountResolver {

	}
}

func (r *Resolver) UserAccount() api.UserAccountResolver {
	panic("implement me")
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

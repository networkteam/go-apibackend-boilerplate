package graph

import (
	"myvendor.mytld/myproject/backend/api"
	"myvendor.mytld/myproject/backend/finder"
	"myvendor.mytld/myproject/backend/handler"
)

type Resolver struct {
	api.ResolverDependencies
	api.ResolverConfig

	handler *handler.Handler
	finder  *finder.Finder
}

func NewResolver(deps api.ResolverDependencies, config api.ResolverConfig) *Resolver {
	return &Resolver{
		ResolverDependencies: deps,
		ResolverConfig:       config,
		handler: handler.NewHandler(deps.DB, deps.Config, handler.Deps{
			TimeSource:    deps.TimeSource,
			Mailer:        deps.Mailer,
			MeterProvider: deps.MeterProvider,
		}),
		finder: finder.NewFinder(deps.DB, deps.TimeSource),
	}
}

package handler

import (
	"context"

	"github.com/getsentry/sentry-go"
)

type cronJob struct {
	sentryHub *sentry.Hub
}

// Init must be called when building the cron job from the main goroutine
func (c *cronJob) init() {
	// Build a goroutine local Sentry hub
	c.sentryHub = sentry.CurrentHub().Clone()
	c.sentryHub.Scope().SetTag("section", "cron")
}

func (c cronJob) context() context.Context {
	ctx := context.Background()
	ctx = sentry.SetHubOnContext(ctx, c.sentryHub)
	return ctx
}

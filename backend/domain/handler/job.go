package handler

import (
	"context"
	"errors"

	"github.com/getsentry/sentry-go"
	"github.com/networkteam/slogutils"
	"github.com/robfig/cron"

	"myvendor.mytld/myproject/backend/domain/types"
)

func CommandHandlerJob[T any](ctx context.Context, commandHandler func(ctx context.Context, cmd T) error, cmd T) cron.Job {
	return cron.FuncJob(func() {
		logger := slogutils.FromContext(ctx)

		hub := sentry.GetHubFromContext(ctx)
		if hub == nil {
			hub = sentry.CurrentHub()
		}
		hub = hub.Clone()
		hub.Scope().SetTag("section", "cron")
		ctx = sentry.SetHubOnContext(ctx, hub)

		defer func(ctx context.Context) {
			if recovered := recover(); recovered != nil {
				// Logged error will be sent to Sentry via the slog integration
				logger.ErrorContext(ctx,
					"Panic while running job",
					slogutils.ErrorKey, recovered,
				)
			}
		}(ctx)

		err := commandHandler(ctx, cmd)
		if err != nil {
			if errors.Is(err, types.ErrHandlerLocked) {
				logger.WarnContext(ctx,
					"Cron job skipped because handler is locked",
					"cmd", cmd,
				)
				return
			}

			logger.ErrorContext(ctx,
				"Cron job failed",
				"cmd", cmd,
				slogutils.Err(err),
			)
		}
	})
}

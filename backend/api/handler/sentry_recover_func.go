package handler

import (
	"context"

	"github.com/friendsofgo/errors"
	"github.com/getsentry/sentry-go"
	"github.com/networkteam/slogutils"
)

func sentryRecoverFunc(ctx context.Context, err any) error {
	var newErr error
	if realErr, ok := err.(error); ok {
		newErr = realErr
	} else {
		newErr = errors.Errorf("%s", err)
	}

	var hub *sentry.Hub
	if sentry.HasHubOnContext(ctx) {
		hub = sentry.GetHubFromContext(ctx)
	} else {
		hub = sentry.CurrentHub()
	}

	logger := slogutils.FromContext(ctx)

	eID := hub.RecoverWithContext(ctx, err)
	if eID != nil {
		logger.ErrorContext(ctx, "Recovered panic and captured with Sentry",
			"sentryEventId", *eID,
			slogutils.Err(newErr))
	} else {
		// Let's assume no event ID means no Sentry configured (e.g. in development)
		newErr = errors.WithStack(newErr)

		logger.ErrorContext(ctx, "Recovered panic", slogutils.Err(newErr))
	}

	return newErr
}

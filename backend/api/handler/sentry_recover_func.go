package handler

import (
	"context"

	"github.com/apex/log"
	"github.com/friendsofgo/errors"
	"github.com/getsentry/sentry-go"
)

func sentryRecoverFunc(ctx context.Context, err interface{}) error {
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

	eID := hub.RecoverWithContext(ctx, err)
	if eID != nil {
		log.
			WithError(newErr).
			WithField("sentryEventId", *eID).
			Errorf("Recovered panic and captured with Sentry")
	} else {
		// Let's assume no event ID means no Sentry configured (e.g. in development)
		newErr = errors.WithStack(newErr)

		log.
			WithError(newErr).
			Errorf("Recovered panic: %+v", newErr)
	}

	return newErr
}

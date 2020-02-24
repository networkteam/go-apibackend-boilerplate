package notification

import (
	"context"

	"github.com/apex/log"

	"myvendor.mytld/myproject/backend/logger"
)

func NewLoggingNotifier() *LoggingNotifier {
	log.Warn("=====================================")
	log.Warn(" Running with logging notifier only! ")
	log.Warn("=====================================")

	return &LoggingNotifier{}
}

type LoggingNotifier struct {
}

var _ Notifier = new(LoggingNotifier)

func (n *LoggingNotifier) Notify(ctx context.Context, registration DeviceRegistrationProvider, payload PayloadProvider) error {
	log := logger.GetLogger(ctx).
		WithField("component", "loggingNotifier")

	log.
		WithField("deviceToken", registration.GetDeviceToken()).
		WithField("deviceOS", registration.GetDeviceOS()).
		WithField("message", payload.GetMessage()).
		WithField("data", payload.GetData()).
		Info("Request to send notification")

	return nil
}

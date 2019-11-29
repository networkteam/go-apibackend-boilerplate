package notification

import (
	"github.com/apex/log"
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

func (n *LoggingNotifier) Notify(registration DeviceRegistrationProvider, payload PayloadProvider) error {
	log.
		WithField("deviceToken", registration.GetDeviceToken()).
		WithField("deviceOS", registration.GetDeviceOS()).
		WithField("message", payload.GetMessage()).
		WithField("data", payload.GetData()).
		Info("Request to send notification")

	return nil
}

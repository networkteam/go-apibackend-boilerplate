package notification

import (
	"context"
	"testing"

	"github.com/apex/log"
)

type TestNotifier struct {
	Assertion func(t *testing.T, registration DeviceRegistrationProvider, payload PayloadProvider) error
	t         *testing.T
	callCount int
}

func NewTestNotifier(t *testing.T) *TestNotifier {
	return &TestNotifier{
		t: t,
	}
}

var _ Notifier = new(TestNotifier)

func (n *TestNotifier) Notify(ctx context.Context, registration DeviceRegistrationProvider, payload PayloadProvider) (err error) {
	n.callCount++
	if n.Assertion == nil || n.t == nil {
		log.WithField("registration", registration).
			WithField("payload", payload).
			Debug("received notification")
	} else {
		err = n.Assertion(n.t, registration, payload)
	}

	return
}

func (n *TestNotifier) CallCount() int {
	return n.callCount
}

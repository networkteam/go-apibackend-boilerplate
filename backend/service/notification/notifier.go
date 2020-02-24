package notification

import "context"

type Notifier interface {
	Notify(ctx context.Context, registration DeviceRegistrationProvider, payload PayloadProvider) error
}

type PayloadProvider interface {
	GetMessage() string
	GetData() interface{}
}

type DeviceRegistrationProvider interface {
	GetDeviceToken() string
	GetDeviceOS() string
}

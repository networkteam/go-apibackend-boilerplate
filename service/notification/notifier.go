package notification

type Notifier interface {
	Notify(registration DeviceRegistrationProvider, payload PayloadProvider) error
}

type PayloadProvider interface {
	GetMessage() string
	GetData() interface{}
}

type DeviceRegistrationProvider interface {
	GetDeviceToken() string
	GetDeviceOS() string
}

package sentry

import "github.com/getsentry/sentry-go"

func NewHubMock() (*sentry.Hub, *TransportMock) {
	transportMock := &TransportMock{}
	client, err := sentry.NewClient(sentry.ClientOptions{
		Transport: transportMock,
	})
	if err != nil {
		panic(err)
	}
	mockHub := sentry.NewHub(client, sentry.NewScope())
	return mockHub, transportMock
}

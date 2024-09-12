package handler

import "go.opentelemetry.io/otel"

//nolint:gochecknoglobals
var meter = otel.Meter("myvendor.mytld/myproject/backend/handler")

func mustInstrument[T any](instrument T, err error) T {
	if err != nil {
		panic(err)
	}
	return instrument
}

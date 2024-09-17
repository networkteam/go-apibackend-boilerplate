package handler

import (
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
)

func mustInstrument[T any](instrument T, err error) T {
	if err != nil {
		panic(err)
	}
	return instrument
}

type instrumentation struct {
	loginSuccessCounter metric.Int64Counter
	loginFailedCounter  metric.Int64Counter
}

func initInstrumentation(provider metric.MeterProvider) instrumentation {
	if provider == nil {
		provider = noop.NewMeterProvider()
	}

	meter := provider.Meter("myvendor.mytld/myproject/backend/handler")

	return instrumentation{
		loginSuccessCounter: mustInstrument(meter.Int64Counter(
			"login.success.counter",
			metric.WithDescription("Number of successful logins."),
			metric.WithUnit("{call}"),
		)),
		loginFailedCounter: mustInstrument(meter.Int64Counter(
			"login.failed.counter",
			metric.WithDescription("Number of failed logins."),
			metric.WithUnit("{call}"),
		)),
	}
}

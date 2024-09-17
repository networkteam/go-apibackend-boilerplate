package telemetry

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func SetupTestMeter(t *testing.T) (*sdkmetric.ManualReader, metric.MeterProvider) {
	t.Helper()

	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))

	t.Cleanup(func() {
		err := provider.Shutdown(context.Background())
		assert.NoError(t, err)
	})

	return reader, provider
}

func AssertMeterCounter(t *testing.T, reader sdkmetric.Reader, scope, name string, want int64) {
	t.Helper()

	metricsData := metricdata.ResourceMetrics{}
	err := reader.Collect(context.Background(), &metricsData)
	require.NoError(t, err)

	scopeMetrics, found := find(metricsData.ScopeMetrics, func(m metricdata.ScopeMetrics) bool {
		return m.Scope.Name == scope
	})
	if !found {
		t.Fatalf("metrics for scope %q not found", scope)
	}
	metrics, found := find(scopeMetrics.Metrics, func(m metricdata.Metrics) bool {
		return m.Name == name
	})
	if !found {
		t.Fatalf("metrics for name %q not found", name)
	}

	agg, found := metrics.Data.(metricdata.Sum[int64])
	if !found {
		t.Fatalf("metrics for name %q is not a counter", name)
	}
	var sum int64
	for _, dp := range agg.DataPoints {
		sum += dp.Value
	}

	assert.Equal(t, want, sum, "sum of metric %q in scope %q", name, scope)
}

func find[T any](slice []T, predicate func(T) bool) (T, bool) {
	for _, item := range slice {
		if predicate(item) {
			return item, true
		}
	}
	var zero T
	return zero, false
}

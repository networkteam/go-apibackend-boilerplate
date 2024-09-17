package telemetry

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metric2 "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func SetupTestMeter(t *testing.T) (*metric.ManualReader, metric2.MeterProvider) {
	reader := metric.NewManualReader()
	provider := metric.NewMeterProvider(metric.WithReader(reader))

	t.Cleanup(func() {
		err := provider.Shutdown(context.Background())
		assert.NoError(t, err)
	})

	return reader, provider
}

func AssertMeterCounter(t *testing.T, reader metric.Reader, scope, name string, want int64) {
	t.Helper()

	metricsData := metricdata.ResourceMetrics{}
	err := reader.Collect(context.Background(), &metricsData)
	require.NoError(t, err)

	var scopeMetrics *metricdata.ScopeMetrics
	for _, scopeMetrics = range metricsData.ScopeMetrics {
		if scopeMetrics.Scope.Name == scope {
			break
		}
	}
	if scopeMetrics == nil {
		t.Fatalf("metrics for scope %q not found", scope)
	}

	var metrics *metricdata.Metrics
	for _, metrics = range scopeMetrics.Metrics {
		if metrics.Name == name {
			break
		}
	}
	if metrics == nil {
		t.Fatalf("metrics for name %q not found", name)
	}

	agg, ok := metrics.Data.(metricdata.Sum[int64])
	if !ok {
		t.Fatalf("metrics for name %q is not a counter", name)
	}
	var sum int64
	for _, dp := range agg.DataPoints {
		sum += dp.Value
	}

	assert.Equal(t, want, sum, "sum of metric %q in scope %q", name, scope)
}

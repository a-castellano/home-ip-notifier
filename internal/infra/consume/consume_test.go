//go:build integration_tests || unit_tests

package consume

import (
	"context"
	"errors"
	"testing"

	envelope "github.com/a-castellano/go-types/types/envelope"
	"go.opentelemetry.io/otel"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

type mockProcessor struct {
	fail bool
}

func (m mockProcessor) ProcessMessage(ctx context.Context, message string) error {
	if m.fail {
		return errors.New("Fail")
	}
	return nil
}

func TestEmptyData(t *testing.T) {

	ctx := context.Background()
	consumer := NewConsumer(ctx, "testqueue", mockProcessor{})

	emptydata := make([]byte, 0)

	err := consumer.Consume(ctx, emptydata)

	if err != nil {
		t.Fatalf("TestEmptyData should not return error even is data to consume is empty, error was \"%s\"", err.Error())
	}

}

func TestEmptyBody(t *testing.T) {

	ctx := context.Background()
	consumer := NewConsumer(ctx, "testqueue", mockProcessor{})
	emptybody := make([]byte, 0)
	carrier := map[string]string{"testkey": "testValue"}

	data, _ := (&envelope.Envelope{Carrier: carrier, Body: emptybody}).Marshal()

	err := consumer.Consume(ctx, data)

	if err != nil {
		t.Fatalf("TestEmptyBody should not return error even is body to consume is empty, error was \"%s\"", err.Error())
	}

}

func TestValidEnvelopeConsumerFails(t *testing.T) {

	ctx := context.Background()
	consumer := NewConsumer(ctx, "testqueue", mockProcessor{fail: true})

	body := []byte("123.123.123.123")
	carrier := map[string]string{"traceparent": "00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01"}

	data, _ := (&envelope.Envelope{Carrier: carrier, Body: body}).Marshal()

	err := consumer.Consume(ctx, data)

	if err == nil {
		t.Fatalf("TestValidEnvelopeConsumerFails should fail")
	}

	expectedErr := "Fail"
	if err.Error() != expectedErr {
		t.Fatalf("TestValidEnvelopeConsumerFails error should be \"%s\" but it was \"%s\"", expectedErr, err.Error())
	}

}

func TestValidEnvelopeConsumer(t *testing.T) {

	ctx := context.Background()
	consumer := NewConsumer(ctx, "testqueue", mockProcessor{})

	body := []byte("123.123.123.123")
	carrier := map[string]string{"traceparent": "00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01"}

	data, _ := (&envelope.Envelope{Carrier: carrier, Body: body}).Marshal()

	err := consumer.Consume(ctx, data)

	if err != nil {
		t.Fatalf("TestValidEnvelopeConsumer should not fail, error was \"%s\"", err.Error())
	}

}

// consumeOutcomes are the four values the outcome attribute can take, one per
// Consume exit path.
var consumeOutcomes = []string{"success", "error", "malformed", "empty"}

// collectConsumeMetrics runs one delivery per Consume exit path and returns
// the collected metrics of the consume scope. It installs an SDK
// MeterProvider backed by a ManualReader as the global one and restores the
// previous provider on cleanup so later tests keep their no-op default; the
// consumers are built after installing it because the instruments are created
// in NewConsumer. Both consumers write to the same aggregation streams (same
// scope, name, unit and description), so the four deliveries surface as four
// data points of the same two metrics, one per outcome value.
// This helper was written by an AI agent (Claude).
func collectConsumeMetrics(t *testing.T) metricdata.ScopeMetrics {
	t.Helper()

	previousMeterProvider := otel.GetMeterProvider()
	meterReader := sdkmetric.NewManualReader()
	otel.SetMeterProvider(sdkmetric.NewMeterProvider(sdkmetric.WithReader(meterReader)))
	t.Cleanup(func() { otel.SetMeterProvider(previousMeterProvider) })

	ctx := context.Background()
	consumer := NewConsumer(ctx, "testqueue", mockProcessor{})
	failingConsumer := NewConsumer(ctx, "testqueue", mockProcessor{fail: true})

	validData, _ := (&envelope.Envelope{Carrier: map[string]string{}, Body: []byte("123.123.123.123")}).Marshal()
	emptyBodyData, _ := (&envelope.Envelope{Carrier: map[string]string{}, Body: []byte{}}).Marshal()

	// One delivery per outcome; the returned errors are the concern of the
	// path tests above, not of the metric tests.
	_ = consumer.Consume(ctx, validData)                 // outcome success
	_ = failingConsumer.Consume(ctx, validData)          // outcome error
	_ = consumer.Consume(ctx, []byte("not an envelope")) // outcome malformed
	_ = consumer.Consume(ctx, emptyBodyData)             // outcome empty

	// Until Collect is called, Add and Record have only updated the SDK's
	// internal aggregation state; the ManualReader pulls it into
	// ResourceMetrics.
	var collectedMetrics metricdata.ResourceMetrics
	if err := meterReader.Collect(ctx, &collectedMetrics); err != nil {
		t.Fatalf("Collect should not fail, error was %q", err.Error())
	}

	if len(collectedMetrics.ScopeMetrics) != 1 {
		t.Fatalf("expected 1 scope, got %d", len(collectedMetrics.ScopeMetrics))
	}
	scopeMetrics := collectedMetrics.ScopeMetrics[0]
	if scopeMetrics.Scope.Name != componentName {
		t.Errorf("expected scope name %q, got %q", componentName, scopeMetrics.Scope.Name)
	}

	return scopeMetrics
}

// locateMetric finds a collected metric by name; the metrics are located by
// name because an instrument with no measurements emits no data points, so
// positions are not stable.
// This helper was written by an AI agent (Claude).
func locateMetric(scopeMetrics metricdata.ScopeMetrics, name string) *metricdata.Metrics {
	for i, collectedMetric := range scopeMetrics.Metrics {
		if collectedMetric.Name == name {
			return &scopeMetrics.Metrics[i]
		}
	}
	return nil
}

// TestConsumedMessagesCounter checks that every Consume exit path counts its
// delivery exactly once, tagged with the outcome that path decides.
// This test was written by an AI agent (Claude).
func TestConsumedMessagesCounter(t *testing.T) {

	scopeMetrics := collectConsumeMetrics(t)

	counterMetric := locateMetric(scopeMetrics, "homeipnotifier.messages.consumed")
	if counterMetric == nil {
		t.Fatal("expected metric homeipnotifier.messages.consumed to be collected")
	}
	if counterMetric.Unit != "{message}" {
		t.Errorf("expected counter unit %q, got %q", "{message}", counterMetric.Unit)
	}

	sum, ok := counterMetric.Data.(metricdata.Sum[int64])
	if !ok {
		t.Fatalf("expected counter data of type Sum[int64], got %T", counterMetric.Data)
	}
	if !sum.IsMonotonic {
		t.Error("expected a monotonic sum")
	}
	if len(sum.DataPoints) != len(consumeOutcomes) {
		t.Fatalf("expected %d counter data points (one per outcome), got %d", len(consumeOutcomes), len(sum.DataPoints))
	}

	countsByOutcome := make(map[string]int64)
	for _, dataPoint := range sum.DataPoints {
		outcome, found := dataPoint.Attributes.Value("outcome")
		if !found {
			t.Fatal("expected every counter data point to carry the outcome attribute")
		}
		countsByOutcome[outcome.AsString()] = dataPoint.Value
	}
	for _, expectedOutcome := range consumeOutcomes {
		if countsByOutcome[expectedOutcome] != 1 {
			t.Errorf("expected 1 message counted with outcome %q, got %d", expectedOutcome, countsByOutcome[expectedOutcome])
		}
	}

}

// TestProcessingDurationHistogram checks that every Consume exit path records
// its handling duration exactly once, tagged with the outcome that path
// decides.
// This test was written by an AI agent (Claude).
func TestProcessingDurationHistogram(t *testing.T) {

	scopeMetrics := collectConsumeMetrics(t)

	histogramMetric := locateMetric(scopeMetrics, "homeipnotifier.message.processing.duration")
	if histogramMetric == nil {
		t.Fatal("expected metric homeipnotifier.message.processing.duration to be collected")
	}
	if histogramMetric.Unit != "s" {
		t.Errorf("expected histogram unit %q, got %q", "s", histogramMetric.Unit)
	}

	histogram, ok := histogramMetric.Data.(metricdata.Histogram[float64])
	if !ok {
		t.Fatalf("expected histogram data of type Histogram[float64], got %T", histogramMetric.Data)
	}
	if len(histogram.DataPoints) != len(consumeOutcomes) {
		t.Fatalf("expected %d histogram data points (one per outcome), got %d", len(consumeOutcomes), len(histogram.DataPoints))
	}

	for _, dataPoint := range histogram.DataPoints {
		outcome, found := dataPoint.Attributes.Value("outcome")
		if !found {
			t.Fatal("expected every histogram data point to carry the outcome attribute")
		}
		if dataPoint.Count != 1 {
			t.Errorf("expected 1 duration recorded with outcome %q, got %d", outcome.AsString(), dataPoint.Count)
		}
	}

}

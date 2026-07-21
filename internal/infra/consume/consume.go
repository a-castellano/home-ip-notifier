// Package consume is the inbound adapter of the service: it unwraps the
// envelopes received from the message broker, continues the distributed trace
// they carry and hands the payload to the use case.
package consume

import (
	"context"
	"errors"
	logger "github.com/a-castellano/go-services/infra/logger"
	opentelemetry "github.com/a-castellano/go-services/infra/opentelemetry"
	envelope "github.com/a-castellano/go-types/types/envelope"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"time"
)

const componentName = "github.com/a-castellano/home-ip-notifier/internal/infra/consume"

// Processor is what the consumer needs from the use case; app.Announcer
// satisfies it. The interface is declared here, where it is consumed (Go
// idiom), so the adapter can be tested against a fake.
type Processor interface {
	ProcessMessage(ctx context.Context, message string) error
}

// Consumer turns the raw deliveries of one queue into use-case calls.
type Consumer struct {
	processor        Processor
	queue            string
	consumedMessages metric.Int64Counter
	processDuration  metric.Float64Histogram
}

// NewConsumer returns a Consumer for queue that delegates to processor. It
// also creates the Consumer's metric instruments against the global
// MeterProvider; the context is only used to reach the logger during
// construction, it is not stored. A failed instrument registration is logged
// and otherwise ignored — the returned instrument is usable anyway, and
// telemetry must never prevent the consumer from being built.
func NewConsumer(ctx context.Context, queue string, processor Processor) Consumer {

	log := logger.FromContext(ctx).With("operation", "NewConsumer")
	log.DebugContext(ctx, "creating new consumer")

	otelMeter := otel.Meter(componentName)

	// define metrics
	consumedMessages, consumedMessagesErr := otelMeter.Int64Counter(
		"homeipnotifier.messages.consumed",
		metric.WithDescription("Number of consumed messages"),
		metric.WithUnit("{message}"),
	)
	if consumedMessagesErr != nil {
		log.ErrorContext(ctx, "cannot register homeipnotifier.messages.consumed otel meter", "error", consumedMessagesErr)
	}

	processDuration, processDurationErr := otelMeter.Float64Histogram(
		"homeipnotifier.message.processing.duration",
		metric.WithDescription("Duration of the message processing"),
		metric.WithUnit("s"),
	)
	if processDurationErr != nil {
		log.ErrorContext(ctx, "cannot register homeipnotifier.message.processing.duration otel meter", "error", processDurationErr)
	}

	return Consumer{processor: processor, queue: queue, consumedMessages: consumedMessages, processDuration: processDuration}
}

// Consume handles one delivery. Every path opens a CONSUMER span so failures
// are visible in the trace backend, not only in logs: a valid envelope joins
// the producer's trace as a child of the context it carries; a malformed one
// has no context to extract, so its span is a local root trace marking the
// poisoned message. An empty body also joins the producer's trace (the
// envelope itself parsed fine) with the error recorded on the span. Malformed
// envelopes and empty bodies are then dropped (nil is returned): consumption
// is auto-ack, so failing would not requeue them. For valid envelopes it
// returns whatever the use case returns.
//
// Every delivery — dropped and failed ones included — is also counted once
// and its handling duration recorded once, both tagged with an outcome
// attribute (success, malformed, empty or error), so error rates can be
// derived from the counter.
func (c Consumer) Consume(ctx context.Context, receivedData []byte) error {

	start := time.Now()

	outcome := "success"

	// The closure is required: a plain deferred call would evaluate its
	// arguments right here, freezing time.Since at ~0 and outcome at
	// "success" instead of the value the exit path decides. It reads the ctx
	// reassigned by Start below, so both exemplars link to this delivery's
	// CONSUMER span — the SpanContext survives span.End, which runs first
	// (LIFO). Renaming the span's ctx would silently break that link.
	defer func() {
		outcomeAttribute := metric.WithAttributes(attribute.String("outcome", outcome))

		c.consumedMessages.Add(ctx, 1, outcomeAttribute)
		c.processDuration.Record(ctx, time.Since(start).Seconds(), outcomeAttribute)
	}()

	log := logger.FromContext(ctx).With("operation", "consume")

	log.DebugContext(ctx, "unmarshaling envelope from received data")

	receivedEnvelope, unmarshalErr := envelope.Unmarshal(receivedData)
	if unmarshalErr == nil {
		// Valid envelope: join the producer's trace before starting the span.
		ctx = opentelemetry.Extract(ctx, receivedEnvelope)
	}

	// Started in every path: with a valid envelope it is a child of the
	// remote context; with a malformed one there is nothing to extract and
	// it becomes a local root trace.
	ctx, span := otel.Tracer(componentName).Start(ctx, "process "+c.queue,
		trace.WithSpanKind(trace.SpanKindConsumer),
		trace.WithAttributes(
			attribute.String("messaging.destination.name", c.queue),
			attribute.String("messaging.operation.type", "process"),
		))
	defer span.End()

	if unmarshalErr != nil {
		// Deepest (and only) span of this path: event and status here.
		span.RecordError(unmarshalErr)
		span.SetStatus(codes.Error, "cannot unmarshal envelope")
		log.ErrorContext(ctx, "cannot unmarshal data", "error", unmarshalErr.Error())
		outcome = "malformed"
		return nil
	}

	if len(receivedEnvelope.Body) == 0 {
		errEmptyBody := errors.New("received body is empty")
		span.RecordError(errEmptyBody)
		span.SetStatus(codes.Error, errEmptyBody.Error())
		log.ErrorContext(ctx, errEmptyBody.Error())
		outcome = "empty"
		return nil
	}

	plainMessage := string(receivedEnvelope.Body)

	log.DebugContext(ctx, "process message", "message", plainMessage)

	processErr := c.processor.ProcessMessage(ctx, plainMessage)

	if processErr != nil {
		// Status only: the error event and the log are already
		// recorded closest to the point of error
		span.SetStatus(codes.Error, "message process has failed")
		outcome = "error"
		return processErr
	}

	return nil
}

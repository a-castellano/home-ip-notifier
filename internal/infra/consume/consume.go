// Package consume is the inbound adapter of the service: it unwraps the
// envelopes received from the message broker, continues the distributed trace
// they carry and hands the payload to the use case.
package consume

import (
	"context"

	logger "github.com/a-castellano/go-services/infra/logger"
	opentelemetry "github.com/a-castellano/go-services/infra/opentelemetry"
	envelope "github.com/a-castellano/go-types/types/envelope"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const tracerName = "github.com/a-castellano/home-ip-notifier/internal/infra/consume"

// Processor is what the consumer needs from the use case; app.Announcer
// satisfies it. The interface is declared here, where it is consumed (Go
// idiom), so the adapter can be tested against a fake.
type Processor interface {
	ProcessMessage(ctx context.Context, message string) error
}

// Consumer turns the raw deliveries of one queue into use-case calls.
type Consumer struct {
	processor Processor
	queue     string
}

// NewConsumer returns a Consumer for queue that delegates to processor.
func NewConsumer(queue string, processor Processor) Consumer {
	return Consumer{processor: processor, queue: queue}
}

// Consume handles one delivery. Malformed envelopes and empty bodies are
// logged and dropped (nil is returned): they carry no trace to join and, with
// auto-ack consumption, failing would not requeue them. For valid envelopes it
// opens a CONSUMER span as a child of the trace context carried in the
// envelope and returns whatever the use case returns.
func (c Consumer) Consume(ctx context.Context, receivedData []byte) error {

	log := logger.FromContext(ctx).With("operation", "consume")
	log.DebugContext(ctx, "unmarshaling envelope from received data")

	receivedEnvelope, unmarshalErr := envelope.Unmarshal(receivedData)
	if unmarshalErr != nil {
		log.ErrorContext(ctx, "cannot unmarshal data", "error", unmarshalErr.Error())
		return nil
	}

	if len(receivedEnvelope.Body) == 0 {
		errorString := "received body is empty"
		log.ErrorContext(ctx, errorString)
		return nil
	}

	ctx = opentelemetry.Extract(ctx, receivedEnvelope)
	ctx, span := otel.Tracer(tracerName).Start(ctx, "process "+c.queue,
		trace.WithSpanKind(trace.SpanKindConsumer),
		trace.WithAttributes(
			attribute.String("messaging.destination.name", c.queue),
			attribute.String("messaging.operation.type", "process"),
		))
	defer span.End()

	plainMessage := string(receivedEnvelope.Body)

	log.DebugContext(ctx, "process message", "message", plainMessage)

	processErr := c.processor.ProcessMessage(ctx, plainMessage)

	if processErr != nil {
		// Status only: the error event and the log are already
		// recorded closest to the point of error
		span.SetStatus(codes.Error, "message process has failed")
		return processErr
	}

	return nil
}

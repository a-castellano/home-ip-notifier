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

// same signature as app.Announcer.ProcessMessage.
type Processor interface {
	ProcessMessage(ctx context.Context, message string) error
}

type Consumer struct {
	processor Processor
	queue     string
}

func NewConsumer(queue string, processor Processor) Consumer {
	return Consumer{processor: processor, queue: queue}
}

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
		errorString := "message process has failed"
		log.ErrorContext(ctx, errorString, "error", processErr)
		// Status only: the error event is already recorded by the
		// child span
		span.SetStatus(codes.Error, errorString)
		return processErr
	}

	return nil
}

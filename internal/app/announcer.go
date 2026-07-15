package app

import (
	"context"

	logger "github.com/a-castellano/go-services/infra/logger"
	domain "github.com/a-castellano/home-ip-notifier/internal/domain"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const subject = "Home IP has changed"
const tracerName = "github.com/a-castellano/home-ip-notifier/internal/app"

type Announcer struct {
	notifier domain.Notifier
}

func NewAnnouncer(notifier domain.Notifier) Announcer {
	return Announcer{notifier: notifier}
}

func (a Announcer) ProcessMessage(ctx context.Context, message string) error {

	ctx, span := otel.Tracer(tracerName).Start(ctx, "ProcessMessage",
		trace.WithSpanKind(trace.SpanKindInternal))
	defer span.End()

	log := logger.FromContext(ctx).With("operation", "ProcessMessage")
	log.DebugContext(ctx, "Notifying message")
	err := a.notifier.Notify(ctx, subject, message)
	if err != nil {
		errorString := "error during Notify call"
		log.ErrorContext(ctx, errorString, "error", err)
		// Status only: the error event is already recorded by the
		// child span
		span.SetStatus(codes.Error, errorString)
		return a.notifier.Notify(ctx, subject, message)
	}
	return nil
}

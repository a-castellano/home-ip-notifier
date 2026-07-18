// Package app holds the single use case of the service: turning a received
// IP-change message into a notification. It depends only on the domain ports,
// never on infrastructure.
package app

import (
	"context"

	logger "github.com/a-castellano/go-services/infra/logger"
	domain "github.com/a-castellano/home-ip-notifier/internal/domain"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// subject titles every notification: the service has a single use case, so
// the subject is a business constant here rather than configuration.
const subject = "Home IP has changed"

const tracerName = "github.com/a-castellano/home-ip-notifier/internal/app"

// Announcer is the application use case. Its only dependency is the
// domain.Notifier port, so it has zero knowledge of SMTP or RabbitMQ.
type Announcer struct {
	notifier domain.Notifier
}

// NewAnnouncer returns an Announcer that delivers through notifier.
func NewAnnouncer(notifier domain.Notifier) Announcer {
	return Announcer{notifier: notifier}
}

// ProcessMessage sends the received IP-change message through the Notifier
// port under the fixed subject. The error is returned undecorated: what to do
// about a failed delivery is a policy decided at the composition root.
func (a Announcer) ProcessMessage(ctx context.Context, message string) error {

	ctx, span := otel.Tracer(tracerName).Start(ctx, "ProcessMessage",
		trace.WithSpanKind(trace.SpanKindInternal))
	defer span.End()

	log := logger.FromContext(ctx).With("operation", "ProcessMessage")
	log.DebugContext(ctx, "Notifying message")
	err := a.notifier.Notify(ctx, subject, message)
	if err != nil {
		// Status only: the error event and the log are already
		// recorded closest to the point of error
		span.SetStatus(codes.Error, "error during Notify call")
		return err
	}
	return nil
}

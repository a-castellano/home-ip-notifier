// Package notify is the outbound adapter of the service: it implements the
// domain.Notifier port on top of the go-services notificator backed by an
// SMTP MailSender, so severity levels and instrumentation come from the
// library instead of being reimplemented here.
package notify

import (
	"context"

	logger "github.com/a-castellano/go-services/infra/logger"
	"github.com/a-castellano/go-services/infra/smtp"
	"github.com/a-castellano/go-services/services/notificator"
	notification "github.com/a-castellano/go-types/types/notification"
	smtpconfig "github.com/a-castellano/go-types/types/smtp"
)

// MailNotificator implements domain.Notifier by email. The destination does
// not travel through the port: it is deployment configuration, captured here
// at construction time.
type MailNotificator struct {
	driver      *notificator.Notificator
	destination string
}

// NewNotificator wires the SMTP delivery chain (TLS dialer → sender →
// MailSender → notificator) and binds it to the configured destination.
func NewNotificator(ctx context.Context, config *smtpconfig.Config, destination string) *MailNotificator {
	log := logger.FromContext(ctx).With("operation", "NewNotificator")

	log.DebugContext(ctx, "creating notificator")

	sender := smtp.NewSender(config, smtp.TLSDialer{})
	mailSender := smtp.NewMailSender(sender)

	mailDriver := notificator.NewNotificator(mailSender)

	return &MailNotificator{driver: mailDriver, destination: destination}
}

// Notify builds a Notification for the configured destination (severity is
// the NewNotification default, Info) and delegates delivery to the
// notificator. Empty title or message is rejected by the notification
// constructor itself.
func (n *MailNotificator) Notify(ctx context.Context, title string, message string) error {
	log := logger.FromContext(ctx).With("operation", "Notify")

	log.DebugContext(ctx, "prepare new notification")

	newMessage, newMessageError := notification.NewNotification(
		n.destination,
		title,
		message,
	)

	if newMessageError != nil {
		errorString := "failed to create a new notification"
		log.ErrorContext(ctx, errorString, "error", newMessageError)
		return newMessageError
	}
	// Send failures are already logged and recorded by the driver at the
	// point of error; just propagate.
	return n.driver.Notify(ctx, newMessage)
}

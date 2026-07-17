package notify

import (
	"context"

	logger "github.com/a-castellano/go-services/infra/logger"
	"github.com/a-castellano/go-services/infra/smtp"
	"github.com/a-castellano/go-services/services/notificator"
	notification "github.com/a-castellano/go-types/types/notification"
	smtpconfig "github.com/a-castellano/go-types/types/smtp"
)

type MailNotificator struct {
	driver      *notificator.Notificator
	destination string
}

func NewNotificator(ctx context.Context, config *smtpconfig.Config, destination string) *MailNotificator {
	log := logger.FromContext(ctx).With("operation", "NewNotificator")

	log.DebugContext(ctx, "creating notificator")

	sender := smtp.NewSender(config, smtp.TLSDialer{})
	mailSender := smtp.NewMailSender(sender)

	mailDriver := notificator.NewNotificator(mailSender)

	return &MailNotificator{driver: mailDriver, destination: destination}
}

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

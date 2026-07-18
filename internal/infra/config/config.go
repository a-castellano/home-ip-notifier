// Package config loads the service configuration from environment variables,
// composing the SMTP and RabbitMQ configs already provided by go-types with
// the two values that belong to this service: the destination address and the
// queue to consume from.
package config

import (
	"cmp"
	"context"
	"errors"
	"net/mail"
	"os"

	logger "github.com/a-castellano/go-services/infra/logger"
	rabbitmqconfig "github.com/a-castellano/go-types/types/rabbitmq"
	smtpconfig "github.com/a-castellano/go-types/types/smtp"
)

// ErrMissingDestination is returned when the DESTINATION env variable is not set.
var ErrMissingDestination = errors.New("DESTINATION env variable must be set")

// Config holds everything the composition root needs to wire the service:
// where to consume from (RabbitMQ + queue) and where to deliver to
// (SMTP + destination address).
type Config struct {
	SMTPConfig  *smtpconfig.Config
	Destination string

	NotifyQueue    string
	RabbitmqConfig *rabbitmqconfig.Config
}

// NewConfig reads and validates the environment: DESTINATION must be a valid
// e-mail address, the SMTP_* and RABBITMQ_* groups are validated by their
// go-types constructors, and NOTIFY_QUEUE_NAME falls back to
// "home-ip-monitor-notifications", the monitor's default notification queue.
func NewConfig(ctx context.Context) (*Config, error) {
	config := Config{}

	var rabbitmqConfigError, smtpConfigError error

	log := logger.FromContext(ctx).With("operation", "NewConfig")

	destination, destinationVariableFound := os.LookupEnv("DESTINATION")

	if !destinationVariableFound {
		log.ErrorContext(ctx, "error retrieving destination", "error", ErrMissingDestination)
		return nil, ErrMissingDestination
	}

	_, mailErr := mail.ParseAddress(destination)
	if mailErr != nil {
		log.ErrorContext(ctx, "cannot validate destination, it should be a valid e-mail", "error", mailErr)
		return nil, mailErr
	}

	config.Destination = destination
	log.DebugContext(ctx, "destination set", "destination", config.Destination)

	config.NotifyQueue = cmp.Or(os.Getenv("NOTIFY_QUEUE_NAME"), "home-ip-monitor-notifications")
	log.DebugContext(ctx, "queue where messages will be processed has been set", "queue", config.NotifyQueue)

	config.RabbitmqConfig, rabbitmqConfigError = rabbitmqconfig.NewConfig()
	if rabbitmqConfigError != nil {
		log.ErrorContext(ctx, "error setting rabbitmq config", "error", rabbitmqConfigError)
		return nil, rabbitmqConfigError
	}

	log.DebugContext(ctx, "rabbitmq config set", "rabbitmq config", config.RabbitmqConfig)

	config.SMTPConfig, smtpConfigError = smtpconfig.NewConfig()
	if smtpConfigError != nil {
		log.ErrorContext(ctx, "error setting smtp config", "error", smtpConfigError)
		return nil, smtpConfigError
	}

	log.DebugContext(ctx, "smtp config set", "smtp config", config.SMTPConfig)

	return &config, nil
}

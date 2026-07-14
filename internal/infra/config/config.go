package config

import (
	"cmp"
	"context"
	"errors"
	"os"

	logger "github.com/a-castellano/go-services/infra/logger"
	rabbitmqconfig "github.com/a-castellano/go-types/types/rabbitmq"
	smtpconfig "github.com/a-castellano/go-types/types/smtp"
)

// ErrMissingDestination is returned when the DESTINATION env variable is not set.
var ErrMissingDestination = errors.New("DESTINATION env variable must be set")

type Config struct {
	SMTPConfig  *smtpconfig.Config
	Destination string

	NotifyQueue    string
	RabbitmqConfig *rabbitmqconfig.Config
}

func NewConfig(ctx context.Context) (*Config, error) {
	config := Config{}

	var rabbitmqConfigError, smtpConfigError error

	log := logger.FromContext(ctx).With("operation", "NewConfig")

	destination, destinationVariableFound := os.LookupEnv("DESTINATION")

	if !destinationVariableFound {
		log.ErrorContext(ctx, "error retrieving destination", "error", ErrMissingDestination)
		return nil, ErrMissingDestination
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

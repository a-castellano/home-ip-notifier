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

	if destinationVariableFound == false {
		errorString := "DESTINATION env variable must be set"
		log.ErrorContext(ctx, "error retriveing destiantion", "error", errorString)
		return nil, errors.New(errorString)
	}
	config.Destination = destination
	log.DebugContext(ctx, "destiantion set", "destination", config.Destination)

	config.NotifyQueue = cmp.Or(os.Getenv("NOTIFY_QUEUE_NAME"), "home-ip-monitor-notifications")
	log.DebugContext(ctx, "queue where messages will be process has ben set", "queue", config.NotifyQueue)

	config.RabbitmqConfig, rabbitmqConfigError = rabbitmqconfig.NewConfig()
	if rabbitmqConfigError != nil {
		log.ErrorContext(ctx, "error setting rabbimq config", "error", rabbitmqConfigError)
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

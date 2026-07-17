package main

import (
	"context"
	systemlog "log"
	"os"
	"os/signal"
	"syscall"

	logger "github.com/a-castellano/go-services/infra/logger"
	opentelemetry "github.com/a-castellano/go-services/infra/opentelemetry"
	rabbitmq "github.com/a-castellano/go-services/infra/rabbitmq"
	messagebroker "github.com/a-castellano/go-services/services/messagebroker"
	otelconfig "github.com/a-castellano/go-types/types/opentelemetry"
	slogconfig "github.com/a-castellano/go-types/types/slog"
	announce "github.com/a-castellano/home-ip-notifier/internal/app"
	config "github.com/a-castellano/home-ip-notifier/internal/infra/config"
	consume "github.com/a-castellano/home-ip-notifier/internal/infra/consume"
	notify "github.com/a-castellano/home-ip-notifier/internal/infra/notify"
)

func run(ctx context.Context, cancel context.CancelFunc) error {

	log := logger.FromContext(ctx).With("operation", "main.run")
	log.DebugContext(ctx, "Loading config")

	otelConfig, otelConfigErr := otelconfig.NewConfig()
	if otelConfigErr != nil {
		log.ErrorContext(ctx, "telemetry config has errors", "error", otelConfigErr)
		return otelConfigErr
	}

	shutdown, err := opentelemetry.SetupOpenTelemetry(ctx, otelConfig)
	if err != nil {
		// Telemetry failed to start; the app keeps running without it.
		log.ErrorContext(ctx, "telemetry setup failed", "error", err)
	}
	defer func() {
		if err := shutdown(ctx); err != nil {
			log.ErrorContext(ctx, "telemetry shutdown failed", "error", err)
		}
	}()

	appConfig, configErr := config.NewConfig(ctx)

	if configErr != nil {
		log.ErrorContext(ctx, "Error loading app config", "error", configErr)
		return configErr
	}

	log.InfoContext(ctx, "Initiating required services")
	log.DebugContext(ctx, "Defining rabbitmq instance")
	rabbitmqClient := rabbitmq.NewRabbitmqClient(appConfig.RabbitmqConfig)
	log.DebugContext(ctx, "Defining messagebroker instance")
	messageBroker := messagebroker.MessageBroker{Client: rabbitmqClient}

	messagesReceived := make(chan []byte)
	receiveErrors := make(chan error)

	// Set up signal handling for graceful shutdown (SIGINT, SIGTERM)
	signalChannel := make(chan os.Signal, 2)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

	// Start signal handler goroutine
	go func() {
		sig := <-signalChannel
		switch sig {
		case os.Interrupt, syscall.SIGTERM:
			cancel()
		}
	}()

	log.DebugContext(ctx, "creating notificator")
	notificator := notify.NewNotificator(ctx, appConfig.SMTPConfig, appConfig.Destination)
	log.DebugContext(ctx, "creating announer")
	announcer := announce.NewAnnouncer(notificator)
	log.DebugContext(ctx, "creating consumer")
	consumer := consume.NewConsumer(appConfig.NotifyQueue, announcer)

	go messageBroker.ReceiveMessages(ctx, appConfig.NotifyQueue, messagesReceived, receiveErrors)

	log.InfoContext(ctx, "waiting for messages")

	// Main message processing loop
	for {
		select {
		case receivedError := <-receiveErrors:
			// Handle RabbitMQ connection or message receiving errors
			log.ErrorContext(ctx, receivedError.Error())
			return receivedError
		case messageReceived := <-messagesReceived:
			log.InfoContext(ctx, "processing new message")
			consumer.Consume(ctx, messageReceived)

		case <-ctx.Done():
			// Graceful shutdown when context is cancelled
			log.InfoContext(ctx, "execution finished")
			return nil
		}
	}

}

func main() {

	// First, initiate logger
	logConfig, err := slogconfig.NewConfig()
	if err != nil {
		systemlog.Fatal(err)
	}

	appLogger := logger.NewLogger(logConfig)
	appContext, cancel := context.WithCancel(context.Background())
	ctx := logger.WithLogger(appContext, appLogger)

	runErr := run(ctx, cancel)
	if runErr != nil {
		appLogger.ErrorContext(ctx, "home-ip-notifier failed", "error", runErr)
		os.Exit(1)
	}
}

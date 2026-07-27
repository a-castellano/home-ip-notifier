// home-ip-notifier consumes IP-change notifications published by
// home-ip-monitor on RabbitMQ and delivers them by e-mail. This package is
// the composition root: it wires config, adapters and the use case together
// and owns the process lifecycle.
package main

import (
	"context"
	systemlog "log"
	"os"
	"os/signal"
	"syscall"
	"time"

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

// run owns the service lifecycle: telemetry setup, config, wiring and the
// message loop. It blocks until the signal context is cancelled (returns nil)
// or the broker reports a receive error (returned to main).
func run(ctx context.Context) error {

	// Graceful shutdown: SIGINT/SIGTERM cancel the context
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

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
		// By the time this runs the signal context is already cancelled;
		// give the exporters their own deadline to flush pending spans.
		shutdownCtx, cancelShutdown := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
		defer cancelShutdown()
		if err := shutdown(shutdownCtx); err != nil {
			log.ErrorContext(shutdownCtx, "telemetry shutdown failed", "error", err)
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

	log.DebugContext(ctx, "creating notificator")
	notificator := notify.NewNotificator(ctx, appConfig.SMTPConfig, appConfig.Destination)
	log.DebugContext(ctx, "creating announcer")
	announcer := announce.NewAnnouncer(notificator)
	log.DebugContext(ctx, "creating consumer")
	consumer := consume.NewConsumer(ctx, appConfig.NotifyQueue, announcer)

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
			// Failed messages are dropped on purpose: consumption is
			// auto-ack, so exiting would not requeue them, and the
			// failure is already logged and recorded in the trace.
			_ = consumer.Consume(ctx, messageReceived)

		case <-ctx.Done():
			// Graceful shutdown when context is cancelled
			log.InfoContext(ctx, "execution finished")
			return nil
		}
	}

}

// main only builds the logger and the root context and decides the exit
// code; everything else happens inside run so its deferred cleanups execute
// before the process exits (os.Exit here would skip defers placed in main).
func main() {

	// First, initiate logger
	logConfig, err := slogconfig.NewConfig()
	if err != nil {
		systemlog.Fatal(err)
	}

	appLogger := logger.NewLogger(logConfig, opentelemetry.NewSlogHandler(logConfig.AppName))
	ctx := logger.WithLogger(context.Background(), appLogger)

	runErr := run(ctx)
	if runErr != nil {
		appLogger.ErrorContext(ctx, "home-ip-notifier failed", "error", runErr)
		os.Exit(1)
	}
}

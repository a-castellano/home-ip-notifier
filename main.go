package main

import (
	"context"
	"log"
	"log/syslog"
	"os"
	"os/signal"
	"syscall"

	messagebroker "github.com/a-castellano/go-services/messagebroker"
	config "github.com/a-castellano/home-ip-notifier/config"
	mailutils "github.com/a-castellano/home-ip-notifier/mail"
)

// main is the entry point of the application.
// It sets up logging, configuration, RabbitMQ connection, and starts the message processing loop.
func main() {

	// Configure logger to write to the syslog for better system integration
	logwriter, e := syslog.New(syslog.LOG_INFO, "home-ip-notifier")
	if e == nil {
		log.SetOutput(logwriter)
		// Remove timestamp since syslog already provides it
		log.SetFlags(0)
	}

	log.Print("Loading config")

	// Initialize application configuration from environment variables
	appConfig, configError := config.NewConfig()

	if configError != nil {
		log.Print(configError.Error())
		os.Exit(1)
	}

	log.Print("Creating RabbitMQ client")

	// Create a cancellable context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())

	// Initialize RabbitMQ client and message broker
	rabbitmqClient := messagebroker.NewRabbitmqClient(appConfig.RabbitmqConfig)
	messageBroker := messagebroker.MessageBroker{Client: rabbitmqClient}

	// Create channels for message processing and error handling
	messagesReceived := make(chan []byte)
	receiveErrors := make(chan error)

	log.Print("Define os signal management")

	// Set up signal handling for graceful shutdown (SIGINT, SIGTERM)
	signalChannel := make(chan os.Signal, 2)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

	// Start signal handler goroutine
	go func() {
		sig := <-signalChannel
		switch sig {
		case os.Interrupt:
			cancel()
		case syscall.SIGTERM:
			cancel()
		}
	}()

	// Start message receiver in a separate goroutine
	go messageBroker.ReceiveMessages(ctx, appConfig.NotifyQueue, messagesReceived, receiveErrors)

	log.Print("Waiting for messages")

	// Main message processing loop
	for {
		select {
		case receivedError := <-receiveErrors:
			// Handle RabbitMQ connection or message receiving errors
			log.Print(receivedError.Error())
			os.Exit(1)
		case messageReceived := <-messagesReceived:
			// Process received IP change notification
			messageToSend := string(messageReceived)
			log.Printf("Received new message: %s", messageToSend)
			log.Print("Sending Email")

			// Send email notification about IP change
			sendError := mailutils.SendEmail(appConfig, messageToSend)

			if sendError != nil {
				log.Print(sendError.Error())
				os.Exit(1)
			}

		case <-ctx.Done():
			// Graceful shutdown when context is cancelled
			log.Print("Execution finished")
			os.Exit(0)
		}
	}

}

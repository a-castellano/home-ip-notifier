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

func main() {

	// Configure logger to write to the syslog.
	logwriter, e := syslog.New(syslog.LOG_INFO, "home-ip-notifier")
	if e == nil {
		log.SetOutput(logwriter)
		// Remove timestamp
		log.SetFlags(0)
	}

	// Now from anywhere else in your program, you can use this:
	log.Print("Loading config")

	appConfig, configErr := config.NewConfig()

	if configErr != nil {
		log.Print(configErr.Error())
		os.Exit(1)
	}

	log.Print("Creating RabbitMQ client")
	ctx, cancel := context.WithCancel(context.Background())

	rabbitmqClient := messagebroker.NewRabbitmqClient(appConfig.RabbitmqConfig)
	messageBroker := messagebroker.MessageBroker{Client: rabbitmqClient}

	messagesReceived := make(chan []byte)
	receiveErrors := make(chan error)

	log.Print("Define os signal management")
	signalChannel := make(chan os.Signal, 2)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-signalChannel
		switch sig {
		case os.Interrupt:
			cancel()
		case syscall.SIGTERM:
			cancel()
		}
	}()

	go messageBroker.ReceiveMessages(ctx, appConfig.NotifyQueue, messagesReceived, receiveErrors)

	log.Print("Waiting for messages")

	for {
		select {
		case receivedError := <-receiveErrors:
			log.Print(receivedError.Error())
			os.Exit(1)
		case messageReceived := <-messagesReceived:
			messageToSend := string(messageReceived)
			log.Printf("Received new message: %s", messageToSend)
			log.Print("Sending Email")
			sendErr := mailutils.SendEmail(appConfig, messageToSend)

			if sendErr != nil {
				log.Print(sendErr.Error())
				os.Exit(1)
			}

		case <-ctx.Done():
			log.Print("Execution finished")
			os.Exit(0)
		}
	}

}

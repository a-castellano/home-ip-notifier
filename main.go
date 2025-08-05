package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"log/syslog"
	"net"
	"net/mail"
	"net/smtp"
	"os"
	"os/signal"
	"syscall"

	messagebroker "github.com/a-castellano/go-services/messagebroker"
	config "github.com/a-castellano/home-ip-notifier/config"
)

func sendEmail(config *config.Config, messageToSend string) error {

	fromMail := fmt.Sprintf("%s@%s", config.MailFrom, config.MailDomain)
	from := mail.Address{Name: "", Address: fromMail}
	to := mail.Address{Name: "", Address: config.Destination}
	subj := "Home IP has changed"

	// Setup headers
	headers := make(map[string]string)
	headers["From"] = from.String()
	headers["To"] = to.String()
	headers["Subject"] = subj

	// Setup message
	var message string
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + messageToSend

	// Connect to the SMTP Server
	servername := fmt.Sprintf("%s:%d", config.SMTPHost, config.SMTPPort)

	host, _, _ := net.SplitHostPort(servername)

	auth := smtp.PlainAuth("", config.SMTPName, config.SMTPPassword, host)

	// TLS config
	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         host,
	}

	// Here is the key, you need to call tls.Dial instead of smtp.Dial
	// for smtp servers running on 465 that require an ssl connection
	// from the very beginning (no starttls)
	conn, err := tls.Dial("tcp", servername, tlsconfig)
	if err != nil {
		return err
	}

	c, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}

	// Auth
	if err = c.Auth(auth); err != nil {
		return err
	}

	// To && From
	if err = c.Mail(from.Address); err != nil {
		return err
	}

	if err = c.Rcpt(to.Address); err != nil {
		return err
	}

	// Data
	w, err := c.Data()
	if err != nil {
		return err
	}

	_, err = w.Write([]byte(message))
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	c.Quit()
	log.Println("Mail sent successfully")

	return nil
}

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
			sendErr := sendEmail(appConfig, messageToSend)

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

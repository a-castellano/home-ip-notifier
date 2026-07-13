package config

import (
	"cmp"
	"errors"
	"fmt"
	"os"
	"strconv"

	rabbitmqconfig "github.com/a-castellano/go-types/rabbitmq"
)

// Config struct contains required config variables for the home-ip-notifier application.
// It includes both SMTP email configuration and RabbitMQ connection settings.
type Config struct {
	// SMTP Email Configuration
	MailFrom          string // Sender email username (without domain)
	MailDomain        string // Sender email domain
	SMTPHost          string // SMTP server hostname
	SMTPPort          int    // SMTP server port
	SMTPName          string // SMTP authentication username
	SMTPPassword      string // SMTP authentication password
	SMTPTLSValidation bool   // Whether to validate TLS certificates
	Destination       string // Recipient email address

	// RabbitMQ Configuration
	NotifyQueue    string                 // Queue name for IP change notifications
	RabbitmqConfig *rabbitmqconfig.Config // RabbitMQ connection settings
}

// NewConfig checks if required environment variables are present and returns a config instance.
// It validates all required SMTP and RabbitMQ configuration variables.
// Returns an error if any required variable is missing or invalid.
func NewConfig() (*Config, error) {
	config := Config{}

	// Check if all required environment variables are defined
	requiredEnvVariables := []string{"MAILFROM", "MAILDOMAIN", "SMTPHOST", "SMTPPORT", "SMTPNAME", "SMTPPASSWORD", "DESTINATION"}

	for _, requiredEnvVariable := range requiredEnvVariables {
		if _, envVariableFound := os.LookupEnv(requiredEnvVariable); !envVariableFound {
			errorString := fmt.Sprintf("%s env variable must be set", requiredEnvVariable)
			return nil, errors.New(errorString)
		}
	}

	// Parse SMTP port from string to integer
	var portAtoiError error
	config.SMTPPort, portAtoiError = strconv.Atoi(os.Getenv("SMTPPORT"))

	if portAtoiError != nil {
		return nil, errors.New("Failed to parse SMTPPORT value")
	}

	// Load SMTP configuration from environment variables
	config.MailFrom = os.Getenv("MAILFROM")
	config.MailDomain = os.Getenv("MAILDOMAIN")
	config.SMTPHost = os.Getenv("SMTPHOST")
	config.SMTPName = os.Getenv("SMTPNAME")
	config.SMTPPassword = os.Getenv("SMTPPASSWORD")
	config.Destination = os.Getenv("DESTINATION")

	// Check SMTP TLS validation setting (defaults to true if not specified)
	config.SMTPTLSValidation = cmp.Or(os.Getenv("SMTPTLSVALIDATION"), "true") == "true"

	// Initialize RabbitMQ configuration
	var rabbitmqConfigError error
	config.RabbitmqConfig, rabbitmqConfigError = rabbitmqconfig.NewConfig()
	if rabbitmqConfigError != nil {
		return nil, rabbitmqConfigError
	}

	// Set notification queue name (defaults to "home-ip-monitor-notifications" if not specified)
	config.NotifyQueue = cmp.Or(os.Getenv("NOTIFY_QUEUE_NAME"), "home-ip-monitor-notifications")

	return &config, nil
}

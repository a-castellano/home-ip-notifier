package config

import (
	"cmp"
	"errors"
	"fmt"
	"os"
	"strconv"

	rabbitmqconfig "github.com/a-castellano/go-types/rabbitmq"
)

// Config struct contians required config variables
type Config struct {
	MailFrom          string
	MailDomain        string
	SMTPHost          string
	SMTPPort          int
	SMTPName          string
	SMTPPassword      string
	SMTPTLSValidation bool
	Destination       string
	NotifyQueue       string
	RabbitmqConfig    *rabbitmqconfig.Config
}

// NewConfig checks if required env variables are present, returns config instance
func NewConfig() (*Config, error) {
	config := Config{}

	// check if all required env variables are defined
	requiredEnvVariables := []string{"MAILFROM", "MAILDOMAIN", "SMTPHOST", "SMTPPORT", "SMTPNAME", "SMTPPASSWORD", "DESTINATION"}

	for _, requiredEnvVariable := range requiredEnvVariables {
		if _, envVariableFound := os.LookupEnv(requiredEnvVariable); !envVariableFound {
			errorString := fmt.Sprintf("%s env variable must be set", requiredEnvVariable)
			return nil, errors.New(errorString)
		}
	}
	var portAtoiErr error
	config.SMTPPort, portAtoiErr = strconv.Atoi(os.Getenv("SMTPPORT"))

	if portAtoiErr != nil {
		return nil, errors.New("Failed to parse SMTPPORT value")
	}

	config.MailFrom = os.Getenv("MAILFROM")
	config.MailDomain = os.Getenv("MAILDOMAIN")
	config.SMTPHost = os.Getenv("SMTPHOST")
	config.SMTPName = os.Getenv("SMTPNAME")
	config.SMTPPassword = os.Getenv("SMTPPASSWORD")
	config.Destination = os.Getenv("DESTINATION")

	//Check SMTP tls validation
	config.SMTPTLSValidation = cmp.Or(os.Getenv("SMTPTLSVALIDATION"), "true") == "true"

	var rabbitmqConfigErr error
	config.RabbitmqConfig, rabbitmqConfigErr = rabbitmqconfig.NewConfig()
	if rabbitmqConfigErr != nil {
		return nil, rabbitmqConfigErr
	}

	config.NotifyQueue = cmp.Or(os.Getenv("NOTIFY_QUEUE_NAME"), "home-ip-monitor-notifications")

	return &config, nil
}

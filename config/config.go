package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	rabbitmqconfig "github.com/a-castellano/go-types/rabbitmq"
)

// Config struct contians required config variables
type Config struct {
	MailFrom       string
	MailDomain     string
	SMTPHost       string
	SMTPPort       int
	SMTPName       string
	SMTPPassword   string
	Destination    string
	RabbitmqConfig *rabbitmqconfig.Config
}

// NewConfig checks if required env variables are present, returns config instance
func NewConfig() (*Config, error) {
	config := Config{}

	// check if all required env variables are defined
	requiredEnvVariables := []string{"MailFrom", "MailDomain", "SMTPHost", "SMTPPort", "SMTPName", "SMTPPassword", "Destination"}

	for _, requiredEnvVariable := range requiredEnvVariables {
		if _, envVariableFound := os.LookupEnv(requiredEnvVariable); !envVariableFound {
			errorString := fmt.Sprintf("%s env variable must be set", requiredEnvVariable)
			return nil, errors.New(errorString)
		}
	}
	var portAtoiErr error
	config.SMTPPort, portAtoiErr = strconv.Atoi(os.Getenv("SMTPPort"))

	if portAtoiErr != nil {
		return nil, errors.New("Failed to parse SMTPPort value")
	}

	config.MailFrom = os.Getenv("MailFrom")
	config.MailDomain = os.Getenv("MailDomain")
	config.SMTPHost = os.Getenv("SMTPHost")
	config.SMTPName = os.Getenv("SMTPName")
	config.SMTPPassword = os.Getenv("SMTPPassword")
	config.Destination = os.Getenv("Destination")

	var rabbitmqConfigErr error
	config.RabbitmqConfig, rabbitmqConfigErr = rabbitmqconfig.NewConfig()
	if rabbitmqConfigErr != nil {
		return nil, rabbitmqConfigErr
	}

	return &config, nil
}

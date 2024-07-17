//go:build integration_tests || unit_tests

package config

import (
	"os"
	"testing"
)

var currentMailFrom string
var currentMailFromDefined bool

var currentMailDomain string
var currentMailDomainDefined bool

var currentSMTPPort string
var currentSMTPPortDefined bool

var currentSMTPHost string
var currentSMTPHostDefined bool

var currentSMTPName string
var currentSMTPNameDefined bool

var currentSMTPPassword string
var currentSMTPPasswordDefined bool

var currentDestination string
var currentDestinationDefined bool

var currentRabbitmqHost string
var currentRabbitmqHostDefined bool

var currentRabbitmqPort string
var currentRabbitmqPortDefined bool

var currentRabbitmqUser string
var currentRabbitmqUserDefined bool

var currentRabbitmqPassword string
var currentRabbitmqPasswordDefined bool

func setUp() {

	if envMailFrom, found := os.LookupEnv("MailFrom"); found {
		currentMailFrom = envMailFrom
		currentMailFromDefined = true
	} else {
		currentMailFromDefined = false
	}

	if envMailDomain, found := os.LookupEnv("MailDomain"); found {
		currentMailDomain = envMailDomain
		currentMailDomainDefined = true
	} else {
		currentMailDomainDefined = false
	}

	if envSMTPHost, found := os.LookupEnv("SMTPHost"); found {
		currentSMTPHost = envSMTPHost
		currentSMTPHostDefined = true
	} else {
		currentSMTPHostDefined = false
	}

	if envSMTPPort, found := os.LookupEnv("SMTPPort"); found {
		currentSMTPPort = envSMTPPort
		currentSMTPPortDefined = true
	} else {
		currentSMTPPortDefined = false
	}

	if envSMTPName, found := os.LookupEnv("SMTPName"); found {
		currentSMTPName = envSMTPName
		currentSMTPNameDefined = true
	} else {
		currentSMTPNameDefined = false
	}

	if envSMTPPassword, found := os.LookupEnv("SMTPPassword"); found {
		currentSMTPPassword = envSMTPPassword
		currentSMTPPasswordDefined = true
	} else {
		currentSMTPPasswordDefined = false
	}

	if envDestination, found := os.LookupEnv("Destination"); found {
		currentDestination = envDestination
		currentDestinationDefined = true
	} else {
		currentDestinationDefined = false
	}

	os.Unsetenv("MailFrom")
	os.Unsetenv("MailDomain")
	os.Unsetenv("SMTPHost")
	os.Unsetenv("SMTPPort")
	os.Unsetenv("SMTPName")
	os.Unsetenv("SMTPPassword")
	os.Unsetenv("Destination")

	os.Unsetenv("RABBITMQ_HOST")
	os.Unsetenv("RABBITMQ_PORT")
	os.Unsetenv("RABBITMQ_DATABASE")
	os.Unsetenv("RABBITMQ_PASSWORD")
}

func teardown() {

	if currentMailFromDefined {
		os.Setenv("MailFrom", currentMailFrom)
	} else {
		os.Unsetenv("MailFrom")
	}

	if currentMailDomainDefined {
		os.Setenv("MailDomain", currentMailDomain)
	} else {
		os.Unsetenv("MailDomain")
	}

	if currentSMTPHostDefined {
		os.Setenv("SMTPHost", currentSMTPHost)
	} else {
		os.Unsetenv("SMTPHost")
	}

	if currentSMTPPortDefined {
		os.Setenv("SMTPPort", currentSMTPPort)
	} else {
		os.Unsetenv("SMTPPort")
	}

	if currentSMTPNameDefined {
		os.Setenv("SMTPName", currentSMTPName)
	} else {
		os.Unsetenv("SMTPName")
	}

	if currentSMTPPasswordDefined {
		os.Setenv("SMTPPassword", currentSMTPPassword)
	} else {
		os.Unsetenv("SMTPPassword")
	}

	if currentDestinationDefined {
		os.Setenv("Destination", currentDestination)
	} else {
		os.Unsetenv("Destination")
	}

	if currentRabbitmqHostDefined {
		os.Setenv("RABBITMQ_HOST", currentRabbitmqHost)
	} else {
		os.Unsetenv("RABBITMQ_HOST")
	}

	if currentRabbitmqPortDefined {
		os.Setenv("RABBITMQ_PORT", currentRabbitmqPort)
	} else {
		os.Unsetenv("RABBITMQ_PORT")
	}

	if currentRabbitmqUserDefined {
		os.Setenv("RABBITMQ_USER", currentRabbitmqUser)
	} else {
		os.Unsetenv("RABBITMQ_USER")
	}

	if currentRabbitmqPasswordDefined {
		os.Setenv("RABBITMQ_PASSWORD", currentRabbitmqPassword)
	} else {
		os.Unsetenv("RABBITMQ_PASSWORD")
	}
}

func TestConfigWithoutEnvVariables(t *testing.T) {

	setUp()
	defer teardown()

	_, err := NewConfig()

	if err == nil {
		t.Errorf("TestConfigWithoutEnvVariables should fail.")
	} else {
		if err.Error() != "MailFrom env variable must be set" {
			t.Errorf("TestConfigWithoutEnvVariables error should be \"MailFrom env variable must be set\" but it was \"%s\".", err.Error())
		}
	}
}

func TestConfigWithInvalidSMTPPort(t *testing.T) {

	setUp()
	defer teardown()

	os.Setenv("MailFrom", "test")
	os.Setenv("MailDomain", "test")
	os.Setenv("SMTPHost", "test")
	os.Setenv("SMTPPort", "invalid")
	os.Setenv("SMTPName", "test")
	os.Setenv("SMTPPassword", "test")
	os.Setenv("Destination", "test")

	_, err := NewConfig()

	if err == nil {
		t.Errorf("TestConfigWithInvalidSMTPPort should fail.")
	} else {
		if err.Error() != "Failed to parse SMTPPort value" {
			t.Errorf("TestConfigWithInvalidSMTPPort error should be \"Failed to parse SMTPPort value\" but it was \"%s\".", err.Error())
		}
	}
}

func TestConfigWithInvalidRabbitMQPort(t *testing.T) {

	setUp()
	defer teardown()

	os.Setenv("MailFrom", "test")
	os.Setenv("MailDomain", "test")
	os.Setenv("SMTPHost", "test")
	os.Setenv("SMTPPort", "25")
	os.Setenv("SMTPName", "test")
	os.Setenv("SMTPPassword", "test")
	os.Setenv("Destination", "test")
	os.Setenv("RABBITMQ_PORT", "invalid")

	_, err := NewConfig()

	if err == nil {
		t.Errorf("TestConfigWithInvalidSMTPPort should fail with invalid RABBITMQ_PORT.")
	}
}

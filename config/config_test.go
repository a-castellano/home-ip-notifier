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

	if envMailFrom, found := os.LookupEnv("MAILFROM"); found {
		currentMailFrom = envMailFrom
		currentMailFromDefined = true
	} else {
		currentMailFromDefined = false
	}

	if envMailDomain, found := os.LookupEnv("MAILDOMAIN"); found {
		currentMailDomain = envMailDomain
		currentMailDomainDefined = true
	} else {
		currentMailDomainDefined = false
	}

	if envSMTPHost, found := os.LookupEnv("SMTPHOST"); found {
		currentSMTPHost = envSMTPHost
		currentSMTPHostDefined = true
	} else {
		currentSMTPHostDefined = false
	}

	if envSMTPPort, found := os.LookupEnv("SMTPPORT"); found {
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

	if envSMTPPassword, found := os.LookupEnv("SMTPPASSWORD"); found {
		currentSMTPPassword = envSMTPPassword
		currentSMTPPasswordDefined = true
	} else {
		currentSMTPPasswordDefined = false
	}

	if envDestination, found := os.LookupEnv("DESTINATION"); found {
		currentDestination = envDestination
		currentDestinationDefined = true
	} else {
		currentDestinationDefined = false
	}

	os.Unsetenv("MAILFROM")
	os.Unsetenv("MAILDOMAIN")
	os.Unsetenv("SMTPHOST")
	os.Unsetenv("SMTPPORT")
	os.Unsetenv("SMTPName")
	os.Unsetenv("SMTPPASSWORD")
	os.Unsetenv("DESTINATION")

	os.Unsetenv("RABBITMQ_HOST")
	os.Unsetenv("RABBITMQ_PORT")
	os.Unsetenv("RABBITMQ_DATABASE")
	os.Unsetenv("RABBITMQ_PASSWORD")
}

func teardown() {

	if currentMailFromDefined {
		os.Setenv("MAILFROM", currentMailFrom)
	} else {
		os.Unsetenv("MAILFROM")
	}

	if currentMailDomainDefined {
		os.Setenv("MAILDOMAIN", currentMailDomain)
	} else {
		os.Unsetenv("MAILDOMAIN")
	}

	if currentSMTPHostDefined {
		os.Setenv("SMTPHOST", currentSMTPHost)
	} else {
		os.Unsetenv("SMTPHOST")
	}

	if currentSMTPPortDefined {
		os.Setenv("SMTPPORT", currentSMTPPort)
	} else {
		os.Unsetenv("SMTPPORT")
	}

	if currentSMTPNameDefined {
		os.Setenv("SMTPName", currentSMTPName)
	} else {
		os.Unsetenv("SMTPName")
	}

	if currentSMTPPasswordDefined {
		os.Setenv("SMTPPASSWORD", currentSMTPPassword)
	} else {
		os.Unsetenv("SMTPPASSWORD")
	}

	if currentDestinationDefined {
		os.Setenv("DESTINATION", currentDestination)
	} else {
		os.Unsetenv("DESTINATION")
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
		if err.Error() != "MAILFROM env variable must be set" {
			t.Errorf("TestConfigWithoutEnvVariables error should be \"MAILFROM env variable must be set\" but it was \"%s\".", err.Error())
		}
	}
}

func TestConfigWithInvalidSMTPPort(t *testing.T) {

	setUp()
	defer teardown()

	os.Setenv("MAILFROM", "test")
	os.Setenv("MAILDOMAIN", "test")
	os.Setenv("SMTPHOST", "test")
	os.Setenv("SMTPPORT", "invalid")
	os.Setenv("SMTPNAME", "test")
	os.Setenv("SMTPPASSWORD", "test")
	os.Setenv("DESTINATION", "test")

	_, err := NewConfig()

	if err == nil {
		t.Errorf("TestConfigWithInvalidSMTPPORT should fail.")
	} else {
		if err.Error() != "Failed to parse SMTPPORT value" {
			t.Errorf("TestConfigWithInvalidSMTPPort error should be \"Failed to parse SMTPPORT value\" but it was \"%s\".", err.Error())
		}
	}
}

func TestConfigWithInvalidRabbitMQPort(t *testing.T) {

	setUp()
	defer teardown()

	os.Setenv("MAILFROM", "test")
	os.Setenv("MAILDOMAIN", "test")
	os.Setenv("SMTPHOST", "test")
	os.Setenv("SMTPPORT", "25")
	os.Setenv("SMTPName", "test")
	os.Setenv("SMTPPASSWORD", "test")
	os.Setenv("DESTINATION", "test")
	os.Setenv("RABBITMQ_PORT", "invalid")

	_, err := NewConfig()

	if err == nil {
		t.Errorf("TestConfigWithInvalidSMTPPort should fail with invalid RABBITMQ_PORT.")
	}
}

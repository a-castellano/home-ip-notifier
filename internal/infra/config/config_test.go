//go:build integration_tests || unit_tests

package config

import (
	"context"
	"errors"
	"os"
	"testing"
)

type envVariable struct {
	Value        string
	IsDefined    bool
	VariableName string
}

var envVariables = map[string]envVariable{
	//smtp
	"smtp_from":        {VariableName: "SMTP_FROM"},
	"smtp_host":        {VariableName: "SMTP_HOST"},
	"smtp_port":        {VariableName: "SMTP_PORT"},
	"smtp_username":    {VariableName: "SMTP_USERNAME"},
	"smtp_password":    {VariableName: "SMTP_PASSWORD"},
	"smtp_validateTLS": {VariableName: "SMTP_VALIDATE_TLS"},
	//rabbitmq
	"rabbitmq_host":     {VariableName: "RABBITMQ_HOST"},
	"rabbitmq_port":     {VariableName: "RABBITMQ_PORT"},
	"rabbitmq_user":     {VariableName: "RABBITMQ_USER"},
	"rabbitmq_password": {VariableName: "RABBITMQ_PASSWORD"},
	//home-ip-notifier
	"destination":       {VariableName: "DESTINATION"},
	"notify_queue_name": {VariableName: "NOTIFY_QUEUE_NAME"},
}

func setUp() {

	for key, variable := range envVariables {

		if envValue, found := os.LookupEnv(variable.VariableName); found {
			variable.Value = envValue
			variable.IsDefined = true
		} else {
			variable.IsDefined = false
		}

		os.Unsetenv(variable.VariableName)

		envVariables[key] = variable
	}

}

func teardown() {

	for _, variable := range envVariables {
		if variable.IsDefined {
			os.Setenv(variable.VariableName, variable.Value)
		} else {
			os.Unsetenv(variable.VariableName)
		}
	}
}

func TestConfigWithoutEnvVariables(t *testing.T) {

	setUp()
	defer teardown()

	ctx := context.Background()
	_, err := NewConfig(ctx)

	if err == nil {
		t.Fatalf("TestConfigWithoutEnvVariables should fail.")
	}
	if !errors.Is(err, ErrMissingDestination) {
		t.Fatalf("TestConfigWithoutEnvVariables error should be \"%s\" but it was \"%s\".", ErrMissingDestination, err.Error())
	}

}

func TestConfigWithoutSmtpConfig(t *testing.T) {

	setUp()
	defer teardown()

	os.Setenv("DESTINATION", "test@windmaker.net")

	ctx := context.Background()
	_, err := NewConfig(ctx)

	if err == nil {
		t.Fatalf("TestConfigWithoutSmtpConfig should fail.")
	}
	expectedError := "env variable \"SMTP_FROM\" must be set, cannot load smtp config"

	if err.Error() != expectedError {
		t.Fatalf("TestConfigWithoutSmtpConfig error should be \"%s\" but it was \"%s\".", expectedError, err.Error())
	}
}

func TestConfigWithoutRabbitmqConfig(t *testing.T) {

	setUp()
	defer teardown()

	os.Setenv("SMTP_FROM", "test@example.com")
	os.Setenv("SMTP_HOST", "test")
	os.Setenv("SMTP_PORT", "25")
	os.Setenv("SMTP_USERNAME", "test")
	os.Setenv("SMTP_PASSWORD", "test")
	os.Setenv("SMTP_VALIDATE_TLS", "anyvaluedifferentfromtrue")

	os.Setenv("DESTINATION", "test@windmaker.net")

	ctx := context.Background()
	_, err := NewConfig(ctx)

	if err != nil {
		t.Fatalf("TestConfigWithoutRabbitmqConfig should not fail as rabbitmq type has defult values")
	}
}

func TestConfigWithInvalidRabbitmqConfig(t *testing.T) {

	setUp()
	defer teardown()

	os.Setenv("SMTP_FROM", "test@example.com")
	os.Setenv("SMTP_HOST", "test")
	os.Setenv("SMTP_PORT", "25")
	os.Setenv("SMTP_USERNAME", "test")
	os.Setenv("SMTP_PASSWORD", "test")
	os.Setenv("SMTP_VALIDATE_TLS", "anyvaluedifferentfromtrue")

	os.Setenv("DESTINATION", "test@windmaker.net")

	os.Setenv("RABBITMQ_PORT", "invalid")

	ctx := context.Background()
	_, err := NewConfig(ctx)

	if err == nil {
		t.Fatalf("TestConfigWithInvalidRabbitmqConfig invalid")
	}
	expectedError := "strconv.Atoi: parsing \"invalid\": invalid syntax"

	if err.Error() != expectedError {
		t.Fatalf("TestConfigWithInvalidRabbitmqConfig error should be \"%s\" but it was \"%s\".", expectedError, err.Error())
	}
}

func TestConfigInvalidDestination(t *testing.T) {

	setUp()
	defer teardown()

	os.Setenv("SMTP_FROM", "test@example.com")
	os.Setenv("SMTP_HOST", "test")
	os.Setenv("SMTP_PORT", "25")
	os.Setenv("SMTP_USERNAME", "test")
	os.Setenv("SMTP_PASSWORD", "test")
	os.Setenv("SMTP_VALIDATE_TLS", "anyvaluedifferentfromtrue")

	os.Setenv("DESTINATION", "invalid")
	os.Setenv("NOTIFY_QUEUE_NAME", "other_queue")

	ctx := context.Background()
	_, err := NewConfig(ctx)

	if err == nil {
		t.Fatalf("TestConfigInvalidDestination should fail.")
	}
	expectedError := "mail: missing '@' or angle-addr"

	if err.Error() != expectedError {
		t.Fatalf("TestConfigInvalidDestination error should be \"%s\" but it was \"%s\".", expectedError, err.Error())
	}

}

func TestConfig(t *testing.T) {

	setUp()
	defer teardown()

	os.Setenv("SMTP_FROM", "test@example.com")
	os.Setenv("SMTP_HOST", "test")
	os.Setenv("SMTP_PORT", "25")
	os.Setenv("SMTP_USERNAME", "test")
	os.Setenv("SMTP_PASSWORD", "test")
	os.Setenv("SMTP_VALIDATE_TLS", "anyvaluedifferentfromtrue")

	os.Setenv("DESTINATION", "test@windmaker.net")
	os.Setenv("NOTIFY_QUEUE_NAME", "other_queue")

	ctx := context.Background()
	config, err := NewConfig(ctx)

	if err != nil {
		t.Fatalf("TestConfig should not fail")
	}

	expectedQueue := "other_queue"
	if config.NotifyQueue != expectedQueue {
		t.Fatalf("TestConfig config.NotifyQueue should be \"%s\" but it was \"%s\"", expectedQueue, config.NotifyQueue)
	}

	expectedSMTPUser := "test"
	if config.SMTPConfig.Username() != expectedSMTPUser {
		t.Fatalf("TestConfig config.SMTPConfig.User should be \"%s\" but it was \"%s\"", expectedSMTPUser, config.SMTPConfig.Username())
	}

	expectedDestination := "test@windmaker.net"
	if config.Destination != expectedDestination {
		t.Fatalf("TestConfig config.Destination should be \"%s\" but it was \"%s\"", expectedDestination, config.Destination)
	}

}

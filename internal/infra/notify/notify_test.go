//go:build integration_tests || unit_tests

package notify

import (
	"context"
	"os"
	"testing"

	smtpconfig "github.com/a-castellano/go-types/types/smtp"
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

func TestNewNotificator(t *testing.T) {

	setUp()
	defer teardown()

	os.Setenv("SMTP_FROM", "test@example.com")
	os.Setenv("SMTP_HOST", "test")
	os.Setenv("SMTP_PORT", "25")
	os.Setenv("SMTP_USERNAME", "test")
	os.Setenv("SMTP_PASSWORD", "test")
	os.Setenv("SMTP_VALIDATE_TLS", "anyvaluedifferentfromtrue")

	ctx := context.Background()
	config, err := smtpconfig.NewConfig()

	if err != nil {
		t.Fatalf("TestNewNotificator, smtp config should not fail")
	}

	destination := "to@someone.com"
	mailNotificator := NewNotificator(ctx, config, destination)

	if destination != mailNotificator.destination {
		t.Fatalf("TestNewNotificator destination should be \"%s\" but it was \"%s\".", destination, mailNotificator.destination)
	}

}

func TestNotifyEmptyDestination(t *testing.T) {
	setUp()
	defer teardown()

	os.Setenv("SMTP_FROM", "test@example.com")
	os.Setenv("SMTP_HOST", "test")
	os.Setenv("SMTP_PORT", "25")
	os.Setenv("SMTP_USERNAME", "test")
	os.Setenv("SMTP_PASSWORD", "test")
	os.Setenv("SMTP_VALIDATE_TLS", "anyvaluedifferentfromtrue")

	ctx := context.Background()
	config, err := smtpconfig.NewConfig()

	if err != nil {
		t.Fatalf("TestNewNotificator, smtp config should not fail")
	}

	destination := ""
	mailNotificator := NewNotificator(ctx, config, destination)

	notifyErr := mailNotificator.Notify(ctx, "any title", "message")

	if notifyErr == nil {
		t.Fatalf("TestNotifyEmptyDestination should fail")
	}

	expecterErr := "notification property cannot be empty: destination"
	if notifyErr.Error() != expecterErr {
		t.Fatalf("TestNotifyEmptyDestination error should be \"%s\" but it was \"%s\".", expecterErr, notifyErr.Error())

	}
}

func TestNotifyError(t *testing.T) {
	setUp()
	defer teardown()

	os.Setenv("SMTP_FROM", "test@example.com")
	os.Setenv("SMTP_HOST", "test")
	os.Setenv("SMTP_PORT", "25")
	os.Setenv("SMTP_USERNAME", "test")
	os.Setenv("SMTP_PASSWORD", "test")
	os.Setenv("SMTP_VALIDATE_TLS", "anyvaluedifferentfromtrue")

	ctx := context.Background()
	config, err := smtpconfig.NewConfig()

	if err != nil {
		t.Fatalf("TestNewNotificator, smtp config should not fail")
	}

	destination := "to@someone.com"
	mailNotificator := NewNotificator(ctx, config, destination)

	notifyErr := mailNotificator.Notify(ctx, "any title", "message")

	if notifyErr == nil {
		t.Fatalf("TestNotifyError should fail")
	}
}

//go:build integration_tests || unit_tests

package notify

import (
	"context"
	"os"
	"testing"

	smtpconfig "github.com/a-castellano/go-types/types/smtp"
)

func TestNotify(t *testing.T) {
	setUp()
	defer teardown()

	os.Setenv("SMTP_FROM", "test@example.com")
	os.Setenv("SMTP_HOST", "mailhog")
	os.Setenv("SMTP_PORT", "6465")
	os.Setenv("SMTP_USERNAME", "test")
	os.Setenv("SMTP_PASSWORD", "test")
	os.Setenv("SMTP_VALIDATE_TLS", "false")

	ctx := context.Background()
	config, err := smtpconfig.NewConfig()

	if err != nil {
		t.Fatalf("TestNewNotificator, smtp config should not fail")
	}

	destination := "to@someone.com"
	mailNotificator := NewNotificator(ctx, config, destination)

	notifyErr := mailNotificator.Notify(ctx, "any title", "message")

	if notifyErr != nil {
		t.Fatalf("TestNotify should not fail, error was \"%s\"", notifyErr.Error())
	}
}

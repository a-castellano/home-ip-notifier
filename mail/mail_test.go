//go:build integration_tests

package mail

import (
	"testing"

	config "github.com/a-castellano/home-ip-notifier/config"
)

// TestInvalidMailHost tests that SendEmail fails when given an unresolvable SMTP host
// This ensures the application properly handles SMTP connection failures
func TestInvalidMailHost(t *testing.T) {

	// Create test configuration with unresolvable host
	appConfig := config.Config{
		MailFrom:          "from@windmaker.net",
		MailDomain:        "windmaker.net",
		SMTPHost:          "non-resolvable-host",
		SMTPPort:          25,
		SMTPName:          "realname",
		SMTPPassword:      "secretPassword",
		SMTPTLSValidation: false,
		Destination:       "to@windmaker.net",
	}

	sendError := SendEmail(&appConfig, "Test message")
	if sendError == nil {
		t.Errorf("TestInvalidMailHost should fail.")
	}
}

// TestValidMailTLSError tests that SendEmail fails with TLS validation errors when certificates are invalid
// This ensures proper TLS certificate validation behavior
func TestValidMailTLSError(t *testing.T) {

	// Create test configuration with TLS validation enabled
	appConfig := config.Config{
		MailFrom:          "from@windmaker.net",
		MailDomain:        "windmaker.net",
		SMTPHost:          "mailhog",
		SMTPPort:          6465,
		SMTPName:          "realname",
		SMTPPassword:      "secretPassword",
		SMTPTLSValidation: true,
		Destination:       "to@windmaker.net",
	}

	sendError := SendEmail(&appConfig, "Test message")
	if sendError == nil {
		t.Errorf("TestInvalidMailHost should fail.")
	} else {
		// Verify the specific TLS error message
		expectedError := "tls: failed to verify certificate: x509: certificate is not valid for any names, but wanted to match mailhog"
		if sendError.Error() != expectedError {
			t.Errorf("TestValidMailTLSError error should be \"%s\" but it was \"%s\".", expectedError, sendError.Error())
		}
	}
}

// TestValidMail tests that SendEmail succeeds with valid configuration
// This ensures the email sending functionality works correctly in normal conditions
func TestValidMail(t *testing.T) {

	// Create test configuration with TLS validation disabled for testing
	appConfig := config.Config{
		MailFrom:          "from@windmaker.net",
		MailDomain:        "windmaker.net",
		SMTPHost:          "mailhog",
		SMTPPort:          6465,
		SMTPName:          "realname",
		SMTPPassword:      "secretPassword",
		SMTPTLSValidation: false,
		Destination:       "to@windmaker.net",
	}

	sendError := SendEmail(&appConfig, "Test message")
	if sendError != nil {
		t.Errorf("TestValidMail should not fail.")
	}
}

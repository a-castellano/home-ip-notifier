//go:build integration_tests

package mail

import (
	"testing"

	config "github.com/a-castellano/home-ip-notifier/config"
)

func TestInvalidMailHost(t *testing.T) {

	appConfig := config.Config{MailFrom: "from@windmaker.net", MailDomain: "windmaker.net", SMTPHost: "non-resolvable-host", SMTPPort: 25, SMTPName: "realname", SMTPPassword: "secretPassword", SMTPTLSValidation: false, Destination: "to@windmaker.net"}

	sendError := SendEmail(&appConfig, "Test message")
	if sendError == nil {
		t.Errorf("TestInvalidMailHost should fail.")
	}
}

func TestValidMailTLSError(t *testing.T) {

	appConfig := config.Config{MailFrom: "from@windmaker.net", MailDomain: "windmaker.net", SMTPHost: "mailhog", SMTPPort: 6465, SMTPName: "realname", SMTPPassword: "secretPassword", SMTPTLSValidation: true, Destination: "to@windmaker.net"}

	sendError := SendEmail(&appConfig, "Test message")
	if sendError == nil {
		t.Errorf("TestInvalidMailHost should fail.")
	} else {

		if sendError.Error() != "tls: failed to verify certificate: x509: certificate is not valid for any names, but wanted to match mailhog" {
			t.Errorf("TestValidMailTLSError error should be \"tls: failed to verify certificate: x509: certificate is not valid for any names, but wanted to match mailhog\" but it was \"%s\".", sendError.Error())
		}
	}

}

func TestValidMail(t *testing.T) {

	appConfig := config.Config{MailFrom: "from@windmaker.net", MailDomain: "windmaker.net", SMTPHost: "mailhog", SMTPPort: 6465, SMTPName: "realname", SMTPPassword: "secretPassword", SMTPTLSValidation: false, Destination: "to@windmaker.net"}

	sendError := SendEmail(&appConfig, "Test message")
	if sendError != nil {
		t.Errorf("TestValidMail should not fail.")
	}
}

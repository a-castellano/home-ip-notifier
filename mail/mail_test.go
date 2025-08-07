//go:build integration_tests

package mail

import (
	config "github.com/a-castellano/home-ip-notifier/config"
	"testing"
)

func TestInvalidMailHost(t *testing.T) {

	appConfig := config.Config{MailFrom: "from@windmaker.net", MailDomain: "windmaker.net", SMTPHost: "non-resolvable-host", SMTPPort: 25, SMTPName: "realname", SMTPPassword: "secretPassword", SMTPTLSValidation: false, Destination: "to@windmaker.net"}

	sendErr := SendEmail(&appConfig, "Test message")
	if sendErr == nil {
		t.Errorf("TestInvalidMailHost should fail.")
	}
}

func TestValidMailTLSError(t *testing.T) {

	appConfig := config.Config{MailFrom: "from@windmaker.net", MailDomain: "windmaker.net", SMTPHost: "stunnel", SMTPPort: 6465, SMTPName: "realname", SMTPPassword: "secretPassword", SMTPTLSValidation: true, Destination: "to@windmaker.net"}

	sendErr := SendEmail(&appConfig, "Test message")
	if sendErr == nil {
		t.Errorf("TestInvalidMailHost should fail.")
	} else {

		if sendErr.Error() != "tls: failed to verify certificate: x509: certificate is not valid for any names, but wanted to match stunnel" {
			t.Errorf("TestValidMailTLSError error should be \"tls: failed to verify certificate: x509: certificate is not valid for any names, but wanted to match stunnel\" but it was \"%s\".", sendErr.Error())
		}
	}

}

func TestValidMail(t *testing.T) {

	appConfig := config.Config{MailFrom: "from@windmaker.net", MailDomain: "windmaker.net", SMTPHost: "stunnel", SMTPPort: 6465, SMTPName: "realname", SMTPPassword: "secretPassword", SMTPTLSValidation: false, Destination: "to@windmaker.net"}

	sendErr := SendEmail(&appConfig, "Test message")
	if sendErr != nil {
		t.Errorf("TestValidMail should not fail.")
	}
}

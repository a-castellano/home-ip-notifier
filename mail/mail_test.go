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

func TestInvalidMailFrom(t *testing.T) {

	appConfig := config.Config{MailFrom: "from@windmaker.net", MailDomain: "windmaker.net", SMTPHost: "mailhog", SMTPPort: 1025, SMTPName: "realname", SMTPPassword: "secretPassword", SMTPTLSValidation: false, Destination: "to@windmaker.net"}

	sendErr := SendEmail(&appConfig, "Test message")
	if sendErr == nil {
		t.Errorf("TestInvalidMailHost should fail.")
	} else {

		if sendErr.Error() != "strconv.Atoi: parsing \"invalidport\": invalid syntax" {
			t.Errorf("TestConfigWithInvalidRedisPort error should be \"strconv.Atoi: parsing \"invalidport\": invalid syntax\" but it was \"%s\".", sendErr.Error())
		}
	}

}

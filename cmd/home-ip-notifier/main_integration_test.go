//go:build integration_tests

package main

import (
	"context"
	"encoding/json"
	rabbitmq "github.com/a-castellano/go-services/infra/rabbitmq"
	messagebroker "github.com/a-castellano/go-services/services/messagebroker"
	envelope "github.com/a-castellano/go-types/types/envelope"
	rabbitmqconfig "github.com/a-castellano/go-types/types/rabbitmq"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
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
	"app_name":          {VariableName: "APP_NAME"},
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

func TestInvalidOtelConfig(t *testing.T) {

	setUp()
	defer teardown()

	os.Setenv("SMTP_FROM", "test@example.com")
	os.Setenv("SMTP_HOST", "test")
	os.Setenv("SMTP_PORT", "25")
	os.Setenv("SMTP_USERNAME", "test")
	os.Setenv("SMTP_PASSWORD", "test")
	os.Setenv("SMTP_VALIDATE_TLS", "anyvaluedifferentfromtrue")

	os.Setenv("DESTINATION", "alvaro@windmaker.net")
	os.Setenv("NOTIFY_QUEUE_NAME", "other_queue")

	ctx := context.Background()
	err := run(ctx)

	if err == nil {
		t.Fatalf("TestInvalidOtelConfig should fail, otel config is invalid")
	}

}

func TestInvalidDestination(t *testing.T) {

	setUp()
	defer teardown()

	os.Setenv("APP_NAME", "home-ip-notifier")

	os.Setenv("SMTP_FROM", "test@example.com")
	os.Setenv("SMTP_HOST", "test")
	os.Setenv("SMTP_PORT", "25")
	os.Setenv("SMTP_USERNAME", "test")
	os.Setenv("SMTP_PASSWORD", "test")
	os.Setenv("SMTP_VALIDATE_TLS", "anyvaluedifferentfromtrue")

	os.Setenv("DESTINATION", "invalid")
	os.Setenv("NOTIFY_QUEUE_NAME", "other_queue")

	ctx := context.Background()
	err := run(ctx)

	if err == nil {
		t.Fatalf("TestInvalidDestination should fail, destination is invalid")
	}

	expectedErr := "mail: missing '@' or angle-addr"
	if err.Error() != expectedErr {
		t.Fatalf("TestInvalidDestination error should be \"%s\" but it was \"%s\".", expectedErr, err.Error())
	}

}

const mailhogAPI = "http://mailhog:8025"

// clearMailhogMessages empties the MailHog mailbox so leftover emails from
// previous runs cannot satisfy this run's assertions.
// This helper was written by an AI agent (Claude).
func clearMailhogMessages() error {
	request, requestErr := http.NewRequest(http.MethodDelete, mailhogAPI+"/api/v1/messages", nil)
	if requestErr != nil {
		return requestErr
	}
	response, deleteErr := http.DefaultClient.Do(request)
	if deleteErr != nil {
		return deleteErr
	}
	return response.Body.Close()
}

// mailhogHasMessage reports whether any email in the MailHog mailbox contains
// expectedBody in its body.
// This helper was written by an AI agent (Claude).
func mailhogHasMessage(expectedBody string) (bool, error) {
	response, getErr := http.Get(mailhogAPI + "/api/v2/messages")
	if getErr != nil {
		return false, getErr
	}
	defer response.Body.Close()

	var messages struct {
		Items []struct {
			Content struct {
				Body string
			}
		}
	}
	if decodeErr := json.NewDecoder(response.Body).Decode(&messages); decodeErr != nil {
		return false, decodeErr
	}

	for _, item := range messages.Items {
		if strings.Contains(item.Content.Body, expectedBody) {
			return true, nil
		}
	}
	return false, nil
}

// Parts of this test were written by an AI agent (Claude), delimited by
// "AI-written" comments in the body: the MailHog mailbox cleanup, the error
// checks on the rabbitmq config and the envelope marshalling, and the run
// lifecycle handling (error channel, polling and timeouts) from the run
// goroutine to the end. The environment setup and the message publishing are
// hand-written.
func TestRetrieveMessage(t *testing.T) {

	setUp()
	defer teardown()

	os.Setenv("APP_NAME", "home-ip-notifier")

	// Set environment variables for RabbitMQ with valid credentials
	os.Setenv("RABBITMQ_HOST", "rabbitmq")
	os.Setenv("RABBITMQ_PORT", "5672")
	os.Setenv("RABBITMQ_USER", "guest")
	os.Setenv("RABBITMQ_PASSWORD", "guest")

	os.Setenv("SMTP_FROM", "test@example.com")
	os.Setenv("SMTP_HOST", "mailhog")
	os.Setenv("SMTP_PORT", "6465")
	os.Setenv("SMTP_USERNAME", "test")
	os.Setenv("SMTP_PASSWORD", "test")
	os.Setenv("SMTP_VALIDATE_TLS", "false")

	os.Setenv("DESTINATION", "alvaro@windmaker.net")
	os.Setenv("NOTIFY_QUEUE_NAME", "home-ip-notifier-integration-tests")

	// AI-written (Claude): mailbox cleanup.
	if clearErr := clearMailhogMessages(); clearErr != nil {
		t.Fatalf("TestRetrieveMessage could not clear the MailHog mailbox, error was \"%s\"", clearErr.Error())
	}

	rabbitmqConfig, rabbitmqConfigErr := rabbitmqconfig.NewConfig()
	// AI-written (Claude): this error check.
	if rabbitmqConfigErr != nil {
		t.Fatalf("TestRetrieveMessage could not load rabbitmq config, error was \"%s\"", rabbitmqConfigErr.Error())
	}
	queueName := "home-ip-notifier-integration-tests"

	rabbitmqClient := rabbitmq.NewRabbitmqClient(rabbitmqConfig)
	messageBroker := messagebroker.MessageBroker{Client: rabbitmqClient}

	body := []byte("123.123.123.123")
	carrier := map[string]string{"traceparent": "00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01"}

	data, marshalErr := (&envelope.Envelope{Carrier: carrier, Body: body}).Marshal()
	// AI-written (Claude): this error check.
	if marshalErr != nil {
		t.Fatalf("TestRetrieveMessage could not marshal the test envelope, error was \"%s\"", marshalErr.Error())
	}

	sendError := messageBroker.SendMessage(context.Background(), queueName, data)

	if sendError != nil {
		t.Fatalf("TestRetrieveMessage should not fail when test message is sent, error was \"%s\"", sendError.Error())
	}

	// AI-written (Claude): everything from here to the end of the test — run
	// lifecycle handling (error channel, MailHog polling and timeouts).
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	runResult := make(chan error, 1)
	go func() {
		runResult <- run(ctx)
	}()

	// Poll MailHog until the notification email arrives; run exiting early or
	// the deadline expiring are both failures.
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()
	deadline := time.After(15 * time.Second)

	delivered := false
	for !delivered {
		select {
		case runErr := <-runResult:
			t.Fatalf("TestRetrieveMessage run returned before the message was delivered, error was \"%v\"", runErr)
		case <-deadline:
			t.Fatalf("TestRetrieveMessage timed out waiting for the notification email to reach MailHog")
		case <-ticker.C:
			found, checkErr := mailhogHasMessage("123.123.123.123")
			if checkErr != nil {
				t.Fatalf("TestRetrieveMessage could not query MailHog, error was \"%s\"", checkErr.Error())
			}
			delivered = found
		}
	}

	cancel()

	select {
	case runErr := <-runResult:
		if runErr != nil {
			t.Fatalf("TestRetrieveMessage run should finish without error after cancellation, error was \"%s\"", runErr.Error())
		}
	case <-time.After(5 * time.Second):
		t.Fatalf("TestRetrieveMessage run did not return after context cancellation")
	}

}

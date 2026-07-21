//go:build integration_tests || unit_tests

package consume

import (
	"context"
	"errors"
	envelope "github.com/a-castellano/go-types/types/envelope"
	"testing"
)

type mockProcessor struct {
	fail bool
}

func (m mockProcessor) ProcessMessage(ctx context.Context, message string) error {
	if m.fail {
		return errors.New("Fail")
	}
	return nil
}

func TestEmptyData(t *testing.T) {

	ctx := context.Background()
	consumer := NewConsumer(ctx, "testqueue", mockProcessor{})

	emptydata := make([]byte, 0)

	err := consumer.Consume(ctx, emptydata)

	if err != nil {
		t.Fatalf("TestEmptyData should not return error even is data to consume is empty, error was \"%s\"", err.Error())
	}

}

func TestEmptyBody(t *testing.T) {

	ctx := context.Background()
	consumer := NewConsumer(ctx, "testqueue", mockProcessor{})
	emptybody := make([]byte, 0)
	carrier := map[string]string{"testkey": "testValue"}

	data, _ := (&envelope.Envelope{Carrier: carrier, Body: emptybody}).Marshal()

	err := consumer.Consume(ctx, data)

	if err != nil {
		t.Fatalf("TestEmptyBody should not return error even is body to consume is empty, error was \"%s\"", err.Error())
	}

}

func TestValidEnvelopeConsumerFails(t *testing.T) {

	ctx := context.Background()
	consumer := NewConsumer(ctx, "testqueue", mockProcessor{fail: true})

	body := []byte("123.123.123.123")
	carrier := map[string]string{"traceparent": "00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01"}

	data, _ := (&envelope.Envelope{Carrier: carrier, Body: body}).Marshal()

	err := consumer.Consume(ctx, data)

	if err == nil {
		t.Fatalf("TestValidEnvelopeConsumerFails should fail")
	}

	expectedErr := "Fail"
	if err.Error() != expectedErr {
		t.Fatalf("TestValidEnvelopeConsumerFails error should be \"%s\" but it was \"%s\"", expectedErr, err.Error())
	}

}

func TestValidEnvelopeConsumer(t *testing.T) {

	ctx := context.Background()
	consumer := NewConsumer(ctx, "testqueue", mockProcessor{})

	body := []byte("123.123.123.123")
	carrier := map[string]string{"traceparent": "00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01"}

	data, _ := (&envelope.Envelope{Carrier: carrier, Body: body}).Marshal()

	err := consumer.Consume(ctx, data)

	if err != nil {
		t.Fatalf("TestValidEnvelopeConsumer should not fail, error was \"%s\"", err.Error())
	}

}

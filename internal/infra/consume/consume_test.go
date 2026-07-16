//go:build integration_tests || unit_tests

package consume

import (
	"context"
	"errors"
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
	consumer := NewConsumer("testqueue", mockProcessor{})

	emptydata := make([]byte, 0)

	err := consumer.Consume(ctx, emptydata)

	if err != nil {
		t.Fatalf("TestEmptyData should not return error even is data to consume is empty, error was \"%s\"", err.Error())
	}

}

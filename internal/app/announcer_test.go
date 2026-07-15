//go:build integration_tests || unit_tests

package app

import (
	"context"
	"errors"
	"testing"
)

type mockNotifier struct {
	failOnNotify bool
}

func (m mockNotifier) Notify(ctx context.Context, subject string, message string) error {
	if m.failOnNotify {
		return errors.New("fail")
	}
	return nil
}

func TestFailedNotify(t *testing.T) {

	ctx := context.Background()
	mock := mockNotifier{failOnNotify: true}

	announcer := NewAnnouncer(mock)
	err := announcer.ProcessMessage(ctx, "message")

	if err == nil {
		t.Fatalf("TestFailedNotify should fail.")
	}
	expectedError := "fail"

	if err.Error() != expectedError {
		t.Fatalf("TestFailedNotify error should be \"%s\" but it was \"%s\".", expectedError, err.Error())
	}
}

func TestNotify(t *testing.T) {

	ctx := context.Background()
	mock := mockNotifier{failOnNotify: false}

	announcer := NewAnnouncer(mock)
	err := announcer.ProcessMessage(ctx, "message")

	if err != nil {
		t.Fatalf("TestFailedNotify should not fail, error was \"%s\"", err.Error())
	}
}

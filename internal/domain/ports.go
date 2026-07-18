// Package domain holds the application-level ports of the service. It depends
// on nothing but the standard library: the interfaces declared here are
// implemented by the infra adapters and consumed by the app layer, so
// dependencies always point inwards.
package domain

import (
	"context"
)

// Notifier is the outbound port the use case delivers through. It is defined
// at the level this application needs — title and message; the destination
// does not vary per message, it is deployment configuration and lives in the
// adapter that implements the port.
type Notifier interface {
	Notify(ctx context.Context, title string, message string) error
}

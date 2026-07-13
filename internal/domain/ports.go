package domain

import (
	"context"
)

type Notifier interface {
	Notify(ctx context.Context, title string, message string) error
}

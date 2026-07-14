package app

import (
	"context"

	domain "github.com/a-castellano/home-ip-notifier/internal/domain"
)

const subject = "Home IP has changed"

type Announcer struct {
	notifier domain.Notifier
}

func NewAnnouncer(notifier domain.Notifier) Announcer {
	return Announcer{notifier: notifier}
}

func (a Announcer) ProcessMessage(ctx context.Context, message string) error {

	return a.notifier.Notify(ctx, subject, message)
}

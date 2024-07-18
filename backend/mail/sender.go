package mail

import (
	"context"

	gomail "github.com/wneessen/go-mail"
)

type Sender interface {
	Send(ctx context.Context, message *gomail.Msg) error
}

package jobqueue

import (
	"context"

	"myvendor.mytld/myproject/backend/domain/command"
)

//nolint:interfacebloat
type Queue interface {
	Close() error

	// AccountSendWelcomeEmail enqueues a task to send a welcome email to a newly created account.
	//
	// Note: It serves as an example, so the deferred processing parameter is included.
	AccountSendWelcomeEmail(ctx context.Context, cmd command.AccountSendWelcomeEmail, deferProcessing bool) error
}

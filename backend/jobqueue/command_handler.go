package jobqueue

import (
	"context"

	"myvendor.mytld/myproject/backend/domain/command"
)

// CommandHandler is an interface for command handler methods that will be used by a jobqueue server to process tasks.
// A handler.Handler will be used for the actual implementation and a mock can be used for testing.
type CommandHandler interface {
	AccountSendWelcomeEmail(ctx context.Context, cmd command.AccountSendWelcomeEmail) error
}

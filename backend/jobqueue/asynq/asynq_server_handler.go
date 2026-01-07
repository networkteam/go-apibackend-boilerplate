package asynq

import (
	"context"
	"encoding/json"

	"github.com/friendsofgo/errors"
	"github.com/hibiken/asynq"

	"myvendor.mytld/myproject/backend/domain/command"
	"myvendor.mytld/myproject/backend/domain/types"
)

func (s *Server) handleTaskAccountSendWelcomeEmail(ctx context.Context, task *asynq.Task) error {
	var cmd command.AccountSendWelcomeEmail
	err := json.Unmarshal(task.Payload(), &cmd)
	if err != nil {
		return errors.Wrap(err, "unmarshalling payload")
	}

	err = s.handler.AccountSendWelcomeEmail(ctx, cmd)
	if errors.Is(err, types.ErrHandlerLocked) {
		s.logger.WarnContext(ctx, "Handler is locked by another process, skipping task", "taskType", task.Type())
		return asynq.RevokeTask
	}

	return err
}

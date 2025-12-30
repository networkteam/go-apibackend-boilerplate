package asynq_test

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gofrs/uuid"
	asynqPkg "github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"myvendor.mytld/myproject/backend/domain/command"
	"myvendor.mytld/myproject/backend/domain/types"
	"myvendor.mytld/myproject/backend/jobqueue/asynq"
)

func defaultServerOpts() asynq.ServerOpt {
	return asynq.ServerOpt{
		Concurrency:     1,
		ShutdownTimeout: 5 * time.Second,
		RetryDelayFunc: func(n int, e error, t *asynqPkg.Task) time.Duration {
			return 50 * time.Millisecond
		},
		DelayedTaskCheckInterval: 500 * time.Millisecond,
	}
}

func createTestQueueAndServer(t *testing.T, queueOpts asynq.QueueOpt, serverOpts asynq.ServerOpt) (*asynq.Queue, *asynqPkg.Inspector, *CommandHandlerMock) {
	s := miniredis.RunT(t)

	redisOpt := asynqPkg.RedisClientOpt{
		Addr: s.Addr(),
	}

	queue := asynq.NewQueue(redisOpt, queueOpts)
	ctx := context.Background()

	mockCommandHandler := &CommandHandlerMock{}
	srv := asynq.NewAsynqServer(ctx, redisOpt, serverOpts, mockCommandHandler)
	t.Cleanup(func() {
		srv.Shutdown()
	})

	err := srv.Start()
	require.NoError(t, err)

	inspector := asynqPkg.NewInspector(redisOpt)
	t.Cleanup(func() {
		inspector.Close()
	})

	return queue, inspector, mockCommandHandler
}

func TestQueue_AccountSendWelcomeEmail(t *testing.T) {
	t.Parallel()

	queue, _, mockCommandHandler := createTestQueueAndServer(t, asynq.DefaultQueueOpts(), defaultServerOpts())
	ctx := context.Background()

	// Queue the task

	cmd := command.AccountSendWelcomeEmail{
		AccountID:    uuid.FromStringOrNil("00000000-0000-0000-0000-000000000001"),
		EmailAddress: "test@example.com",
	}
	err := queue.AccountSendWelcomeEmail(ctx, cmd, false)
	require.NoError(t, err)

	// Check that the task was processed by the server

	require.Eventually(t, func() bool {
		return len(mockCommandHandler.AccountSendWelcomeEmailCalls()) > 0
	}, 10*time.Second, 50*time.Millisecond)
	assert.Equal(t, cmd, mockCommandHandler.AccountSendWelcomeEmailCalls()[0].Cmd, "command should match")
}

func TestQueue_AccountSendWelcomeEmail_Defer_And_Deduplicate(t *testing.T) {
	t.Parallel()

	opts := asynq.DefaultQueueOpts()
	config := opts.Tasks[asynq.TaskTypeAccountSendWelcomeEmail]
	config.DeferredProcessingDelay = 3 * time.Second
	opts.Tasks[asynq.TaskTypeAccountSendWelcomeEmail] = config
	queue, _, mockCommandHandler := createTestQueueAndServer(t, opts, defaultServerOpts())
	ctx := context.Background()

	// Queue the task multiple times

	cmd := command.AccountSendWelcomeEmail{
		AccountID:    uuid.FromStringOrNil("00000000-0000-0000-0000-000000000001"),
		EmailAddress: "test@example.com",
	}
	err := queue.AccountSendWelcomeEmail(ctx, cmd, true)
	require.NoError(t, err)
	err = queue.AccountSendWelcomeEmail(ctx, cmd, true)
	require.NoError(t, err)
	err = queue.AccountSendWelcomeEmail(ctx, cmd, true)
	require.NoError(t, err)

	// Check that the task was processed by the server (give a little more time, since the deferred processing takes more time to pick up the task)

	time.Sleep(1*time.Second + 500*time.Millisecond)
	require.Len(t, mockCommandHandler.AccountSendWelcomeEmailCalls(), 0, "command should not be processed as it is deferred")

	require.Eventually(t, func() bool {
		return len(mockCommandHandler.AccountSendWelcomeEmailCalls()) > 0
	}, 30*time.Second, 50*time.Millisecond)
	assert.Len(t, mockCommandHandler.AccountSendWelcomeEmailCalls(), 1, "command should be deduplicated")
	assert.Equal(t, cmd, mockCommandHandler.AccountSendWelcomeEmailCalls()[0].Cmd, "command should match")
}

func TestQueue_ErrorHandling_Retry(t *testing.T) {
	t.Parallel()

	opts := asynq.DefaultQueueOpts()
	config := opts.Tasks[asynq.TaskTypeAccountSendWelcomeEmail]
	config.Opts.MaxRetry = 3
	opts.Tasks[asynq.TaskTypeAccountSendWelcomeEmail] = config
	queue, _, mockCommandHandler := createTestQueueAndServer(t, opts, defaultServerOpts())
	i := 0
	mockCommandHandler.AccountSendWelcomeEmailFunc = func(ctx context.Context, cmd command.AccountSendWelcomeEmail) error {
		if i < 3 {
			i++
			return assert.AnError
		}
		return nil
	}

	ctx := context.Background()

	// Queue the task

	cmd := command.AccountSendWelcomeEmail{
		AccountID:    uuid.FromStringOrNil("00000000-0000-0000-0000-000000000001"),
		EmailAddress: "test@example.com",
	}
	err := queue.AccountSendWelcomeEmail(ctx, cmd, false)
	require.NoError(t, err)

	// Check that the task was processed by the server

	require.Eventually(t, func() bool {
		return len(mockCommandHandler.AccountSendWelcomeEmailCalls()) >= 3
	}, 10*time.Second, 50*time.Millisecond)
}

func TestQueue_HandlerLocked_Revokes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		queueFunc func(q *asynq.Queue, ctx context.Context) error
		setupMock func(mock *CommandHandlerMock, callCount *int)
		queueName string
	}{
		{
			// Note: using AccountSendWelcomeEmail command here just for testing purposes
			name: "AccountSendWelcomeEmail",
			queueFunc: func(q *asynq.Queue, ctx context.Context) error {
				return q.AccountSendWelcomeEmail(ctx, command.AccountSendWelcomeEmail{
					AccountID:    uuid.FromStringOrNil("00000000-0000-0000-0000-000000000002"),
					EmailAddress: "test@example.com",
				}, false)
			},
			setupMock: func(mock *CommandHandlerMock, callCount *int) {
				mock.AccountSendWelcomeEmailFunc = func(ctx context.Context, cmd command.AccountSendWelcomeEmail) error {
					*callCount++
					if *callCount == 1 {
						return nil
					}
					return types.ErrHandlerLocked
				}
			},
			queueName: asynq.QueueNameNotifications,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			serverOpts := defaultServerOpts()
			serverOpts.RetryDelayFunc = func(n int, e error, t *asynqPkg.Task) time.Duration {
				return 5 * time.Second
			}

			queue, inspector, mockCommandHandler := createTestQueueAndServer(t, asynq.DefaultQueueOpts(), serverOpts)
			ctx := context.Background()

			callCount := 0
			tt.setupMock(mockCommandHandler, &callCount)

			// Queue first task - should succeed

			err := tt.queueFunc(queue, ctx)
			require.NoError(t, err)

			require.Eventually(t, func() bool {
				return callCount == 1
			}, 5*time.Second, 50*time.Millisecond, "first task should be processed successfully")

			// Queue second task - should return ErrHandlerLocked and be revoked

			err = tt.queueFunc(queue, ctx)
			require.NoError(t, err)

			require.Eventually(t, func() bool {
				return callCount == 2
			}, 5*time.Second, 50*time.Millisecond, "second task should be called and return locked error")

			// Wait to ensure no retry is scheduled

			time.Sleep(2 * time.Second)

			// Verify no retry tasks in queue

			retryTasks, err := inspector.ListRetryTasks(tt.queueName)
			require.NoError(t, err)
			assert.Empty(t, retryTasks, "no retry tasks should be in queue")

			// Verify no active tasks

			activeTasks, err := inspector.ListActiveTasks(tt.queueName)
			require.NoError(t, err)
			assert.Empty(t, activeTasks, "no active tasks should be in queue")

			// Verify no pending tasks

			pendingTasks, err := inspector.ListPendingTasks(tt.queueName)
			require.NoError(t, err)
			assert.Empty(t, pendingTasks, "no pending tasks should be in queue")

			// Final assertion: handler was called exactly twice

			assert.Equal(t, 2, callCount, "handler should be called exactly twice")
		})
	}
}

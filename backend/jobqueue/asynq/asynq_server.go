package asynq

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/friendsofgo/errors"
	"github.com/hibiken/asynq"
	"github.com/networkteam/slogutils"
	"golang.org/x/sys/unix"

	"myvendor.mytld/myproject/backend/domain/types"
	"myvendor.mytld/myproject/backend/jobqueue"
	"myvendor.mytld/myproject/backend/security/authentication"
)

type ServerOpt struct {
	Concurrency              int
	ShutdownTimeout          time.Duration
	RetryDelayFunc           asynq.RetryDelayFunc
	DelayedTaskCheckInterval time.Duration
}

type Server struct {
	srv     *asynq.Server
	opt     ServerOpt
	mux     *asynq.ServeMux
	handler jobqueue.CommandHandler
	logger  *slog.Logger
}

func (s *Server) Start() error {
	err := s.srv.Start(s.mux)
	if err != nil {
		return errors.Wrap(err, "starting Asynq server")
	}
	return nil
}

func (s *Server) WaitForSignals() {
	s.logger.Info("Send signal TSTP to stop processing new tasks")
	s.logger.Info("Send signal TERM or INT to terminate the process")

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, unix.SIGTERM, unix.SIGINT, unix.SIGTSTP)
	for {
		sig := <-sigs
		if sig == unix.SIGTSTP {
			s.srv.Stop()
			continue
		}
		break
	}
}

func (s *Server) Shutdown() {
	s.srv.Shutdown()
}

func NewAsynqServer(ctx context.Context, redisOpt asynq.RedisClientOpt, serverOpt ServerOpt, handler jobqueue.CommandHandler) *Server {
	logger := slogutils.FromContext(ctx).With("component", "jobqueue")
	srv := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Concurrency:     serverOpt.Concurrency,
			Logger:          asynqSlogLogger{logger: logger},
			ShutdownTimeout: serverOpt.ShutdownTimeout,
			Queues: map[string]int{
				QueueNameNotifications: 6,
			},
			BaseContext: func() context.Context {
				return authentication.WithAuthContext(ctx, authentication.AuthContext{
					Authenticated: false,
					Role:          types.RoleJobqueue,
				})
			},
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				retried, _ := asynq.GetRetryCount(ctx)
				maxRetry, _ := asynq.GetMaxRetry(ctx)
				if maxRetry > 0 && retried >= maxRetry {
					err = fmt.Errorf("retries exhausted (%d/%d) for task %s: %w", retried, maxRetry, task.Type(), err)
				}

				logger.ErrorContext(ctx, "Processing of task failed", "taskType", task.Type(), slogutils.Err(err))
			}),
			IsFailure: func(err error) bool {
				// Locked errors shouldn't count as failures, we will just retry them
				return !errors.Is(err, types.ErrHandlerLocked)
			},
			RetryDelayFunc:           serverOpt.RetryDelayFunc,
			DelayedTaskCheckInterval: serverOpt.DelayedTaskCheckInterval,
		},
	)

	mux := asynq.NewServeMux()

	asynqServer := &Server{
		srv:     srv,
		mux:     mux,
		opt:     serverOpt,
		handler: handler,
		logger:  logger,
	}

	mux.HandleFunc(
		TaskTypeAccountSendWelcomeEmail,
		asynqServer.handleTaskAccountSendWelcomeEmail,
	)

	return asynqServer
}

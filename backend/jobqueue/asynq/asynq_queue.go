package asynq

import (
	"context"
	"encoding/json"
	"time"

	"github.com/friendsofgo/errors"
	"github.com/hibiken/asynq"
	"github.com/networkteam/slogutils"

	"myvendor.mytld/myproject/backend/domain/command"
	"myvendor.mytld/myproject/backend/jobqueue"
)

const (
	QueueNameNotifications = "notifications"
)

const (
	TaskTypeAccountSendWelcomeEmail = "accountSendWelcomeEmail"
)

type Queue struct {
	client    *asynq.Client
	queueOpts QueueOpt
}

type TaskOpts struct {
	// MaxRetry is the number of maximum retries for the task.
	MaxRetry int
	// Timeout is the timeout for each task run.
	Timeout   time.Duration
	UniqueTTL time.Duration
	ProcessIn time.Duration
}

type TaskConfig struct {
	// Task execution configuration
	Opts TaskOpts

	// DeferredProcessingDelay is the delay before processing when deferProcessing=true
	DeferredProcessingDelay time.Duration

	// QueueName is the asynq queue name for this task type
	QueueName string

	// RequireUnique enforces that UniqueTTL must be set (validates configuration)
	RequireUnique bool

	// CLI metadata for auto-generating flags
	FlagPrefix string // e.g. "jobqueue-task-account-send-welcome-email"
	EnvPrefix  string // e.g. "JOBQUEUE_TASK_ACCOUNT_SEND_WELCOME_EMAIL"
	Usage      string // e.g. "account send welcome email task"
}

var _ jobqueue.Queue = &Queue{}

type QueueOpt struct {
	// Tasks is a map of task type name -> configuration
	Tasks map[string]TaskConfig
}

func DefaultQueueOpts() QueueOpt {
	return QueueOpt{
		Tasks: map[string]TaskConfig{
			TaskTypeAccountSendWelcomeEmail: {
				Opts: TaskOpts{
					MaxRetry:  5,
					Timeout:   1 * time.Hour,
					UniqueTTL: 8 * time.Hour,
				},
				DeferredProcessingDelay: 5 * time.Minute,
				QueueName:               QueueNameNotifications,
				RequireUnique:           true,
				FlagPrefix:              "jobqueue-task-account-send-welcome-email",
				EnvPrefix:               "JOBQUEUE_TASK_ACCOUNT_SEND_WELCOME_EMAIL",
				Usage:                   "account send welcome email task",
			},
		},
	}
}

func NewQueue(redisOpt asynq.RedisClientOpt, queueOpt QueueOpt) *Queue {
	// Validate task configuration
	for taskType, config := range queueOpt.Tasks {
		if config.RequireUnique && config.Opts.UniqueTTL == 0 {
			panic(errors.Errorf("task %s requires UniqueTTL to be set but got 0", taskType))
		}
	}

	return &Queue{
		client:    asynq.NewClient(redisOpt),
		queueOpts: queueOpt,
	}
}

// Here we implement the jobqueue.Queue interface methods

func (q *Queue) AccountSendWelcomeEmail(ctx context.Context, cmd command.AccountSendWelcomeEmail, deferProcessing bool) error {
	return q.enqueueTask(ctx, TaskTypeAccountSendWelcomeEmail, cmd, deferProcessing)
}

// enqueueTask is a common helper for enqueueing tasks with consistent behavior
func (q *Queue) enqueueTask(ctx context.Context, taskType string, payload any, deferProcessing bool) error {
	config, ok := q.queueOpts.Tasks[taskType]
	if !ok {
		return errors.Errorf("task configuration not found: %s", taskType)
	}

	taskOpts := config.Opts
	if deferProcessing && config.DeferredProcessingDelay > 0 {
		taskOpts.ProcessIn = config.DeferredProcessingDelay
	}

	logger := slogutils.FromContext(ctx).
		With(
			"component", "jobqueue",
			"taskType", taskType,
			"taskOpts", taskOpts,
			"queueName", config.QueueName,
		)
	logger.Log(ctx, slogutils.LevelTrace, "Queueing task")

	task, err := q.newTask(taskType, payload, taskOpts)
	if err != nil {
		return errors.Wrap(err, "creating task")
	}

	info, err := q.client.EnqueueContext(ctx, task, asynq.Queue(config.QueueName))
	if errors.Is(err, asynq.ErrDuplicateTask) {
		logger.DebugContext(ctx, "Duplicate task in queue")
		return nil
	} else if err != nil {
		return errors.Wrap(err, "enqueueing task")
	}

	logger.DebugContext(ctx, "Enqueued task in queue", "taskID", info.ID)
	return nil
}

func (q *Queue) newTask(taskType string, payloadData any, taskOpts TaskOpts) (*asynq.Task, error) {
	var options = []asynq.Option{
		asynq.MaxRetry(taskOpts.MaxRetry),
		asynq.Timeout(taskOpts.Timeout),
	}

	if taskOpts.UniqueTTL > 0 {
		options = append(options, asynq.Unique(taskOpts.UniqueTTL))
	}

	if taskOpts.ProcessIn > 0 {
		options = append(options, asynq.ProcessIn(taskOpts.ProcessIn))
	}

	payloadBytes, err := json.Marshal(payloadData)
	if err != nil {
		return nil, errors.Wrap(err, "marshalling payload")
	}

	return asynq.NewTask(
		taskType,
		payloadBytes,
		options...,
	), nil
}

func (q *Queue) Close() error {
	return q.client.Close()
}

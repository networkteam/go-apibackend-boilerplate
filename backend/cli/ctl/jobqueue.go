package main

import (
	"time"

	"github.com/hibiken/asynq"
	"github.com/networkteam/slogutils"
	"github.com/urfave/cli/v2"

	"myvendor.mytld/myproject/backend/domain/handler"
	"myvendor.mytld/myproject/backend/jobqueue"
	jobqueue_asynq "myvendor.mytld/myproject/backend/jobqueue/asynq"
)

func createAsynqRedisClientOpt(c *cli.Context) asynq.RedisClientOpt {
	return asynq.RedisClientOpt{
		Addr:     c.String("jobqueue-redis-address"),
		Password: c.String("jobqueue-redis-password"),
		DB:       c.Int("jobqueue-redis-db"),
	}
}

func createAsynqServer(c *cli.Context, handler *handler.Handler) *jobqueue_asynq.Server {
	return jobqueue_asynq.NewAsynqServer(c.Context, createAsynqRedisClientOpt(c), jobqueue_asynq.ServerOpt{
		Concurrency:     c.Int("jobqueue-concurrency"),
		ShutdownTimeout: c.Duration("jobqueue-shutdown-timeout"),
	}, handler)
}

func createAsynqQueue(c *cli.Context) jobqueue.Queue {
	logger := slogutils.FromContext(c.Context)

	if c.String("jobqueue-redis-address") != "" {
		opts := jobqueue_asynq.DefaultQueueOpts()

		// Read task configurations from CLI flags
		for taskType, config := range opts.Tasks {
			// Read max retry
			if c.IsSet(config.FlagPrefix + "-max-retry") {
				config.Opts.MaxRetry = c.Int(config.FlagPrefix + "-max-retry")
			}

			// Read timeout
			if c.IsSet(config.FlagPrefix + "-timeout") {
				config.Opts.Timeout = c.Duration(config.FlagPrefix + "-timeout")
			}

			// Read unique TTL
			if c.IsSet(config.FlagPrefix + "-unique-ttl") {
				config.Opts.UniqueTTL = c.Duration(config.FlagPrefix + "-unique-ttl")
			}

			// Read deferred processing delay (if task supports it)
			if config.DeferredProcessingDelay > 0 && c.IsSet(config.FlagPrefix+"-deferred-processing-delay") {
				config.DeferredProcessingDelay = c.Duration(config.FlagPrefix + "-deferred-processing-delay")
			}

			// Update the config in the map
			opts.Tasks[taskType] = config
		}

		return jobqueue_asynq.NewQueue(
			createAsynqRedisClientOpt(c),
			opts,
		)
	}

	logger.Warn("No jobqueue-redis-address set, using fixture job queue (do not use in production)")
	return jobqueue.NewFixture()
}

func jobqueueMainFlags() []cli.Flag {
	defaultOpts := jobqueue_asynq.DefaultQueueOpts()

	flags := []cli.Flag{
		&cli.StringFlag{
			Name:    "jobqueue-redis-address",
			Usage:   "Job queue Redis address (empty to use in-memory job queue)",
			EnvVars: []string{"JOBQUEUE_REDIS_ADDRESS"},
		},
		&cli.StringFlag{
			Name:    "jobqueue-redis-password",
			Usage:   "Job queue Redis password",
			EnvVars: []string{"JOBQUEUE_REDIS_PASSWORD"},
		},
		&cli.IntFlag{
			Name:    "jobqueue-redis-db",
			Usage:   "Job queue Redis DB",
			EnvVars: []string{"JOBQUEUE_REDIS_DB"},
		},
	}

	// Auto-generate flags for each task type
	for _, config := range defaultOpts.Tasks {
		// MaxRetry flag
		flags = append(flags, &cli.IntFlag{
			Name:    config.FlagPrefix + "-max-retry",
			Usage:   "Maximum retries for " + config.Usage,
			Value:   config.Opts.MaxRetry,
			EnvVars: []string{config.EnvPrefix + "_MAX_RETRY"},
		})

		// Timeout flag
		flags = append(flags, &cli.DurationFlag{
			Name:    config.FlagPrefix + "-timeout",
			Usage:   "Timeout for " + config.Usage,
			Value:   config.Opts.Timeout,
			EnvVars: []string{config.EnvPrefix + "_TIMEOUT"},
		})

		// UniqueTTL flag (only for tasks that require uniqueness)
		if config.RequireUnique || config.Opts.UniqueTTL > 0 {
			flags = append(flags, &cli.DurationFlag{
				Name:    config.FlagPrefix + "-unique-ttl",
				Usage:   "Duration of uniqueness for " + config.Usage,
				Value:   config.Opts.UniqueTTL,
				EnvVars: []string{config.EnvPrefix + "_UNIQUE_TTL"},
			})
		}

		// Deferred processing delay flag (only for tasks that support it)
		if config.DeferredProcessingDelay > 0 {
			flags = append(flags, &cli.DurationFlag{
				Name:    config.FlagPrefix + "-deferred-processing-delay",
				Usage:   "Task processing delay for " + config.Usage + " (if deferred processing is enabled)",
				Value:   config.DeferredProcessingDelay,
				EnvVars: []string{config.EnvPrefix + "_DEFERRED_PROCESSING_DELAY"},
			})
		}
	}

	return flags
}

func jobqueueWorkerFlags() []cli.Flag {
	return []cli.Flag{
		&cli.IntFlag{
			Name:    "jobqueue-concurrency",
			Usage:   "Maximum amount of concurrent tasks executed by the worker",
			Value:   10,
			EnvVars: []string{"JOBQUEUE_CONCURRENCY"},
		},
		&cli.DurationFlag{
			Name:  "jobqueue-shutdown-timeout",
			Usage: "Shutdown timeout to finish all pending tasks",
			Value: 16 * time.Second,
		},
	}
}

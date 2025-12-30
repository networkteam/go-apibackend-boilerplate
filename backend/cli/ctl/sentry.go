package main

import (
	"log/slog"
	"os"

	"github.com/friendsofgo/errors"
	"github.com/getsentry/sentry-go"
	sentryslog "github.com/getsentry/sentry-go/slog"
	"github.com/networkteam/slogutils"
	slogmulti "github.com/samber/slog-multi"
	"github.com/urfave/cli/v2"

	integration_sentry "myvendor.mytld/myproject/backend/integration/sentry"
)

func sentryFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    "sentry-dsn",
			Usage:   "Sentry DSN (will be disabled if empty)",
			EnvVars: []string{"SENTRY_DSN"},
		},
		&cli.StringFlag{
			Name:    "sentry-environment",
			Usage:   "Sentry environment",
			EnvVars: []string{"SENTRY_ENVIRONMENT"},
			Value:   "development",
		},
		&cli.StringFlag{
			Name:    "sentry-release",
			Usage:   "Release version for Sentry",
			EnvVars: []string{"SENTRY_RELEASE"},
		},
		&cli.BoolFlag{
			Name:    "sentry-debug",
			Usage:   "Enable debug logging for Sentry",
			EnvVars: []string{"SENTRY_DEBUG"},
		},
		&cli.BoolFlag{
			Name:    "sentry-log-integration",
			Usage:   "Enable slog integration for Sentry",
			EnvVars: []string{"SENTRY_LOG_INTEGRATION"},
		},
	}
}

func initializeSentry(c *cli.Context, process string) error {
	logger := slogutils.FromContext(c.Context)

	sentryDSN := c.String("sentry-dsn")
	sentryEnvironment := c.String("sentry-environment")
	sentryRelease := c.String("sentry-release")
	sentryDebug := c.Bool("sentry-debug")
	sentryLogIntegration := c.Bool("sentry-log-integration")

	if sentryDSN == "" {
		logger.Info("No Sentry DSN set: Sentry disabled")

		return nil
	}

	sentryOptions := sentry.ClientOptions{
		Dsn:         sentryDSN,
		Environment: sentryEnvironment,
		Release:     sentryRelease,
		DebugWriter: os.Stderr,
		Debug:       sentryDebug,
		EnableLogs:  sentryLogIntegration,
	}

	logger.Info(
		"Initializing Sentry",
		"dsn", sentryDSN,
		"environment", sentryEnvironment,
		"release", sentryRelease,
	)

	sentry.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetTags(map[string]string{"process": process})
	})

	err := sentry.Init(sentryOptions)
	if err != nil {
		return errors.Wrap(err, "initializing Sentry")
	}

	return nil
}

func initializeSentrySlog(c *cli.Context) *slog.Logger {
	ctx := c.Context

	baseLogger := slogutils.FromContext(c.Context)

	defaultHandler := baseLogger.Handler()
	handler := sentryslog.Option{
		Converter:  integration_sentry.ConvertSlogToEvent,
		EventLevel: []slog.Level{slog.LevelError},
		LogLevel:   []slog.Level{slog.LevelInfo, slog.LevelWarn, slog.LevelError, sentryslog.LevelFatal},
	}.NewSentryHandler(ctx)

	logger := slog.New(
		slogmulti.Fanout(handler, defaultHandler),
	)

	baseLogger.InfoContext(c.Context, "Enabled sentry slog integration")

	return setDefaultLoggerAndContext(c, logger)
}

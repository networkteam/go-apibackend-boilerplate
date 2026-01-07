package main

import (
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5/tracelog"
	"github.com/mattn/go-isatty"
	"github.com/networkteam/slogutils"
	"github.com/networkteam/slogutils/buffering"
	"github.com/urfave/cli/v2"
)

// setDefaultLoggerAndContext sets a new logger as the default logger and for the CLI context.
// It is important to call this function instead of slog.SetDefault directly, so that the CLI context
// also gets the new logger and can forward it to other parts of the application.
func setDefaultLoggerAndContext(c *cli.Context, logger *slog.Logger) *slog.Logger {
	slog.SetDefault(logger)
	c.Context = slogutils.WithLogger(c.Context, logger)
	return logger
}

func setLogProcess(c *cli.Context, process string) *slog.Logger {
	// Use a logger instance with predeclared component field
	logger := slog.With("process", process)
	return setDefaultLoggerAndContext(c, logger)
}

func setupInitialLogging() *buffering.Handler {
	bufferingHandler := buffering.New()
	// Here it is okay to use slog.SetDefault since we don't have a context yet
	slog.SetDefault(slog.New(bufferingHandler))
	return bufferingHandler
}

func setLogHandler(c *cli.Context) slog.Handler {
	var handler slog.Handler
	logLevel := verbosityToSlogLevel(c.Int("verbosity"))
	if !c.Bool("disable-ansi") && (isatty.IsTerminal(os.Stdout.Fd()) || c.Bool("force-ansi")) {
		handler = slogutils.NewCLIHandler(os.Stderr, &slogutils.CLIHandlerOptions{
			Level: logLevel,
		})
	} else {
		handler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: logLevel,
		})
	}
	setDefaultLoggerAndContext(c, slog.New(handler))

	return handler
}

func verbosityToSlogLevel(verbosity int) slog.Level {
	if verbosity <= 1 {
		return slog.LevelError
	}

	switch verbosity {
	case 2:
		return slog.LevelWarn
	case 3:
		return slog.LevelInfo
	case 4:
		return slog.LevelDebug
	}

	return slogutils.LevelTrace
}

func verbosityToTracelogLevel(verbosity int) tracelog.LogLevel {
	if verbosity <= 1 {
		return tracelog.LogLevelNone
	}

	switch verbosity {
	case 2:
		return tracelog.LogLevelError
	case 3:
		return tracelog.LogLevelWarn
	case 4:
		return tracelog.LogLevelWarn
	case 5:
		return tracelog.LogLevelInfo
	}

	return tracelog.LogLevelDebug
}

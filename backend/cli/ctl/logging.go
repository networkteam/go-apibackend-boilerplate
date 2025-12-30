package main

import (
	"log/slog"

	"github.com/networkteam/slogutils"
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

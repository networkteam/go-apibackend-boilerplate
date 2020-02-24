package logger

import (
	"context"

	"github.com/apex/log"
)

type ctxKey int

const loggerKey ctxKey = iota

type Interface interface {
	log.Interface
}

// WithLogger returns a context with the given logger
func WithLogger(ctx context.Context, logger Interface) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// GetLogger gets the logger with field context or the default logger
func GetLogger(ctx context.Context) Interface {
	if logger, ok := ctx.Value(loggerKey).(Interface); ok {
		return logger
	}
	return log.Log
}

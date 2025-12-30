package asynq

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/hibiken/asynq"
)

type asynqSlogLogger struct {
	logger *slog.Logger
}

var _ asynq.Logger = asynqSlogLogger{}

func (a asynqSlogLogger) Debug(args ...any) {
	a.logger.Debug(fmt.Sprint(args...))
}

func (a asynqSlogLogger) Info(args ...any) {
	a.logger.Info(fmt.Sprint(args...))
}

func (a asynqSlogLogger) Warn(args ...any) {
	a.logger.Warn(fmt.Sprint(args...))
}

func (a asynqSlogLogger) Error(args ...any) {
	// Ignore task ID conflicts to prevent logging of errors when multiple schedulers are running
	if len(args) > 0 {
		if err, ok := args[0].(error); ok {
			// Error is not wrapped, so we resort to string matching...
			if strings.HasSuffix(err.Error(), asynq.ErrTaskIDConflict.Error()) {
				a.logger.Debug("Ignored task ID conflict")
				return
			}
		}
	}

	a.logger.Error(fmt.Sprint(args...))
}

func (a asynqSlogLogger) Fatal(args ...any) {
	a.logger.Error(fmt.Sprint(args...))
	os.Exit(1)
}

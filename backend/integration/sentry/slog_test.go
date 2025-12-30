package sentry_test

import (
	"errors"
	"log/slog"
	"testing"
	"time"

	fog_errors "github.com/friendsofgo/errors"
	sentrygo "github.com/getsentry/sentry-go"
	"github.com/networkteam/slogutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"myvendor.mytld/myproject/backend/domain/types"
	"myvendor.mytld/myproject/backend/integration/sentry"
)

// testSkippableError is a test error that implements ErrorReportSkipper
type testSkippableError struct {
	msg string
}

func (e testSkippableError) Error() string {
	return e.msg
}

func (e testSkippableError) ErrorReportSkip() {}

// Verify it implements the interface
var _ types.ErrorReportSkipper = testSkippableError{}

func TestConvertSlogToEvent(t *testing.T) {
	tests := []struct {
		name        string
		loggerAttr  []slog.Attr
		recordAttrs []slog.Attr
		expectNil   bool
	}{
		{
			name:        "skip event when sentinel in logger attributes",
			loggerAttr:  []slog.Attr{types.SentrySkipEventAttr()},
			recordAttrs: []slog.Attr{},
			expectNil:   true,
		},
		{
			name:        "skip event when sentinel in record attributes",
			loggerAttr:  []slog.Attr{},
			recordAttrs: []slog.Attr{types.SentrySkipEventAttr()},
			expectNil:   true,
		},
		{
			name:        "create event when no sentinel attribute",
			loggerAttr:  []slog.Attr{},
			recordAttrs: []slog.Attr{slog.String("key", "value")},
			expectNil:   false,
		},
		{
			name:        "create event with empty attributes",
			loggerAttr:  []slog.Attr{},
			recordAttrs: []slog.Attr{},
			expectNil:   false,
		},
		{
			name:        "skip event when error implements ErrorReportSkipper",
			loggerAttr:  []slog.Attr{},
			recordAttrs: []slog.Attr{slogutils.Err(testSkippableError{msg: "expected error"})},
			expectNil:   true,
		},
		{
			name:        "skip event when error wraps ErrorReportSkipper",
			loggerAttr:  []slog.Attr{},
			recordAttrs: []slog.Attr{slogutils.Err(fog_errors.Wrap(testSkippableError{msg: "expected error"}, "doing something"))},
			expectNil:   true,
		},
		{
			name:        "create event when error does not implement ErrorReportSkipper",
			loggerAttr:  []slog.Attr{},
			recordAttrs: []slog.Attr{slogutils.Err(errors.New("unexpected error"))},
			expectNil:   false,
		},
		{
			name:       "create event when mixed skippable and non-skippable errors",
			loggerAttr: []slog.Attr{},
			recordAttrs: []slog.Attr{
				slogutils.Err(testSkippableError{msg: "expected error"}),
				slog.Any("error2", errors.New("important error")),
			},
			expectNil: false,
		},
		{
			name:       "skip event when multiple skippable errors",
			loggerAttr: []slog.Attr{},
			recordAttrs: []slog.Attr{
				slogutils.Err(testSkippableError{msg: "error1"}),
				slog.Any("error2", testSkippableError{msg: "error2"}),
			},
			expectNil: true,
		},
		{
			name:       "create event when multiple non-skippable errors",
			loggerAttr: []slog.Attr{},
			recordAttrs: []slog.Attr{
				slogutils.Err(errors.New("error1")),
				slog.Any("error2", errors.New("error2")),
			},
			expectNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test hub
			hub := sentrygo.CurrentHub()
			require.NotNil(t, hub)

			// Create a test record
			record := slog.NewRecord(time.Now(), slog.LevelError, "test message", 0)
			for _, attr := range tt.recordAttrs {
				record.AddAttrs(attr)
			}

			// Call the function
			event := sentry.ConvertSlogToEvent(false, nil, tt.loggerAttr, []string{}, &record, hub)

			// Assert
			if tt.expectNil {
				assert.Nil(t, event, "event should be skipped")
			} else {
				assert.NotNil(t, event, "event should be created")
			}
		})
	}
}

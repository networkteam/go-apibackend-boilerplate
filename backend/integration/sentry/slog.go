package sentry

import (
	"errors"
	"log/slog"

	sentrygo "github.com/getsentry/sentry-go"
	sentryslog "github.com/getsentry/sentry-go/slog"

	"myvendor.mytld/myproject/backend/domain/types"
)

func ConvertSlogToEvent(addSource bool, replaceAttr func(groups []string, a slog.Attr) slog.Attr, loggerAttr []slog.Attr, groups []string, record *slog.Record, hub *sentrygo.Hub) *sentrygo.Event {
	// return nil if the sentinel log attribute is present to skip raising an event in sentry
	if types.HasSentrySkipEventAttr(loggerAttr, record) {
		return nil
	}

	// return nil if error implements ErrorReportSkipper interface
	if hasErrorReportSkipper(loggerAttr, record) {
		return nil
	}

	// return default convertSlogToEvent if sentinel attribute is not present
	return sentryslog.DefaultConverter(addSource, replaceAttr, loggerAttr, groups, record, hub)
}

func hasErrorReportSkipper(loggerAttr []slog.Attr, record *slog.Record) bool {
	hasErrors := false
	allErrorsSkippable := true

	// Check logger attributes for errors
	for _, attr := range loggerAttr {
		if err, ok := attr.Value.Any().(error); ok {
			hasErrors = true
			var skipper types.ErrorReportSkipper
			if !errors.As(err, &skipper) {
				// Found a non-skippable error, must report
				return false
			}
		}
	}

	// Check record attributes for errors
	if record != nil {
		record.Attrs(func(attr slog.Attr) bool {
			if err, ok := attr.Value.Any().(error); ok {
				hasErrors = true
				var skipper types.ErrorReportSkipper
				if !errors.As(err, &skipper) {
					// Found a non-skippable error, must report
					allErrorsSkippable = false
					return false // stop iteration
				}
			}
			return true
		})
	}

	// Only skip if we found errors AND all of them are skippable
	return hasErrors && allErrorsSkippable
}

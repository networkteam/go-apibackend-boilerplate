package types

import (
	"log/slog"
	"slices"
)

func SentrySkipEventAttr() slog.Attr {
	return slog.Bool("sentry-skip-event", true)
}

func HasSentrySkipEventAttr(loggerAttr []slog.Attr, record *slog.Record) bool {
	// check logger attributes for sentinel value
	if slices.ContainsFunc(loggerAttr, func(a slog.Attr) bool {
		return a.Equal(SentrySkipEventAttr())
	}) {
		return true
	}

	// Check record attributes for sentinel value
	foundSentinel := false
	if record != nil {
		record.Attrs(func(a slog.Attr) bool {
			if a.Equal(SentrySkipEventAttr()) {
				foundSentinel = true
				return false // stop iteration
			}
			return true
		})
	}

	return foundSentinel
}

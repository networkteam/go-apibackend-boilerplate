package types

import "errors"

// HandlerLockedError is returned by handlers when they cannot acquire an advisory lock
type HandlerLockedError struct{}

// ErrorReportSkip implements ErrorReportSkipper to indicate this error should be skipped in Sentry reporting
func (e HandlerLockedError) ErrorReportSkip() {}

var _ ErrorReportSkipper = HandlerLockedError{}

func (e HandlerLockedError) Error() string {
	return "handler is locked by another process"
}

func (e HandlerLockedError) Is(target error) bool {
	_, ok := target.(HandlerLockedError)
	return ok
}

// ErrHandlerLocked is a sentinel error for handler lock acquisition failures
var ErrHandlerLocked = HandlerLockedError{}

// IsHandlerLockedError checks if an error is or wraps a HandlerLockedError
func IsHandlerLockedError(err error) bool {
	return errors.Is(err, ErrHandlerLocked)
}

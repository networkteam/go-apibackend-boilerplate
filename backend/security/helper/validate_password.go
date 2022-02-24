package helper

import "errors"

const MinPasswordLength = 8

var ErrPasswordTooShort = errors.New("tooShort")

func ValidatePassword(password string) error {
	if len(password) < MinPasswordLength {
		return ErrPasswordTooShort
	}

	return nil
}

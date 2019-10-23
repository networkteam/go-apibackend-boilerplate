package domain

import "errors"

const MinPasswordLength = 8

var ErrPasswordTooShort = errors.New("password too short")

func ValidatePassword(password []byte) error {
	if len(password) < MinPasswordLength {
		return ErrPasswordTooShort
	}

	return nil
}

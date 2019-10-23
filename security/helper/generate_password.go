package helper

import (
	"github.com/sethvargo/go-password/password"
)

func GeneratePassword(length int) (string, error) {
	pwd, err := password.Generate(length, 10, 10, false, false)
	if err != nil {
		return "", err
	}

	return pwd, nil
}

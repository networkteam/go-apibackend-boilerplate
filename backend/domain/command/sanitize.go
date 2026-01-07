package command

import "strings"

func SanitizeEmailAddress(emailAddress string) string {
	emailAddress = strings.TrimSpace(emailAddress)
	emailAddress = strings.ToLower(emailAddress)
	return emailAddress
}

func SanitizeUsername(username string) string {
	username = strings.TrimSpace(username)
	username = strings.ToLower(username)
	return username
}

func SanitizePassword(password string) string {
	return strings.TrimSpace(password)
}

package domain

import (
	"strings"
	"time"
)

// Bcrypt hash cost (defaults to 10), but 12 is more secure
const defaultHashCost = 12

// Config holds the base configuration used by various parts of the application
type Config struct {
	AppName string
	// Base URL of the app (for sending mails and linking back)
	AppBaseUrl string
	// Bcrypt hash cost for passwords
	HashCost int
	// Location for date / time calculations, defaults to Europe/Berlin
	Location *time.Location
}

func DefaultConfig() Config {
	location, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		panic(err)
	}
	return Config{
		AppName:  "myproject",
		HashCost: defaultHashCost,
		Location: location,
	}
}
func (c Config) BuildURL(path string) string {
	// Strip slashes to prevent mixup of double slashes or no slashes at all
	baseURL := strings.TrimRight(c.AppBaseUrl, "/")
	p := strings.TrimLeft(path, "/")
	return baseURL + "/" + p
}

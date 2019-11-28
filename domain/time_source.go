package domain

import "time"

type TimeSource interface {
	Now() time.Time
}

package types

import "time"

type TimeSource interface {
	Now() time.Time
}

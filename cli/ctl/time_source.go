package main

import (
	"time"

	"myvendor/myproject/backend/domain"
)

type currentTimeSource struct{}

func (cts currentTimeSource) Now() time.Time {
	return time.Now()
}

func newCurrentTimeSource() domain.TimeSource {
	return currentTimeSource{}
}

package main

import (
	"time"

	"myvendor.mytld/myproject/backend/domain"
)

type currentTimeSource struct{}

func (cts currentTimeSource) Now() time.Time {
	return time.Now()
}

func newCurrentTimeSource() domain.TimeSource {
	return currentTimeSource{}
}

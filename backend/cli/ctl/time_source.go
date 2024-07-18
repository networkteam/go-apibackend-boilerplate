package main

import (
	"time"

	"github.com/urfave/cli/v2"

	"myvendor.mytld/myproject/backend/domain"
)

type currentTimeSource struct{}

func (cts currentTimeSource) Now() time.Time {
	return time.Now()
}

func newCurrentTimeSource(_ *cli.Context) (domain.TimeSource, error) {
	// TODO Get location from CLI context and store in time source

	return currentTimeSource{}, nil
}

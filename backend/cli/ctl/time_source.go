package main

import (
	"time"

	"github.com/urfave/cli/v2"

	"myvendor.mytld/myproject/backend/domain/types"
)

type currentTimeSource struct{}

func (cts currentTimeSource) Now() time.Time {
	return time.Now()
}

func newCurrentTimeSource(_ *cli.Context) (types.TimeSource, error) {
	// TODO Get location from CLI context and store in time source

	return currentTimeSource{}, nil
}

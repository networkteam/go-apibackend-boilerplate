package main

import (
	"time"

	"github.com/networkteam/slogutils"
	"github.com/urfave/cli/v2"

	"myvendor.mytld/myproject/backend/domain/types"
)

type currentTimeSource struct{}

func (cts currentTimeSource) Now() time.Time {
	return time.Now()
}

type timeSourceWithDelta struct {
	delta types.DateDuration
}

func (ts timeSourceWithDelta) Now() time.Time {
	return ts.delta.Apply(time.Now())
}

func newCurrentTimeSource(c *cli.Context) (types.TimeSource, error) {
	logger := slogutils.FromContext(c.Context)

	deltaStr := c.String("test-time-delta")
	if deltaStr == "" {
		return currentTimeSource{}, nil
	}

	delta, err := types.ParseDateDuration(deltaStr)
	if err != nil {
		return nil, err
	}

	t := timeSourceWithDelta{delta: delta}

	if !delta.IsZero() {
		logger.Warn("Test time delta is active", "delta", delta.String(), "effectiveNow", t.Now())
	}

	return t, nil
}

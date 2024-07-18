package test

import (
	"time"
)

// FixedTimeSource is a constant time for Now()
type FixedTimeSource time.Time

// Now returns the fixed time
func (fts FixedTimeSource) Now() time.Time {
	return (time.Time)(fts)
}

// Add the given duration and return a _new_ time source
func (fts FixedTimeSource) Add(d time.Duration) FixedTimeSource {
	t := (time.Time)(fts)
	return FixedTimeSource(t.Add(d))
}

// AddDate adds the given year, month and day to the fixed time
func (fts FixedTimeSource) AddDate(y int, m int, d int) FixedTimeSource {
	t := (time.Time)(fts)
	return FixedTimeSource(t.AddDate(y, m, d))
}

func MustFixedTimeSource(isoTime string) FixedTimeSource {
	t, err := time.Parse(time.RFC3339, isoTime)
	if err != nil {
		panic(err)
	}
	return FixedTimeSource(t)
}

// FixedTime returns a fixed time for testing
func FixedTime() FixedTimeSource {
	return MustFixedTimeSource("2020-09-23T08:34:57.321Z")
}

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

func MustFixedTimeSource(isoTime string) FixedTimeSource {
	t, err := time.Parse(time.RFC3339, isoTime)
	if err != nil {
		panic(err)
	}
	return FixedTimeSource(t)
}

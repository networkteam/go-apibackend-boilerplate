package helper

import (
	"myvendor/myproject/backend/test"
)

// FixedTime returns a fixed time for testing
func FixedTime() test.FixedTimeSource {
	return test.MustFixedTimeSource("2018-12-17T14:42:57.321Z")
}

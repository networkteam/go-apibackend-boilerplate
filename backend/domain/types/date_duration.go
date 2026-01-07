package types

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	fog_errors "github.com/friendsofgo/errors"
)

// DateDuration represents a duration that can span years, months, days, and hours.
// This is useful for time deltas that need calendar-aware arithmetic (AddDate).
type DateDuration struct {
	Years  int
	Months int
	Days   int
	Hours  time.Duration
}

// ErrInvalidDateDuration is returned when a date duration string cannot be parsed.
var ErrInvalidDateDuration = errors.New("invalid date duration")

var dateDurationRegex = regexp.MustCompile(`^(-?\d+)([yMdh])`)

// ParseDateDuration parses a duration string in the format "1y2M3d4h".
// Supported units: y (years), M (months), d (days), h (hours).
// Values can be negative (e.g., "-1y", "2M-3d").
// Examples:
//   - "1y" → 1 year forward
//   - "-2M" → 2 months back
//   - "30d" → 30 days forward
//   - "1y6M15d" → 1 year, 6 months, 15 days forward
//   - "2h" → 2 hours forward
//   - "-1y2M" → 1 year 2 months back (mixed signs not recommended but supported)
func ParseDateDuration(value string) (DateDuration, error) {
	if value == "" {
		return DateDuration{}, nil
	}

	var dateDuration DateDuration
	remaining := value

	for remaining != "" {
		matches := dateDurationRegex.FindStringSubmatch(remaining)
		if matches == nil {
			return DateDuration{}, fog_errors.Wrapf(ErrInvalidDateDuration, "invalid format: %q (expected format like '1y2M3d4h')", value)
		}

		valueStr := matches[1]
		unit := matches[2]

		value, err := strconv.Atoi(valueStr)
		if err != nil {
			return DateDuration{}, fog_errors.Wrapf(ErrInvalidDateDuration, "invalid number: %q", valueStr)
		}

		switch unit {
		case "y":
			dateDuration.Years += value
		case "M":
			dateDuration.Months += value
		case "d":
			dateDuration.Days += value
		case "h":
			dateDuration.Hours += time.Duration(value) * time.Hour
		default:
			return DateDuration{}, fog_errors.Wrapf(ErrInvalidDateDuration, "unknown unit: %q (supported: y, M, d, h)", unit)
		}

		remaining = remaining[len(matches[0]):]
	}

	return dateDuration, nil
}

// Apply applies this duration to the given time and returns the result.
func (dd DateDuration) Apply(t time.Time) time.Time {
	return t.AddDate(dd.Years, dd.Months, dd.Days).Add(dd.Hours)
}

// IsZero returns true if this duration is zero (no offset).
func (dd DateDuration) IsZero() bool {
	return dd.Years == 0 && dd.Months == 0 && dd.Days == 0 && dd.Hours == 0
}

// String formats the duration back to a string representation.
func (dd DateDuration) String() string {
	if dd.IsZero() {
		return "0"
	}

	var parts []string
	if dd.Years != 0 {
		parts = append(parts, fmt.Sprintf("%dy", dd.Years))
	}
	if dd.Months != 0 {
		parts = append(parts, fmt.Sprintf("%dM", dd.Months))
	}
	if dd.Days != 0 {
		parts = append(parts, fmt.Sprintf("%dd", dd.Days))
	}
	if dd.Hours != 0 {
		hours := int(dd.Hours.Hours())
		parts = append(parts, fmt.Sprintf("%dh", hours))
	}

	return strings.Join(parts, "")
}

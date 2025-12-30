package types_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"myvendor.mytld/myproject/backend/domain/types"
)

func TestParseDateDuration(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    types.DateDuration
		wantErr bool
	}{
		{
			name:  "empty string returns zero duration",
			input: "",
			want:  types.DateDuration{},
		},
		{
			name:  "1 year forward",
			input: "1y",
			want:  types.DateDuration{Years: 1},
		},
		{
			name:  "2 months forward",
			input: "2M",
			want:  types.DateDuration{Months: 2},
		},
		{
			name:  "30 days forward",
			input: "30d",
			want:  types.DateDuration{Days: 30},
		},
		{
			name:  "4 hours forward",
			input: "4h",
			want:  types.DateDuration{Hours: 4 * time.Hour},
		},
		{
			name:  "combined: 1y2M3d4h",
			input: "1y2M3d4h",
			want: types.DateDuration{
				Years:  1,
				Months: 2,
				Days:   3,
				Hours:  4 * time.Hour,
			},
		},
		{
			name:  "negative 1 year",
			input: "-1y",
			want:  types.DateDuration{Years: -1},
		},
		{
			name:  "negative 6 months",
			input: "-6M",
			want:  types.DateDuration{Months: -6},
		},
		{
			name:  "negative 15 days",
			input: "-15d",
			want:  types.DateDuration{Days: -15},
		},
		{
			name:  "mixed signs: 1y-2M",
			input: "1y-2M",
			want: types.DateDuration{
				Years:  1,
				Months: -2,
			},
		},
		{
			name:  "mixed signs: -1y2M3d",
			input: "-1y2M3d",
			want: types.DateDuration{
				Years:  -1,
				Months: 2,
				Days:   3,
			},
		},
		{
			name:  "zero values: 0y0M0d",
			input: "0y0M0d",
			want:  types.DateDuration{},
		},
		{
			name:    "invalid format: abc",
			input:   "abc",
			wantErr: true,
		},
		{
			name:    "invalid unit: 1x",
			input:   "1x",
			wantErr: true,
		},
		{
			name:    "duplicate unit: 1y2y",
			input:   "1y2y",
			want:    types.DateDuration{Years: 3}, // Both values are added
			wantErr: false,
		},
		{
			name:    "no number: y",
			input:   "y",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := types.ParseDateDuration(tt.input)
			if tt.wantErr {
				assert.Error(t, err, "expected error for input")
			} else {
				assert.NoError(t, err, "should parse without error")
				assert.Equal(t, tt.want, got, "parsed duration should match expected")
			}
		})
	}
}

func TestDateDuration_Apply(t *testing.T) {
	baseTime := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		duration types.DateDuration
		want     time.Time
	}{
		{
			name:     "zero duration",
			duration: types.DateDuration{},
			want:     baseTime,
		},
		{
			name:     "1 year forward",
			duration: types.DateDuration{Years: 1},
			want:     time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC),
		},
		{
			name:     "2 months forward",
			duration: types.DateDuration{Months: 2},
			want:     time.Date(2025, 3, 15, 12, 0, 0, 0, time.UTC),
		},
		{
			name:     "30 days forward",
			duration: types.DateDuration{Days: 30},
			want:     time.Date(2025, 2, 14, 12, 0, 0, 0, time.UTC),
		},
		{
			name:     "4 hours forward",
			duration: types.DateDuration{Hours: 4 * time.Hour},
			want:     time.Date(2025, 1, 15, 16, 0, 0, 0, time.UTC),
		},
		{
			name: "combined: 1y2M3d4h",
			duration: types.DateDuration{
				Years:  1,
				Months: 2,
				Days:   3,
				Hours:  4 * time.Hour,
			},
			want: time.Date(2026, 3, 18, 16, 0, 0, 0, time.UTC),
		},
		{
			name:     "negative 1 year",
			duration: types.DateDuration{Years: -1},
			want:     time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
		},
		{
			name:     "negative 6 months",
			duration: types.DateDuration{Months: -6},
			want:     time.Date(2024, 7, 15, 12, 0, 0, 0, time.UTC),
		},
		{
			name:     "negative 15 days",
			duration: types.DateDuration{Days: -15},
			want:     time.Date(2024, 12, 31, 12, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.duration.Apply(baseTime)
			assert.Equal(t, tt.want, got, "applied time should match expected")
		})
	}
}

func TestDateDuration_IsZero(t *testing.T) {
	tests := []struct {
		name     string
		duration types.DateDuration
		want     bool
	}{
		{
			name:     "zero duration",
			duration: types.DateDuration{},
			want:     true,
		},
		{
			name:     "explicit zeros",
			duration: types.DateDuration{Years: 0, Months: 0, Days: 0, Hours: 0},
			want:     true,
		},
		{
			name:     "1 year",
			duration: types.DateDuration{Years: 1},
			want:     false,
		},
		{
			name:     "1 month",
			duration: types.DateDuration{Months: 1},
			want:     false,
		},
		{
			name:     "1 day",
			duration: types.DateDuration{Days: 1},
			want:     false,
		},
		{
			name:     "1 hour",
			duration: types.DateDuration{Hours: time.Hour},
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.duration.IsZero()
			assert.Equal(t, tt.want, got, "IsZero result should match expected")
		})
	}
}

func TestDateDuration_String(t *testing.T) {
	tests := []struct {
		name     string
		duration types.DateDuration
		want     string
	}{
		{
			name:     "zero duration",
			duration: types.DateDuration{},
			want:     "0",
		},
		{
			name:     "1 year",
			duration: types.DateDuration{Years: 1},
			want:     "1y",
		},
		{
			name:     "2 months",
			duration: types.DateDuration{Months: 2},
			want:     "2M",
		},
		{
			name:     "30 days",
			duration: types.DateDuration{Days: 30},
			want:     "30d",
		},
		{
			name:     "4 hours",
			duration: types.DateDuration{Hours: 4 * time.Hour},
			want:     "4h",
		},
		{
			name: "combined: 1y2M3d4h",
			duration: types.DateDuration{
				Years:  1,
				Months: 2,
				Days:   3,
				Hours:  4 * time.Hour,
			},
			want: "1y2M3d4h",
		},
		{
			name:     "negative 1 year",
			duration: types.DateDuration{Years: -1},
			want:     "-1y",
		},
		{
			name:     "negative 6 months",
			duration: types.DateDuration{Months: -6},
			want:     "-6M",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.duration.String()
			assert.Equal(t, tt.want, got, "string representation should match expected")
		})
	}
}

func TestParseDateDuration_RoundTrip(t *testing.T) {
	tests := []string{
		"1y",
		"2M",
		"30d",
		"4h",
		"1y2M3d4h",
		"-1y",
		"-6M",
		"-15d",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			parsed, err := types.ParseDateDuration(input)
			assert.NoError(t, err, "should parse without error")

			stringified := parsed.String()
			reparsed, err := types.ParseDateDuration(stringified)
			assert.NoError(t, err, "should reparse without error")

			assert.Equal(t, parsed, reparsed, "round trip should preserve duration")
		})
	}
}

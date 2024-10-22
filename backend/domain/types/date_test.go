package types_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"myvendor.mytld/myproject/backend/domain/types"
)

func TestDates(t *testing.T) {
	for _, test := range []struct {
		date     types.Date
		loc      *time.Location
		wantStr  string
		wantTime time.Time
	}{
		{
			date:     types.Date{2014, 7, 29},
			loc:      time.Local,
			wantStr:  "2014-07-29",
			wantTime: time.Date(2014, time.July, 29, 0, 0, 0, 0, time.Local),
		},
		{
			date:     types.DateOf(time.Date(2014, 8, 20, 15, 8, 43, 1, time.Local)),
			loc:      time.UTC,
			wantStr:  "2014-08-20",
			wantTime: time.Date(2014, 8, 20, 0, 0, 0, 0, time.UTC),
		},
		{
			date:     types.DateOf(time.Date(999, time.January, 26, 0, 0, 0, 0, time.Local)),
			loc:      time.UTC,
			wantStr:  "0999-01-26",
			wantTime: time.Date(999, 1, 26, 0, 0, 0, 0, time.UTC),
		},
	} {
		if got := test.date.String(); got != test.wantStr {
			t.Errorf("%#v.String() = %q, want %q", test.date, got, test.wantStr)
		}
		if got := test.date.In(test.loc); !got.Equal(test.wantTime) {
			t.Errorf("%#v.In(%v) = %v, want %v", test.date, test.loc, got, test.wantTime)
		}
	}
}

func TestDateIsValid(t *testing.T) {
	for _, test := range []struct {
		date types.Date
		want bool
	}{
		{types.Date{2014, 7, 29}, true},
		{types.Date{2000, 2, 29}, true},
		{types.Date{10000, 12, 31}, true},
		{types.Date{1, 1, 1}, true},
		{types.Date{0, 1, 1}, true},  // year zero is OK
		{types.Date{-1, 1, 1}, true}, // negative year is OK
		{types.Date{1, 0, 1}, false},
		{types.Date{1, 1, 0}, false},
		{types.Date{2016, 1, 32}, false},
		{types.Date{2016, 13, 1}, false},
		{types.Date{1, -1, 1}, false},
		{types.Date{1, 1, -1}, false},
	} {
		got := test.date.IsValid()
		if got != test.want {
			t.Errorf("%#v: got %t, want %t", test.date, got, test.want)
		}
	}
}

func TestParseDate(t *testing.T) {
	for _, test := range []struct {
		str  string
		want types.Date // if empty, expect an error
	}{
		{"2016-01-02", types.Date{2016, 1, 2}},
		{"2016-12-31", types.Date{2016, 12, 31}},
		{"0003-02-04", types.Date{3, 2, 4}},
		{"999-01-26", types.Date{}},
		{"", types.Date{}},
		{"2016-01-02x", types.Date{}},
	} {
		got, err := types.ParseDate(test.str)
		if got != test.want {
			t.Errorf("ParseDate(%q) = %+v, want %+v", test.str, got, test.want)
		}
		if err != nil && test.want != (types.Date{}) {
			t.Errorf("Unexpected error %v from ParseDate(%q)", err, test.str)
		}
	}
}

func TestDateArithmetic(t *testing.T) {
	for _, test := range []struct {
		desc  string
		start types.Date
		end   types.Date
		days  int
	}{
		{
			desc:  "zero days noop",
			start: types.Date{2014, 5, 9},
			end:   types.Date{2014, 5, 9},
			days:  0,
		},
		{
			desc:  "crossing a year boundary",
			start: types.Date{2014, 12, 31},
			end:   types.Date{2015, 1, 1},
			days:  1,
		},
		{
			desc:  "negative number of days",
			start: types.Date{2015, 1, 1},
			end:   types.Date{2014, 12, 31},
			days:  -1,
		},
		{
			desc:  "full leap year",
			start: types.Date{2004, 1, 1},
			end:   types.Date{2005, 1, 1},
			days:  366,
		},
		{
			desc:  "full non-leap year",
			start: types.Date{2001, 1, 1},
			end:   types.Date{2002, 1, 1},
			days:  365,
		},
		{
			desc:  "crossing a leap second",
			start: types.Date{1972, 6, 30},
			end:   types.Date{1972, 7, 1},
			days:  1,
		},
		{
			desc:  "dates before the unix epoch",
			start: types.Date{101, 1, 1},
			end:   types.Date{102, 1, 1},
			days:  365,
		},
	} {
		if got := test.start.AddDays(test.days); got != test.end {
			t.Errorf("[%s] %#v.AddDays(%v) = %#v, want %#v", test.desc, test.start, test.days, got, test.end)
		}
		if got := test.end.DaysSince(test.start); got != test.days {
			t.Errorf("[%s] %#v.Sub(%#v) = %v, want %v", test.desc, test.end, test.start, got, test.days)
		}
	}
}

func TestAddMonths(t *testing.T) {
	for _, test := range []struct {
		desc   string
		start  types.Date
		end    types.Date
		months int
	}{
		{
			desc:   "zero months noop",
			start:  types.Date{2014, 5, 9},
			end:    types.Date{2014, 5, 9},
			months: 0,
		},
		{
			desc:   "crossing a year boundary",
			start:  types.Date{2014, 12, 31},
			end:    types.Date{2015, 1, 31},
			months: 1,
		},
		{
			desc:   "keeps amount of days",
			start:  types.Date{2015, 1, 31},
			end:    types.Date{2015, 3, 3},
			months: 1,
		},
		{
			desc:   "negative number of months",
			start:  types.Date{2015, 1, 1},
			end:    types.Date{2014, 12, 1},
			months: -1,
		},
		{
			desc:   "full year",
			start:  types.Date{2004, 1, 1},
			end:    types.Date{2005, 1, 1},
			months: 12,
		},
	} {
		if got := test.start.AddMonths(test.months); got != test.end {
			t.Errorf("[%s] %#v.AddMonths(%v) = %#v, want %#v", test.desc, test.start, test.months, got, test.end)
		}
	}
}

func TestDateBefore(t *testing.T) {
	for _, test := range []struct {
		d1, d2 types.Date
		want   bool
	}{
		{types.Date{2016, 12, 31}, types.Date{2017, 1, 1}, true},
		{types.Date{2016, 1, 1}, types.Date{2016, 1, 1}, false},
		{types.Date{2016, 12, 30}, types.Date{2016, 12, 31}, true},
	} {
		if got := test.d1.Before(test.d2); got != test.want {
			t.Errorf("%v.Before(%v): got %t, want %t", test.d1, test.d2, got, test.want)
		}
	}
}

func TestDateAfter(t *testing.T) {
	for _, test := range []struct {
		d1, d2 types.Date
		want   bool
	}{
		{types.Date{2016, 12, 31}, types.Date{2017, 1, 1}, false},
		{types.Date{2016, 1, 1}, types.Date{2016, 1, 1}, false},
		{types.Date{2016, 12, 30}, types.Date{2016, 12, 31}, false},
	} {
		if got := test.d1.After(test.d2); got != test.want {
			t.Errorf("%v.After(%v): got %t, want %t", test.d1, test.d2, got, test.want)
		}
	}
}

func TestMarshalJSON(t *testing.T) {
	for _, test := range []struct {
		value interface{}
		want  string
	}{
		{types.Date{1987, 4, 15}, `"1987-04-15"`},
	} {
		bgot, err := json.Marshal(test.value)
		if err != nil {
			t.Fatal(err)
		}
		if got := string(bgot); got != test.want {
			t.Errorf("%#v: got %s, want %s", test.value, got, test.want)
		}
	}
}

func TestUnmarshalJSON(t *testing.T) {
	var d types.Date
	for _, test := range []struct {
		data string
		ptr  interface{}
		want interface{}
	}{
		{`"1987-04-15"`, &d, &types.Date{1987, 4, 15}},
		{`"1987-04-\u0031\u0035"`, &d, &types.Date{1987, 4, 15}},
	} {
		if err := json.Unmarshal([]byte(test.data), test.ptr); err != nil {
			t.Fatalf("%s: %v", test.data, err)
		}
		assert.Equal(t, test.want, test.ptr)
	}

	for _, bad := range []string{"", `""`, `"bad"`, `"1987-04-15x"`,
		`19870415`,     // a JSON number
		`11987-04-15x`, // not a JSON string

	} {
		if json.Unmarshal([]byte(bad), &d) == nil {
			t.Errorf("%q, Date: got nil, want error", bad)
		}
	}
}

func TestDateAt(t *testing.T) {
	d := types.Date{
		Year:  2022,
		Month: 1,
		Day:   28,
	}
	result := d.At(11, 1, 42, time.UTC)

	assert.Equal(t, "2022-01-28T11:01:42Z", result.Format(time.RFC3339))
}

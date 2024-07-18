package domain

// Copied from cloud.google.com/go/civil

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// A Date represents a date (year, month, day).
//
// This type does not include location information, and therefore does not
// describe a unique 24-hour timespan.
type Date struct {
	Year  int        // Year (e.g., 2014).
	Month time.Month // Month of the year (January = 1, ...).
	Day   int        // Day of the month, starting at 1.
}

var _ sql.Scanner = &Date{}
var _ driver.Valuer = Date{}

// DateOf returns the Date in which a time occurs in that time's location.
func DateOf(t time.Time) Date {
	var d Date
	d.Year, d.Month, d.Day = t.Date()
	return d
}

// ParseDate parses a string in RFC3339 full-date format and returns the date value it represents.
func ParseDate(s string) (Date, error) {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return Date{}, err
	}
	return DateOf(t), nil
}

// String returns the date in RFC3339 full-date format.
func (d Date) String() string {
	return fmt.Sprintf("%04d-%02d-%02d", d.Year, d.Month, d.Day)
}

// IsValid reports whether the date is valid.
func (d Date) IsValid() bool {
	return DateOf(d.In(time.UTC)) == d
}

// In returns the time corresponding to time 00:00:00 of the date in the location.
//
// In is always consistent with time.Date, even when time.Date returns a time
// on a different day. For example, if loc is America/Indiana/Vincennes, then both
//
//	time.Date(1955, time.May, 1, 0, 0, 0, 0, loc)
//
// and
//
//	civil.Date{Year: 1955, Month: time.May, Day: 1}.In(loc)
//
// return 23:00:00 on April 30, 1955.
//
// In panics if loc is nil.
func (d Date) In(loc *time.Location) time.Time {
	return time.Date(d.Year, d.Month, d.Day, 0, 0, 0, 0, loc)
}

func (d Date) At(hours int, minutes int, seconds int, loc *time.Location) time.Time {
	return time.Date(d.Year, d.Month, d.Day, hours, minutes, seconds, 0, loc)
}

// AddDays returns the date that is n days in the future.
// n can also be negative to go into the past.
func (d Date) AddDays(n int) Date {
	return DateOf(d.In(time.UTC).AddDate(0, 0, n))
}

// DaysSince returns the signed number of days between the date and s, not including the end day.
// This is the inverse operation to AddDays.
func (d Date) DaysSince(s Date) (days int) {
	// We convert to Unix time so we do not have to worry about leap seconds:
	// Unix time increases by exactly 86400 seconds per day.
	deltaUnix := d.In(time.UTC).Unix() - s.In(time.UTC).Unix()
	return int(deltaUnix / 86400)
}

// AddMonths returns a date with the amount of months (positive or negative) added / subtracted to the date.
// It uses time.Time AddDate semantics, so the day of month could change when the number of days in months differ.
func (d Date) AddMonths(months int) Date {
	return DateOf(d.In(time.UTC).AddDate(0, months, 0))
}

// Before reports whether d occurs before other.
func (d Date) Before(other Date) bool {
	if d.Year != other.Year {
		return d.Year < other.Year
	}
	if d.Month != other.Month {
		return d.Month < other.Month
	}
	return d.Day < other.Day
}

// BeforeOrEqual reports whether d occurs before other or is the same.
func (d Date) BeforeOrEqual(other Date) bool {
	return d == other || d.Before(other)
}

// After reports whether d occurs after other.
func (d Date) After(other Date) bool {
	return other.Before(d)
}

// AfterOrEqual reports whether d occurs after other or is the same.
func (d Date) AfterOrEqual(other Date) bool {
	return d.After(other) || d == other
}

// MarshalText implements the encoding.TextMarshaler interface.
// The output is the result of d.String().
func (d Date) MarshalText() ([]byte, error) {
	return []byte(d.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
// The date is expected to be a string in a format accepted by ParseDate.
func (d *Date) UnmarshalText(data []byte) error {
	var err error
	*d, err = ParseDate(string(data))
	return err
}

// IsZero checks if the date is the zero value
func (d Date) IsZero() bool {
	return d.Day == 0 && d.Month == 0 && d.Year == 0
}

// Weekday returns the day of the week of the date
func (d Date) Weekday() time.Weekday {
	return d.In(time.UTC).Weekday()
}

func (d Date) Value() (driver.Value, error) {
	return d.String(), nil
}

func (d *Date) Scan(src interface{}) error {
	switch value := src.(type) {
	case string:
		parsedDate, err := ParseDate(value)
		if err != nil {
			return err
		}
		*d = parsedDate
	case time.Time:
		*d = DateOf(value)
	default:
		//nolint:goerr113
		return fmt.Errorf("unhandled type: %T", src)
	}

	return nil
}

// NullDate can be used with the standard sql package to represent a
// Date value that can be NULL in the database.
type NullDate struct {
	Date  Date
	Valid bool
}

// Value implements the driver.Valuer interface.
func (u NullDate) Value() (driver.Value, error) {
	if !u.Valid {
		return nil, nil
	}
	// Delegate to Date Value function
	return u.Date.Value()
}

// Scan implements the sql.Scanner interface.
func (u *NullDate) Scan(src interface{}) error {
	if src == nil {
		u.Date, u.Valid = Date{}, false
		return nil
	}

	// Delegate to Date Scan function
	u.Valid = true
	return u.Date.Scan(src)
}

// MarshalJSON marshals the NullDate as null or the nested Date
func (u NullDate) MarshalJSON() ([]byte, error) {
	if !u.Valid {
		return json.Marshal(nil)
	}

	return json.Marshal(u.Date)
}

// UnmarshalJSON unmarshals a NullDate
func (u *NullDate) UnmarshalJSON(b []byte) error {
	if bytes.Equal(b, []byte("null")) {
		u.Date, u.Valid = Date{}, false
		return nil
	}

	if err := json.Unmarshal(b, &u.Date); err != nil {
		return err
	}

	u.Valid = true

	return nil
}

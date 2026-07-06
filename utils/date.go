package utils

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

// Date is a time.Time wrapper that accepts multiple date/datetime formats
// from JSON (YYYY-MM-DD, YYYY-MM-DDTHH:MM:SS, RFC3339) and always stores as time.Time.
// Solves Gin binding error: cannot parse "2026-06-28" as "T" in RFC3339.
type Date struct {
	time.Time
}

// accepted input formats — tried in order
var dateFormats = []string{
	"2006-01-02",
	"2006-01-02T15:04:05Z07:00",
	"2006-01-02T15:04:05",
	"2006-01-02 15:04:05",
	time.RFC3339,
}

func (d *Date) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "" || s == "null" {
		d.Time = time.Time{}
		return nil
	}
	for _, layout := range dateFormats {
		if t, err := time.Parse(layout, s); err == nil {
			d.Time = t
			return nil
		}
	}
	return fmt.Errorf("cannot parse date %q: expected YYYY-MM-DD or RFC3339", s)
}

func (d Date) MarshalJSON() ([]byte, error) {
	if d.Time.IsZero() {
		return []byte("null"), nil
	}
	return []byte(`"` + d.Time.Format("2006-01-02") + `"`), nil
}

// Value implements driver.Valuer so GORM can write this to MySQL.
func (d Date) Value() (driver.Value, error) {
	if d.Time.IsZero() {
		return nil, nil
	}
	return d.Time, nil
}

// Scan implements sql.Scanner so GORM can read this from MySQL.
func (d *Date) Scan(value interface{}) error {
	if value == nil {
		d.Time = time.Time{}
		return nil
	}
	if t, ok := value.(time.Time); ok {
		d.Time = t
		return nil
	}
	return fmt.Errorf("cannot scan %T into Date", value)
}

// ToTimePtr returns a *time.Time pointer (nil if zero).
func (d *Date) ToTimePtr() *time.Time {
	if d == nil || d.Time.IsZero() {
		return nil
	}
	t := d.Time
	return &t
}

// DateFromPtr creates a *Date from a *time.Time.
func DateFromPtr(t *time.Time) *Date {
	if t == nil {
		return nil
	}
	return &Date{Time: *t}
}

// DateTime wraps time.Time and accepts both date-only and full datetime strings.
type DateTime struct {
	time.Time
}

var dateTimeFormats = []string{
	"2006-01-02T15:04",
	"2006-01-02T15:04:05Z07:00",
	"2006-01-02T15:04:05",
	"2006-01-02 15:04:05",
	"2006-01-02",
	time.RFC3339,
}

func (dt *DateTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "" || s == "null" {
		dt.Time = time.Time{}
		return nil
	}
	for _, layout := range dateTimeFormats {
		if t, err := time.Parse(layout, s); err == nil {
			dt.Time = t
			return nil
		}
	}
	return fmt.Errorf("cannot parse datetime %q: expected YYYY-MM-DD or YYYY-MM-DDTHH:MM", s)
}

func (dt DateTime) MarshalJSON() ([]byte, error) {
	if dt.Time.IsZero() {
		return []byte("null"), nil
	}
	return []byte(`"` + dt.Time.Format(time.RFC3339) + `"`), nil
}

func (dt DateTime) Value() (driver.Value, error) {
	if dt.Time.IsZero() {
		return nil, nil
	}
	return dt.Time, nil
}

func (dt *DateTime) Scan(value interface{}) error {
	if value == nil {
		dt.Time = time.Time{}
		return nil
	}
	if t, ok := value.(time.Time); ok {
		dt.Time = t
		return nil
	}
	return fmt.Errorf("cannot scan %T into DateTime", value)
}

func (dt *DateTime) ToTimePtr() *time.Time {
	if dt == nil || dt.Time.IsZero() {
		return nil
	}
	t := dt.Time
	return &t
}

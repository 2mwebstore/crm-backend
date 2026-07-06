package utils

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

// FlexTime accepts multiple date/datetime formats from JSON input:
//
//	"2026-06-28"                  → date only (time set to 00:00:00 UTC)
//	"2026-06-28T14:30"            → datetime without seconds
//	"2026-06-28T14:30:00"         → datetime without timezone
//	"2026-06-28T14:30:00Z"        → RFC3339
//	"2026-06-28T14:30:00+07:00"   → RFC3339 with offset
//	"2026-06-28 14:30:00"         → space-separated datetime
//
// Serializes back to RFC3339 in JSON responses.
// Implements driver.Valuer and sql.Scanner for GORM.
type FlexTime struct {
	time.Time
}

var flexFormats = []string{
	"2006-01-02",
	"2006-01-02T15:04",
	"2006-01-02T15:04:05",
	"2006-01-02T15:04:05Z07:00",
	"2006-01-02T15:04:05Z",
	"2006-01-02 15:04:05",
	time.RFC3339,
	time.RFC3339Nano,
}

func (ft *FlexTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)

	if s == "" || s == "null" {
		ft.Time = time.Time{}
		return nil
	}

	loc, err := time.LoadLocation("Asia/Phnom_Penh")
	if err != nil {
		loc = time.Local
	}

	for _, layout := range flexFormats {
		var t time.Time

		// Layouts that already include timezone information
		if strings.Contains(layout, "Z07:00") ||
			layout == time.RFC3339 ||
			layout == time.RFC3339Nano {

			t, err = time.Parse(layout, s)
		} else {
			// Parse as Cambodia local time
			t, err = time.ParseInLocation(layout, s, loc)
		}

		if err == nil {
			ft.Time = t
			return nil
		}
	}

	return fmt.Errorf("cannot parse %q as a date/datetime", s)
}
func (ft FlexTime) MarshalJSON() ([]byte, error) {
	if ft.Time.IsZero() {
		return []byte("null"), nil
	}
	return []byte(`"` + ft.Time.Format(time.RFC3339) + `"`), nil
}

func (ft FlexTime) Value() (driver.Value, error) {
	if ft.Time.IsZero() {
		return nil, nil
	}
	return ft.Time, nil
}

func (ft *FlexTime) Scan(value interface{}) error {
	if value == nil {
		ft.Time = time.Time{}
		return nil
	}
	if t, ok := value.(time.Time); ok {
		ft.Time = t
		return nil
	}
	return fmt.Errorf("cannot scan %T into FlexTime", value)
}

// ToTimePtr returns a *time.Time (nil if zero).
func (ft *FlexTime) ToTimePtr() *time.Time {
	if ft == nil || ft.Time.IsZero() {
		return nil
	}
	t := ft.Time
	return &t
}

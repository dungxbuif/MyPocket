package entity

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

const calendarDateLayout = "2006-01-02"

// CalendarDate represents a date label without a timezone or time of day.
type CalendarDate struct{ time.Time }

func ParseCalendarDate(value string) (CalendarDate, error) {
	parsed, err := time.Parse(calendarDateLayout, value)
	if err != nil || parsed.Format(calendarDateLayout) != value {
		return CalendarDate{}, fmt.Errorf("invalid calendar date %q", value)
	}
	return CalendarDate{Time: parsed}, nil
}

func ParseMonth(value string) (time.Time, error) {
	month, err := time.Parse("2006-01", value)
	if err != nil || month.Format("2006-01") != value {
		return time.Time{}, fmt.Errorf("invalid month %q", value)
	}
	return month.UTC(), nil
}

func MonthRangeUTC(month time.Time, location *time.Location) (time.Time, time.Time, error) {
	if location == nil {
		return time.Time{}, time.Time{}, fmt.Errorf("account timezone is required")
	}
	year, number, _ := month.UTC().Date()
	start := time.Date(year, number, 1, 0, 0, 0, 0, location)
	next := time.Date(year, number+1, 1, 0, 0, 0, 0, location)
	return start.UTC(), next.UTC(), nil
}

func IsMonthComplete(month time.Time, now time.Time, location *time.Location) bool {
	if location == nil {
		return false
	}
	localNow := now.In(location)
	monthYear, monthNumber, _ := month.UTC().Date()
	return localNow.Year() > monthYear || (localNow.Year() == monthYear && localNow.Month() > monthNumber)
}

func (d CalendarDate) Value() (driver.Value, error) {
	if d.IsZero() {
		return nil, nil
	}
	return d.UTC().Format(calendarDateLayout), nil
}

func (d *CalendarDate) Scan(value any) error {
	if value == nil {
		d.Time = time.Time{}
		return nil
	}
	switch typed := value.(type) {
	case time.Time:
		d.Time = time.Date(typed.Year(), typed.Month(), typed.Day(), 0, 0, 0, 0, time.UTC)
		return nil
	case string:
		parsed, err := ParseCalendarDate(typed)
		if err != nil {
			return err
		}
		d.Time = parsed.Time
		return nil
	case []byte:
		parsed, err := ParseCalendarDate(string(typed))
		if err != nil {
			return err
		}
		d.Time = parsed.Time
		return nil
	default:
		return fmt.Errorf("cannot scan calendar date from %T", value)
	}
}

func (d CalendarDate) MarshalJSON() ([]byte, error) {
	if d.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(d.UTC().Format(calendarDateLayout))
}

func (d *CalendarDate) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		d.Time = time.Time{}
		return nil
	}
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	parsed, err := ParseCalendarDate(value)
	if err != nil {
		return err
	}
	d.Time = parsed.Time
	return nil
}

func (d CalendarDate) String() string {
	if d.IsZero() {
		return ""
	}
	return d.UTC().Format(calendarDateLayout)
}

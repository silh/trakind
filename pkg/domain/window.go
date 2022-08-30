package domain

import (
	"errors"
	"strings"
	"time"
)

// DateFormat is default format that is returned by the server.
var DateFormat = "2006-01-02"
var TimeFormat = "15:04"

// Date is used to parse the date format used by
type Date time.Time

func (d *Date) UnmarshalJSON(bytes []byte) error {
	if string(bytes) == "null" {
		return nil
	}
	s := strings.ReplaceAll(string(bytes), "\"", "")
	date, err := time.Parse(DateFormat, s)
	if err != nil {
		return err
	}
	*d = Date(date)
	return nil
}

func (d *Date) MarshalJSON() ([]byte, error) {
	t := time.Time(*d)
	if y := t.Year(); y < 0 || y >= 10000 {
		return nil, errors.New("Time.MarshalJSON: year outside of range [0,9999]")
	}

	b := make([]byte, 0, len(DateFormat)+2)
	b = append(b, '"')
	b = t.AppendFormat(b, DateFormat)
	b = append(b, '"')
	return b, nil
}

func (d *Date) String() string {
	date := time.Time(*d)
	return date.Format(DateFormat)
}

func (d *Date) Before(another Date) bool {
	return time.Time(another).Before(time.Time(*d))
}

type TimeOfDay time.Time

func (t *TimeOfDay) String() string {
	date := time.Time(*t)
	return date.Format(TimeFormat)
}

func (t *TimeOfDay) UnmarshalJSON(bytes []byte) error {
	if string(bytes) == "null" {
		return nil
	}
	s := strings.ReplaceAll(string(bytes), "\"", "")
	date, err := time.Parse(TimeFormat, s)
	if err != nil {
		return err
	}
	*t = TimeOfDay(date)
	return nil
}

func (t *TimeOfDay) MarshalJSON() ([]byte, error) {
	stdTime := time.Time(*t)
	b := make([]byte, 0, len(TimeFormat)+2)
	b = append(b, '"')
	b = stdTime.AppendFormat(b, TimeFormat)
	b = append(b, '"')
	return b, nil
}

// TimeWindow describes one time open window in IND schedule.
type TimeWindow struct {
	//Key       string    `json:"key"`
	Date      Date      `json:"date"`
	StartTime TimeOfDay `json:"startTime"`
	EndTime   TimeOfDay `json:"endTime"`
	Parts     int       `json:"parts"` // number of people
}

// DatesResponse is full response received from API.
type DatesResponse struct {
	Status string       `json:"status"`
	Data   []TimeWindow `json:"data"`
}

func ParseWindowDate(value string) (Date, error) {
	date, err := time.Parse(DateFormat, value)
	return Date(date), err
}

package domain

import (
	"errors"
	"strings"
	"time"
)

// TimeFormat is default format that is returned by the server.
var TimeFormat = "2006-01-02"

// WindowDate is used to parse the date format used by
type WindowDate time.Time

func (r *WindowDate) UnmarshalJSON(bytes []byte) error {
	if string(bytes) == "null" {
		return nil
	}
	s := strings.ReplaceAll(string(bytes), "\"", "")
	date, err := time.Parse(TimeFormat, s)
	if err != nil {
		return err
	}
	*r = WindowDate(date)
	return nil
}

func (r WindowDate) MarshalJSON() ([]byte, error) {
	t := time.Time(r)
	if y := t.Year(); y < 0 || y >= 10000 {
		return nil, errors.New("Time.MarshalJSON: year outside of range [0,9999]")
	}

	b := make([]byte, 0, len(TimeFormat)+2)
	b = append(b, '"')
	b = t.AppendFormat(b, TimeFormat)
	b = append(b, '"')
	return b, nil
}

func (r WindowDate) String() string {
	date := time.Time(r)
	return date.Format(TimeFormat)
}

func (r WindowDate) Before(another WindowDate) bool {
	return time.Time(another).Before(time.Time(r))
}

// TimeWindow describes one time open window in IND schedule.
type TimeWindow struct {
	Key       string     `json:"key"`
	Date      WindowDate `json:"date"`
	StartTime string     `json:"startTime"`
	EndTime   string     `json:"endTime"`
	Parts     int        `json:"parts"`
}

// DatesResponse is full response received from API.
type DatesResponse struct {
	Status string       `json:"status"`
	Data   []TimeWindow `json:"data"`
}

func ParseWindowDate(value string) (WindowDate, error) {
	date, err := time.Parse(TimeFormat, value)
	return WindowDate(date), err
}

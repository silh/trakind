package domain

const MaxPeopleCount = 6

type ChatID int64

// Subscription to new available TimeWindows.
type Subscription struct {
	ChatID ChatID `json:"chatID"`
	// This is a workaround for us to use Subscription as comparable.
	TrackBefore Date   `json:"trackBefore"`
	PeopleCount int    `json:"peopleCount"`
	Action      string `json:"action,omitempty"`
}

// Matches returns true if Subscription matches given TimeWindow.
func (s *Subscription) Matches(window TimeWindow) bool {
	return s.TrackBefore == Date{} || s.TrackBefore.Before(window.Date)
}

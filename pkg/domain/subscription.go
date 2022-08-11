package domain

type ChatID int64

// Subscription to new available TimeWindows.
type Subscription struct {
	ChatID ChatID `json:"chatID"`
	// This is a workaround for us to use Subscription as comparable.
	TrackBefore WindowDate `json:"trackBefore"`
}

// Matches returns true if Subscription matches given TimeWindow.
func (s *Subscription) Matches(window TimeWindow) bool {
	return s.TrackBefore == WindowDate{} || s.TrackBefore.Before(window.Date)
}

package bots

import (
	"context"
	"encoding/json"
	"fmt"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/silh/trakind/pkg/db"
	"github.com/silh/trakind/pkg/domain"
	"io"
	"net/http"
	"sort"
	"time"
)

var client = http.Client{
	Timeout: 10 * time.Second,
}

type Fetcher struct {
	path        string
	location    domain.Location
	action      domain.Action
	peopleCount int
	interval    time.Duration
	bot         *Bot
}

func NewFetcher(
	location domain.Location,
	action domain.Action,
	peopleCount int,
	interval time.Duration,
	bot *Bot,
) *Fetcher {
	const INDApiPath = "https://oap.ind.nl/oap/api/desks/%s/slots/?productKey=%s&persons=%d"
	return &Fetcher{
		path:        fmt.Sprintf(INDApiPath, location.Code, action.Code, peopleCount),
		location:    location,
		action:      action,
		peopleCount: peopleCount,
		interval:    interval,
		bot:         bot,
	}
}

func (f *Fetcher) Track(ctx context.Context) {
	ticker := time.NewTicker(f.interval)
	for {
		select {
		case <-ctx.Done():
			log.Infow("Stopped tracking", "location", f.location.Code, "peopleCount", f.peopleCount)
			return
		case <-ticker.C:
			f.trackOnce()
		}
	}
}

func (f *Fetcher) trackOnce() {
	log := log.With("location", f.location.Code)
	subscriptions := f.getSubscriptionsFiltered()
	if len(subscriptions) == 0 {
		log.Debug("No subscribers, not fetching")
		return
	}
	datesResponse, err := f.getDates(f.path)
	if err != nil {
		log.Warnw("Error fetching dates", "path", f.path, "err", err)
		return
	}
	windows := datesResponse.Data
	if len(windows) == 0 {
		return
	}
	log.Debugw("Windows available!", "count", len(windows))
	firstAvailableWindow := windows[0]
	for _, subscription := range subscriptions {
		// we only need to check the first one as it's the earliest
		if subscription.Matches(firstAvailableWindow) {
			msgText := fmt.Sprintf(
				"A slot is available for %s at %s on %s at %s and %d more.",
				f.action.Name,
				f.location.Name,
				&firstAvailableWindow.Date,
				&firstAvailableWindow.StartTime,
				countAdditionalWindows(subscription, windows),
			)
			if _, err := f.bot.API.Send(tg.NewMessage(int64(subscription.ChatID), msgText)); err != nil {
				log.Warnw("Failed to send notification", "chat", subscription.ChatID, "err", err)
				// Remove subscription if user is deactivated
				if respErr, ok := err.(tg.Error); ok && respErr.Code == 403 && respErr.Message == "Forbidden: user is deactivated" {
					if err := db.Subscriptions.RemoveFromLocation(f.location.Code, subscription); err != nil {
						log.Warnw("Failed to delete subscription", "chat", subscription.ChatID, "err", err)
					} else {
						log.Infow("Deleted subscription for inactive user", "chat", subscription.ChatID)
					}
				}
			}
		}
	}
}

// getSubscriptionsFiltered returns only subscriptions that match current fetchers action and peopleCount.
// TODO move this to DB
func (f *Fetcher) getSubscriptionsFiltered() []domain.Subscription {
	log := log.With("location", f.location.Code)
	subscriptions, err := db.Subscriptions.GetForLocation(f.location.Code)
	if err != nil {
		log.Warnw("Could not retrieve subscriptions", "err", err)
		return nil
	}
	filtered := make([]domain.Subscription, 0, len(subscriptions))
	for _, subscription := range subscriptions {
		if subscription.Action == f.action.Code &&
			subscription.PeopleCount == f.peopleCount { // this check was previously in domain.Subscription.Matches
			filtered = append(filtered, subscription)
		}
	}
	return filtered
}

func (f *Fetcher) getDates(path string) (domain.DatesResponse, error) {
	resp, err := client.Get(path)
	if err != nil {
		return domain.DatesResponse{}, fmt.Errorf("could not fetch: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return domain.DatesResponse{}, fmt.Errorf("incorrect status code: %d (%s)", resp.StatusCode, resp.Status)
	}
	// response has prefix )]}',\n
	// we need to discard that
	const bytesToDiscard int64 = 6
	if _, err = io.CopyN(io.Discard, resp.Body, bytesToDiscard); err != nil {
		return domain.DatesResponse{}, fmt.Errorf("failed to read first bytes: %w", err)
	}
	var datesResponse domain.DatesResponse
	if err := json.NewDecoder(resp.Body).Decode(&datesResponse); err != nil {
		return domain.DatesResponse{}, fmt.Errorf("failed decoding: %w", err)
	}
	return datesResponse, nil
}

// countAdditionalWindows checks how many windows besides the first one match the subscription.
func countAdditionalWindows(subscription domain.Subscription, windows []domain.TimeWindow) int {
	otherDates := windows[1:]
	// time windows are ordered in the response, so we just need to find the first window that doesn't match
	return sort.Search(len(otherDates), func(i int) bool {
		return !subscription.Matches(otherDates[i])
	})
}

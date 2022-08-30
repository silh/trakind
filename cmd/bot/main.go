package main

import (
	"context"
	"encoding/json"
	"fmt"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/silh/trakind/pkg/bots"
	"github.com/silh/trakind/pkg/db"
	"github.com/silh/trakind/pkg/domain"
	"github.com/silh/trakind/pkg/loggers"
	"io"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"sync"
	"syscall"
	"time"
)

var log = loggers.Logger()

var interval = 1 * time.Minute

func main() {
	apiKey := os.Getenv("TELEGRAM_API_KEY")
	if apiKey == "" {
		log.Fatal("TELEGRAM_API_KEY env variable must be set")
	}
	setUpdateIntervalFromEnv()
	migrateSubscriptions() // temporary step

	bot, err := bots.New(apiKey)
	if err != nil {
		log.Fatalw("Failed to create new bot API", "err", err)
	}

	ctx := context.Background()
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	go func() {
		<-ctx.Done()
		bot.Stop()
	}()

	// Start tracking all available locations
	// Also track for different number of people because calculating it locally somehow doesn't produce the same result
	var wg sync.WaitGroup
	for location := range db.LocationToName {
		for i := 1; i <= 6; i++ {
			wg.Add(1)
			go func(location string, peopleCount int) {
				defer wg.Done()
				track(ctx, location, peopleCount, bot)
			}(location, i)
		}
	}
	bot.Run() // blocks until done
	wg.Wait()
	log.Info("Exiting")
}

func track(ctx context.Context, location string, peopleCount int, botAPI *bots.Bot) {
	log := log.With("location", location)
	// TODO only documents pick up is supported.
	const INDApiPath = "https://oap.ind.nl/oap/api/desks/%s/slots/?productKey=DOC&persons=%d"
	path := fmt.Sprintf(INDApiPath, location, peopleCount)
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ctx.Done():
			log.Info("Stopped tracking")
			return
		case <-ticker.C:
			trackOnce(botAPI, path, location)
		}
	}
}

func trackOnce(bot *bots.Bot, path, location string) {
	log := log.With("location", location)
	datesResponse, err := getDates(path)
	if err != nil {
		log.Warnw("Error fetching dates", "err", err)
		return
	}
	windows := datesResponse.Data
	if len(windows) == 0 {
		return
	}
	log.Debugw("Windows available!", "count", len(windows))
	subscriptions, err := db.Subscriptions.GetForLocation(location)
	if err != nil {
		log.Warnw("Could not retrieve subscriptions", "err", err)
		return
	}
	firstAvailableWindow := windows[0]
	for _, subscription := range subscriptions {
		// we only need to check the first one as it's the earliest
		if subscription.Matches(firstAvailableWindow) {
			msgText := fmt.Sprintf(
				"A slot is available at %s on %s at %s and %d more.",
				db.LocationToName[location],
				&firstAvailableWindow.Date,
				&firstAvailableWindow.StartTime,
				countAdditionalWindows(subscription, windows),
			)
			if _, err := bot.API.Send(tg.NewMessage(int64(subscription.ChatID), msgText)); err != nil {
				log.Warnw("Failed to send notification", "chat", subscription.ChatID)
			}
		}
	}
}

func getDates(path string) (domain.DatesResponse, error) {
	resp, err := http.Get(path)
	if err != nil {
		return domain.DatesResponse{}, fmt.Errorf("could not fetch: %w", err)
	}
	defer resp.Body.Close()
	// response has prefix )]}',\n
	// we need to discard that
	const bytesToDiscard int64 = 6
	if _, err = io.CopyN(io.Discard, resp.Body, bytesToDiscard); err != nil {
		return domain.DatesResponse{}, fmt.Errorf("error reading first bytes: %w", err)
	}
	var datesResponse domain.DatesResponse
	if err := json.NewDecoder(resp.Body).Decode(&datesResponse); err != nil {
		return domain.DatesResponse{}, fmt.Errorf("error decoding: %w", err)
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

func setUpdateIntervalFromEnv() {
	intervalFromEnv := os.Getenv("UPDATE_INTERVAL")
	if intervalFromEnv != "" {
		duration, err := time.ParseDuration(intervalFromEnv)
		if err == nil {
			interval = duration
		} else {
			log.Warnw("Could not parse duration from env UPDATE_INTERVAL", "err", err)
		}
	}
}

func migrateSubscriptions() {
	for _, location := range db.Locations {
		subscriptions, err := db.Subscriptions.GetForLocation(location.Code)
		if err != nil {
			panic(err)
		}
		for _, subscription := range subscriptions {
			if subscription.PeopleCount == 0 {
				subscriptionCopy := domain.Subscription{
					ChatID:      subscription.ChatID,
					TrackBefore: subscription.TrackBefore,
					PeopleCount: 1,
				}
				if err := db.Subscriptions.RemoveFromLocation(location.Code, subscription); err != nil {
					panic(err)
				}
				if err := db.Subscriptions.AddToLocation(location.Code, subscriptionCopy); err != nil {
					panic(err)
				}
			}
		}
	}
}

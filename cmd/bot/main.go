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
	"sync"
	"syscall"
	"time"
)

// INDApiPath TODO only documents pick up is supported for 3 people.
const INDApiPath = "https://oap.ind.nl/oap/api/desks/%s/slots/?productKey=DOC&persons=3"

var log = loggers.Logger()

var interval = 1 * time.Minute

func main() {
	apiKey := os.Getenv("TELEGRAM_API_KEY")
	if apiKey == "" {
		log.Fatal("TELEGRAM_API_KEY env variable must be set")
	}
	setUpdateIntervalFromEnv()

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
	var wg sync.WaitGroup
	for location := range db.LocationToName {
		wg.Add(1)
		go func(location string) {
			defer wg.Done()
			track(ctx, location, bot)
		}(location)
	}
	bot.Run() // blocks until done
	wg.Wait()
	log.Info("Exiting")
}

func track(ctx context.Context, location string, botAPI *bots.Bot) {
	log := log.With("location", location)
	path := fmt.Sprintf(INDApiPath, location)
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
	resp, err := http.Get(path)
	if err != nil {
		log.Warnw("Could not fetch", "err", err)
		return
	}
	defer resp.Body.Close()
	// response has prefix )]}',\n
	// we need to discard that
	var n int64 = 6
	_, err = io.CopyN(io.Discard, resp.Body, n)
	if err != nil {
		log.Warnw("Error reading first bytes", "err", err)
		return
	}
	var datesResponse domain.DatesResponse
	if err := json.NewDecoder(resp.Body).Decode(&datesResponse); err != nil {
		log.Infow("Error decoding", "err", err)
		return
	}
	if len(datesResponse.Data) == 0 {
		return
	}
	firstAvailableWindow := datesResponse.Data[0]
	log.Debugw("Windows available!", "count", len(datesResponse.Data))
	subscriptions, err := db.Subscriptions.GetForLocation(location)
	if err != nil {
		log.Warnw("Could not retrieve subscriptions", "err", err)
		return
	}
	for _, subscription := range subscriptions {
		if subscription.Matches(datesResponse.Data[0]) {
			count := countAdditionalWindows(subscription, datesResponse)
			msgText := fmt.Sprintf(
				"A slot is available at %s on %s at %s and %d more.",
				db.LocationToName[location],
				firstAvailableWindow.Date,
				firstAvailableWindow.StartTime,
				count,
			)
			_, err := bot.API.Send(tg.NewMessage(int64(subscription.ChatID), msgText))
			if err != nil {
				log.Warnw("Failed to send notification", "chat", subscription)
			}
		}
	}
}

func countAdditionalWindows(subscription domain.Subscription, datesResponse domain.DatesResponse) int64 {
	// TODO this can be improved as dates are ordered
	var count int64
	for _, window := range datesResponse.Data[1:] {
		if subscription.Matches(window) {
			count++
		}
	}
	return count
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

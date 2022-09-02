package main

import (
	"context"
	"github.com/silh/trakind/pkg/bots"
	"github.com/silh/trakind/pkg/db"
	"github.com/silh/trakind/pkg/loggers"
	"os"
	"os/signal"
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
	// Also track different types of actions
	var wg sync.WaitGroup
	for _, location := range db.Locations {
		for peopleCount := 1; peopleCount <= 6; peopleCount++ {
			for action := range location.AvailableActions {
				wg.Add(1)
				fetcher := bots.NewFetcher(location, action, peopleCount, interval, bot)
				go func() {
					defer wg.Done()
					fetcher.Track(ctx)
				}()
			}
		}
	}
	go reportNumberOfSubscriptions(ctx)
	bot.Run() // blocks until done
	wg.Wait()
	log.Info("Exiting")
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

// reportNumberOfSubscriptions periodically prints number of subscriptions per location. Has infinite cycle until passed
// context is Done. Doesn't take into consideration the action type
func reportNumberOfSubscriptions(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	for {
		report := make(map[string]int)
		total := 0
		for _, location := range db.Locations {
			countForLocation, err := db.Subscriptions.CountForLocation(location.Code)
			if err != nil {
				log.Warnw("Failed to get count", "location", location, "err", err)
				continue
			}
			report[location.Code] = countForLocation
			total += countForLocation
		}
		report["total"] = total
		log.Infow("Subscription report", "report", report)
		select {
		case <-ticker.C:
			// do nothing
		case <-ctx.Done():
			return
		}
	}
}

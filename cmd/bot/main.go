package main

import (
	"context"
	"encoding/json"
	"fmt"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/silh/trakind/pkg/bots"
	"github.com/silh/trakind/pkg/db"
	"github.com/silh/trakind/pkg/logger"
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

var log = logger.Logger()

var interval = 1 * time.Minute

// TimeWindow describes one time open window in IND schedule.
type TimeWindow struct {
	Key       string `json:"key"`
	Date      string `json:"date"`
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
	Parts     int    `json:"parts"`
}

// DatesResponse is full response received from API.
type DatesResponse struct {
	Status string       `json:"status"`
	Data   []TimeWindow `json:"data"`
}

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
	for location := range db.LocationToChats {
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
	var datesResponse DatesResponse
	if err := json.NewDecoder(resp.Body).Decode(&datesResponse); err != nil {
		log.Infow("Error decoding", "err", err)
		return
	}
	if len(datesResponse.Data) > 0 {
		firstAvailableWindow := datesResponse.Data[0]
		log.Debugw("Windows available!", "count", len(datesResponse.Data))
		msgText := fmt.Sprintf(
			"A slot is available at %s on %s at %s and %d more.",
			db.LocationToName[location],
			firstAvailableWindow.Date,
			firstAvailableWindow.StartTime,
			len(datesResponse.Data)-1,
		)
		db.LocationToChats[location].ForEach(func(chatID int64) {
			_, err := bot.API.Send(tg.NewMessage(chatID, msgText)) // TODO Should we delete a subscriber after update?
			if err != nil {
				log.Warnw("Failed to send notification", "chat", chatID)
			}
		})
	}
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

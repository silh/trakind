package main

import (
	"encoding/json"
	"fmt"
	bot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
	"io"
	"net/http"
	"os"
	"parseind/sets"
	"strings"
	"sync"
	"time"
)

// IndApiPath TODO only documents pick up is supported for 3 people.
const IndApiPath = "https://oap.ind.nl/oap/api/desks/%s/slots/?productKey=DOC&persons=3"

var log *zap.SugaredLogger

var lock sync.RWMutex
var locationToChats = map[string]sets.Set[int64]{
	"AM": sets.NewConcurrent[int64](),
	"DH": sets.NewConcurrent[int64](),
	"ZW": sets.NewConcurrent[int64](),
	"DB": sets.NewConcurrent[int64](),
}

// optionFunc wraps a func, so it satisfies the Option interface.
type optionFunc func(*zap.Logger)

func (f optionFunc) apply(log *zap.Logger) {
	f(log)
}

var count int // TODO this should survive restarts. And everything else as well :D

func init() {
	config := zap.NewDevelopmentConfig()
	config.Encoding = "console"
	config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	logger, err := config.Build()
	if err != nil {
		zap.S().Fatalw("Failed to create logger", "err", err)
	}
	log = logger.Sugar()
}

type Window struct {
	Key       string `json:"key"`
	Date      string `json:"date"`
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
	Parts     int    `json:"parts"`
}

type DatesResponse struct {
	Status string   `json:"status"`
	Data   []Window `json:"data"`
}

func main() {
	apiKey := os.Getenv("TELEGRAM_API_KEY")
	if apiKey == "" {
		log.Fatal("TELEGRAM_API_KEY must be provided")
	}
	botAPI, err := bot.NewBotAPI(apiKey)
	if err != nil {
		log.Fatalw("Failed to create new bot API", "err", err)
	}
	registerCommands(botAPI)
	u := bot.NewUpdate(0)
	u.Timeout = 60
	updatesC := botAPI.GetUpdatesChan(u)

	// Start tracking all available locations
	for k := range locationToChats {
		go track(k, botAPI)
	}

	// TODO improve that...
	for update := range updatesC {
		msg := update.Message
		if msg == nil || !msg.IsCommand() {
			continue
		}
		chatID := update.FromChat().ID
		log := log.With("chat", chatID)
		log.Infow("New command", "resp", msg.Text)
		command := msg.Command()
		args := strings.Split(msg.CommandArguments(), " ")
		if command == "track" {
			if len(args) == 0 {
				// TODO track all????
				log.Warn("Requested to track all locations. Not supported")
				_, err := botAPI.Send(bot.NewMessage(
					chatID,
					"Tracking all locations is not supported at the moment, please specify a location to track",
				))
				if err != nil {
					log.Warnw("Failed to send warning", "err", err)
				}
				continue
			}
			location := args[0]
			func() {
				lock.Lock()
				defer lock.Unlock()
				if _, ok := locationToChats[location]; ok {
					locationToChats[location].Add(chatID)
					log.Infow("New follower", "location", location)
				} else {
					_, err := botAPI.Send(bot.NewMessage(
						chatID,
						"Unsupported location - "+location,
					))
					if err != nil {
						log.Warnw("Failed to notify about incorrect location", "err", err)
					}
				}
			}()
		} else if command == "stoptrack" {
			func() {
				lock.Lock()
				defer lock.Unlock()
				for k := range locationToChats {
					locationToChats[k].Remove(chatID)
				}
			}()
		} else if command == "start" {
			count++
			log.Info("New user", "count", count)
		}
	}
}

func registerCommands(botAPI *bot.BotAPI) {
	commands := bot.NewSetMyCommands(
		bot.BotCommand{
			Command: "track",
			Description: "Specify IND location to track for available timeslots - AM (for Amsterdam), " +
				"DH (for Den Haag), ZW (for Zwolle), DB (for Den Dosch). " +
				"Optionally you can specify the date as DD.MM.YYYY, then you will only get notified about slots before or on that date",
		},
		bot.BotCommand{
			Command:     "stoptrack",
			Description: "Stops all tracking",
		},
	)
	resp, err := botAPI.Request(commands)
	if err != nil {
		log.Fatalw("Failed to register commands", "err", err)
	}
	if !resp.Ok {
		log.Fatalw("Failed to register commands", "code", resp.ErrorCode, "desc", resp.Description)
	}
	log.Infow("Command registration successful")
}

func track(location string, botAPI *bot.BotAPI) {
	path := fmt.Sprintf(IndApiPath, location)
	ticker := time.NewTicker(10 * time.Second)
	for {
		<-ticker.C
		trackOnce(botAPI, path, location)
		// TODO add graceful shutdown
	}
}

func trackOnce(botAPI *bot.BotAPI, path, location string) {
	log := log.With("location", location)
	resp, err := http.Get(path)
	if err != nil {
		log.Warnw("Could not fetch", "err", err)
		return
	}
	defer resp.Body.Close()
	// response has prefix )]}',\n
	// we need to discard that
	var n int64 = 6 // response has prefix )]}',\n
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
		log.Debugw("Date available!", "first", datesResponse.Data[0])
		lock.RLock()
		defer lock.RUnlock()
		locationToChats[location].ForEach(func(chatID int64) {
			_, err := botAPI.Send(bot.NewMessage(chatID, "Hey, you have a new available slot!"))
			if err != nil {
				log.Warnw("Failed to send notification", "chat", chatID)
			}
		})
	}
}

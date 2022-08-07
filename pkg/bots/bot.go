package bots

import (
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/silh/trakind/pkg/db"
	"github.com/silh/trakind/pkg/logger"
	"strings"
)

var log = logger.Logger()

type ChatID int64

type SubscribersDB interface {
	Add(location string, chatID ChatID)
}

type Bot struct {
	API           *tg.BotAPI // FIXME should not expose that
	subscribersDB SubscribersDB

	openChatCount uint64 // TODO this should survive restarts. And everything else should as well :D
}

func New(apiKey string) (*Bot, error) {
	api, err := tg.NewBotAPI(apiKey)
	if err != nil {
		return nil, err
	}
	return &Bot{
		API:           api,
		subscribersDB: nil,
	}, nil
}

// Run starts main loop receiving and processing updates. Exists only when the updates channel is closed.
func (b *Bot) Run() {
	b.registerCommands()
	u := tg.NewUpdate(0)
	u.Timeout = 5
	updatesC := b.API.GetUpdatesChan(u)
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
				_, err := b.API.Send(tg.NewMessage(
					chatID,
					"Tracking all locations is not supported at the moment, please specify a location to track",
				))
				if err != nil {
					log.Warnw("Failed to send warning", "err", err)
				}
				continue
			}
			location := strings.ToUpper(args[0])
			if _, ok := db.LocationToChats[location]; ok {
				db.LocationToChats[location].Add(chatID)
				_, err := b.API.Send(tg.NewMessage(chatID, "You will now get a notification when there is "+
					"an open time window found for the location "+db.LocationToName[location]))
				if err != nil {
					log.Warnw("Failed to notify about subscription", "err", err)
				}
				log.Infow("New follower", "location", location)
				continue
			}
			_, err := b.API.Send(tg.NewMessage(chatID, "Unsupported location - "+location))
			if err != nil {
				log.Warnw("Failed to notify about incorrect location", "err", err)
			}
		} else if command == "stoptrack" {
			for k := range db.LocationToChats {
				db.LocationToChats[k].Remove(chatID)
			}
			_, err := b.API.Send(tg.NewMessage(chatID, "You won't receive new notifications."))
			if err != nil {
				log.Warnw("Failed to notify about unsubscription", "err", err)
			}
		} else if command == "start" {
			b.openChatCount++
			log.Info("New user", "count", b.openChatCount)
		} else if command == "stop" {
			b.openChatCount--
			log.Info("User left", "count", b.openChatCount)
		}
	}
}

// Stop closes update channel and lets a goroutine that is in Run func to exit it.
func (b *Bot) Stop() {
	b.API.StopReceivingUpdates()
}

func (b *Bot) registerCommands() {
	commands := tg.NewSetMyCommands(
		tg.BotCommand{
			Command: "track",
			Description: "Specify IND location to track for available timeslots - AM (for Amsterdam), " +
				"DH (for Den Haag), ZW (for Zwolle), DB (for Den Dosch). " +
				"Optionally you can specify the date as DD.MM.YYYY, then you will only get notified about slots before or on that date",
		},
		tg.BotCommand{
			Command:     "stoptrack",
			Description: "Stops all tracking",
		},
	)
	resp, err := b.API.Request(commands)
	if err != nil {
		log.Fatalw("Failed to register commands", "err", err)
	}
	if !resp.Ok {
		log.Fatalw("Failed to register commands", "code", resp.ErrorCode, "desc", resp.Description)
	}
	log.Infow("Commands registration successful")
}

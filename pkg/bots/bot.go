package bots

import (
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/silh/trakind/pkg/db"
	"github.com/silh/trakind/pkg/domain"
	"github.com/silh/trakind/pkg/logger"
	"go.uber.org/zap"
	"strings"
	"time"
)

var log = logger.Logger()

type SubscribersDB interface {
	Add(location string, chatID domain.ChatID)
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
				b.sendAndForget(
					tg.NewMessage(
						chatID,
						"Tracking all locations is not supported at the moment, please specify a location to track",
					),
					"warning",
					log,
				)
				continue
			}
			location := strings.ToUpper(args[0])
			if _, ok := db.LocationToChats[location]; ok {
				subscription := domain.Subscription{
					ChatID: domain.ChatID(chatID),
				}
				if len(args) > 1 {
					date, err := time.Parse(domain.TimeFormat, args[1])
					if err != nil {
						b.sendAndForget(
							tg.NewMessage(chatID, "Date has incorrect format, please use YYYY-MM-DD"),
							"message about incorrect date",
							log,
						)
						continue
					}
					subscription.TrackBefore = domain.WindowDate(date)
				}
				db.LocationToChats[location].Add(subscription)
				msgText := "You will now get a notification when an open time window" +
					" found for the location " + db.LocationToName[location]
				if (subscription.TrackBefore != domain.WindowDate{}) {
					msgText += " that happens before " + subscription.TrackBefore.String()
				}
				b.sendAndForget(
					tg.NewMessage(chatID, msgText),
					"subsription notification",
					log,
				)
				log.Infow("New follower", "location", location)
				continue
			}
			b.sendAndForget(
				tg.NewMessage(chatID, "Unsupported location - "+location),
				"incorrect location message",
				log,
			)
		} else if command == "stoptrack" {
			for k := range db.LocationToChats {
				// TODO this should be improved
				toDelete := make([]domain.Subscription, 0)
				db.LocationToChats[k].ForEach(func(item domain.Subscription) {
					if item.ChatID == domain.ChatID(chatID) {
						toDelete = append(toDelete, item)
					}
				})
				for _, subscription := range toDelete {
					db.LocationToChats[k].Remove(subscription)
				}
			}
			b.sendAndForget(
				tg.NewMessage(chatID, "You won't receive new notifications."),
				"unsubsribption notification",
				log,
			)
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

// registerCommands registers available bot commands.
func (b *Bot) registerCommands() {
	commands := tg.NewSetMyCommands(
		tg.BotCommand{
			Command: "track",
			Description: "Specify IND location to track for available timeslots - AM (for Amsterdam), " +
				"DH (for Den Haag), ZW (for Zwolle), DB (for Den Dosch). " +
				"Optionally you can specify the date as YYYY-MM-DD, then you will only get notified about " +
				"time windows before that date.",
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

// sendAndForget sends message and logs error if it occurs.
func (b *Bot) sendAndForget(msg tg.MessageConfig, msgType string, log *zap.SugaredLogger) {
	if _, err := b.API.Send(msg); err != nil {
		log.Warnw("Failed to send "+msgType, "err", err)
	}
}

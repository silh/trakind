package bots

import (
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/silh/trakind/pkg/db"
	"github.com/silh/trakind/pkg/domain"
	"strings"
	"time"
)

type Command interface {
	Execute() error
}

type TrackCommand struct {
	Command
	bot    *Bot
	chatID domain.ChatID
	args   []string
}

func NewTrackCommand(bot *Bot, msg *tg.Message) *TrackCommand {
	return &TrackCommand{
		bot:    bot,
		chatID: domain.ChatID(msg.Chat.ID),
		args:   strings.Split(msg.CommandArguments(), " "),
	}
}

func (c *TrackCommand) Execute() error {
	log := log.With("chat", c.chatID)
	if len(c.args) == 0 { //TODO should this be checked in constructor?
		log.Warn("Requested to track all locations. Not supported")
		c.bot.sendAndForget(
			tg.NewMessage(
				int64(c.chatID),
				"Tracking all locations is not supported at the moment, please specify a location to track",
			),
			"warning",
			log,
		)
		return nil // TODO maybe return error here?
	}
	location := strings.ToUpper(c.args[0])
	if _, ok := db.LocationToName[location]; !ok {
		c.bot.sendAndForget(
			tg.NewMessage(int64(c.chatID), "Unsupported location - "+location),
			"incorrect location message",
			log,
		)
		return nil //TODO return error?
	}
	subscription := domain.Subscription{ChatID: c.chatID}
	if len(c.args) > 1 {
		date, err := time.Parse(domain.TimeFormat, c.args[1])
		if err != nil {
			c.bot.sendAndForget(
				tg.NewMessage(int64(c.chatID), "Date has incorrect format, please use YYYY-MM-DD"),
				"message about incorrect date",
				log,
			)
			return nil // TODO return error?
		}
		subscription.TrackBefore = domain.WindowDate(date)
	}
	if err := db.Subscriptions.AddToLocation(location, subscription); err != nil {
		log.Warnw("Failed to store new subscription", "err", err)
		return nil // TODO return error?
	}
	msgText := "You will now get a notification when an open time window" +
		" found for the location " + db.LocationToName[location]
	if (subscription.TrackBefore != domain.WindowDate{}) {
		msgText += " that happens before " + subscription.TrackBefore.String()
	}
	c.bot.sendAndForget(
		tg.NewMessage(int64(c.chatID), msgText),
		"subscription notification",
		log,
	)
	log.Infow("New follower", "location", location)
	return nil
}

type StopTrackCommand struct {
	bot    *Bot
	chatID domain.ChatID
}

func NewStopTrackCommand(bot *Bot, msg *tg.Message) *StopTrackCommand {
	return &StopTrackCommand{
		bot:    bot,
		chatID: domain.ChatID(msg.Chat.ID),
	}
}

func (c *StopTrackCommand) Execute() error {
	log := log.With("chat", c.chatID)
	for location := range db.LocationToName {
		// TODO this should be improved
		subscriptions, err := db.Subscriptions.GetForLocation(location)
		if err != nil {
			log.Warnw("Failed to get subscriptions for delete", "err", err)
			continue
		}
		for _, subscription := range subscriptions {
			if subscription.ChatID == c.chatID {
				if err := db.Subscriptions.RemoveFromLocation(location, subscription); err != nil {
					log.Warnw("Failed to delete subscription", "err", err)
				}
			}
		}
	}
	c.bot.sendAndForget(
		tg.NewMessage(int64(c.chatID), "You won't receive new notifications."),
		"unsubscription notification",
		log,
	)
	return nil // TODO
}

type StartCommand struct {
}

var startCMD = StartCommand{}

func NewStartCommand() *StartCommand {
	return &startCMD // to really new, but we don't have fields atm
}

func (c *StartCommand) Execute() error {
	db.Users.Increment()
	return nil
}

type StopCommand struct {
}

var stopCMD = StartCommand{}

func NewStopCommand() *StartCommand {
	return &stopCMD // to really new, but we don't have fields atm
}

func (c *StopCommand) Execute() error {
	db.Users.Increment()
	return nil
}

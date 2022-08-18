package bots

import (
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/silh/trakind/pkg/domain"
	"github.com/silh/trakind/pkg/loggers"
	"go.uber.org/zap"
)

var log = loggers.Logger()

// SubscribersDB TODO not used, should be moved
type SubscribersDB interface {
	AddToLocation(location string, subscription domain.Subscription) error
	RemoveFromLocation(location string, subscription domain.Subscription) error
	GetForLocation(location string) ([]domain.Subscription, error)
}

type Bot struct {
	API      *tg.BotAPI // FIXME should not expose that
	commands map[string]func(*Bot, *tg.Message) Command
}

func New(apiKey string) (*Bot, error) {
	api, err := tg.NewBotAPI(apiKey)
	if err != nil {
		return nil, err
	}
	return &Bot{
		API: api,
		commands: map[string]func(*Bot, *tg.Message) Command{
			"start":     func(*Bot, *tg.Message) Command { return NewStartCommand() },
			"stop":      func(*Bot, *tg.Message) Command { return NewStopCommand() },
			"track":     func(bot *Bot, msg *tg.Message) Command { return NewTrackCommand(bot, msg) },
			"stoptrack": func(bot *Bot, msg *tg.Message) Command { return NewStopTrackCommand(bot, msg) },
		},
	}, nil
}

// Run starts main loop receiving and processing updates. Exists only when the updates channel is closed.
func (b *Bot) Run() {
	b.registerCommands()
	u := tg.NewUpdate(0)
	u.Timeout = 5
	updatesC := b.API.GetUpdatesChan(u)
	for update := range updatesC {
		msg := update.Message
		if msg == nil || !msg.IsCommand() {
			continue
		}
		log.Infow("New command", "text", msg.Text, "chat", msg.Chat.ID)
		if cmdConstructor, ok := b.commands[msg.Command()]; ok {
			_ = cmdConstructor(b, msg).Execute() // TODO doesn't return any errors at the moment
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
func (b *Bot) sendAndForget(msg tg.MessageConfig, text string, log *zap.SugaredLogger) {
	if _, err := b.API.Send(msg); err != nil {
		log.Warnw("Failed to send notification", "err", err, "text", text)
	}
}

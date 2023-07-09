package bots

import (
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/silh/trakind/pkg/domain"
	"github.com/silh/trakind/pkg/loggers"
	"go.uber.org/zap"
)

var log = loggers.Logger()

var chatFSMs = map[domain.ChatID]*FSM{}

// SubscribersDB TODO not used, should be moved
type SubscribersDB interface {
	AddToLocation(location string, subscription domain.Subscription) error
	RemoveFromLocation(location string, subscription domain.Subscription) error
	GetForLocation(location string) ([]domain.Subscription, error)
}

type Bot struct {
	API *tg.BotAPI // FIXME should not expose that
}

func New(apiKey string) (*Bot, error) {
	api, err := tg.NewBotAPI(apiKey)
	if err != nil {
		return nil, err
	}
	return &Bot{
		API: api,
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
		// FIXME this is a WA, states should handle update instead of message
		if update.MyChatMember != nil &&
			update.MyChatMember.NewChatMember.WasKicked() {
			chatID := domain.ChatID(update.MyChatMember.Chat.ID)
			fsm := chatFSMs[chatID]
			if fsm == nil {
				fsm = NewFSM(chatID, b)
				chatFSMs[chatID] = fsm
			}
			fsm.To(stopCommandState, nil)
			continue
		}
		if msg == nil {
			continue
		}
		chatID := domain.ChatID(msg.Chat.ID)
		fsm := chatFSMs[chatID]
		if fsm == nil {
			fsm = NewFSM(chatID, b)
			chatFSMs[chatID] = fsm
		}
		fsm.Do(msg)
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
			Command:     "track",
			Description: "Start tracking new location",
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

// SendAndForget sends message and logs error if it occurs.
func (b *Bot) SendAndForget(msg tg.MessageConfig, log *zap.SugaredLogger) {
	if _, err := b.API.Send(msg); err != nil {
		// TODO probably need to handle unavailable users here as well
		log.Warnw("Failed to send notification", "err", err, "text", msg.Text)
	}
}

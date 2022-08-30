package bots

import (
	"fmt"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/silh/trakind/pkg/domain"
	"go.uber.org/zap"
)

type FSM struct {
	chatID domain.ChatID
	log    *zap.SugaredLogger
	bot    *Bot

	state State
}

func NewFSM(chatID domain.ChatID, bot *Bot) *FSM {
	return &FSM{
		chatID: chatID,
		log:    log.With("chat", chatID),
		bot:    bot,
		state:  initialState,
	}
}

func (fsm *FSM) To(newState State, msg *tg.Message) {
	fsm.log.Debugw("State transition", "from", fsm.state, "to", newState)
	fsm.state = newState
	fsm.state.To(fsm, msg, fsm.bot)
}

func (fsm *FSM) Do(msg *tg.Message) error {
	return fsm.state.Do(fsm, msg, fsm.bot)
}

type State interface {
	fmt.Stringer
	To(fsm *FSM, msg *tg.Message, bot *Bot)       // TODO maybe return new state at the end?
	Do(fsm *FSM, msg *tg.Message, bot *Bot) error // TODO maybe return new state at the end?
}

var initialState = &InitialState{}
var commandHandlingState = &CommandHandlingState{commands: map[string]State{
	"start":     startCommandState,
	"stop":      stopCommandState,
	"track":     whichLocationState,
	"stoptrack": stopTrackCommandState,
}}
var startCommandState = &StartCommandState{}
var stopCommandState = &StopCommandState{}
var doneState = &DoneState{}
var whichLocationState = &WhichLocationState{}
var stopTrackCommandState = &StopTrackCommandState{}

func newMessage(chatId domain.ChatID, text string) tg.MessageConfig {
	message := tg.NewMessage(int64(chatId), text)
	message.ReplyMarkup = tg.ReplyKeyboardRemove{RemoveKeyboard: true}
	return message
}

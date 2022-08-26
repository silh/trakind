package bots

import (
	"errors"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type DoneState struct {
}

func (s DoneState) String() string {
	return "DoneState"
}

func (s DoneState) To(fsm *FSM, _ *tg.Message, _ *Bot) {
	delete(chatFSMs, fsm.chatID) // TODO this is ugly
	fsm.log.Debug("We are done")
}

func (s DoneState) Do(*FSM, *tg.Message, *Bot) error {
	return errors.New("should never be called")
}

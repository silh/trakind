package bots

import (
	"errors"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/silh/trakind/pkg/db"
)

type StartCommandState struct {
}

func (s StartCommandState) String() string {
	return "StartCommandState"
}

func (s StartCommandState) To(fsm *FSM, msg *tg.Message, _ *Bot) {
	db.Users.Increment()
	fsm.To(doneState, msg)
}

func (s StartCommandState) Do(*FSM, *tg.Message, *Bot) error {
	panic(errors.New("should not be called"))
}

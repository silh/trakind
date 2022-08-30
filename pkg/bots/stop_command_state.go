package bots

import (
	"errors"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/silh/trakind/pkg/db"
)

type StopCommandState struct {
}

func (s StopCommandState) String() string {
	return "StopCommandState"
}

func (s StopCommandState) To(*FSM, *tg.Message, *Bot) {
	db.Users.Decrement()
}

func (s StopCommandState) Do(*FSM, *tg.Message, *Bot) error {
	return errors.New("should never be called")
}

package bots

import (
	"fmt"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/silh/trakind/pkg/db"
)

type WhichActionState struct {
}

func (s *WhichActionState) String() string {
	return "WhichActionState"
}

func (s *WhichActionState) To(fsm *FSM, msg *tg.Message, bot *Bot) {
	toSend := newMessage(fsm.chatID, "Which type of appointment are you interested in?")
	toSend.ReplyMarkup = s.makeReplyKeyboard()
	if _, err := bot.API.Send(toSend); err != nil {
		fsm.log.Warnw("Failed to send message", "msg", toSend.Text, "err", err)
		fsm.To(doneState, msg)
	}
}

func (s *WhichActionState) Do(fsm *FSM, msg *tg.Message, bot *Bot) error {
	if msg.IsCommand() {
		fsm.To(commandHandlingState, msg)
		return nil
	}
	action, ok := db.ActionForName(msg.Text)
	if !ok {
		toSend := newMessage(
			fsm.chatID,
			fmt.Sprintf(
				"Appointment type %s is not supported, please click on a button with one of the available appointment types.",
				msg.Text,
			),
		)
		toSend.ReplyMarkup = s.makeReplyKeyboard()
		if _, err := bot.API.Send(toSend); err != nil {
			fsm.log.Warnw("Failed to send message", "msg", toSend.Text, "err", err)
			fsm.To(doneState, msg)
		}
		return nil
	}
	nextState := &WhichLocationState{action: action}
	fsm.To(nextState, msg)
	return nil
}

func (s *WhichActionState) makeReplyKeyboard() tg.ReplyKeyboardMarkup {
	row := make([]tg.KeyboardButton, 0, len(db.Actions))
	for action := range db.Actions {
		row = append(row, tg.NewKeyboardButton(action.Name))
	}
	return tg.NewOneTimeReplyKeyboard(row)
}

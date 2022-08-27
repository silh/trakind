package bots

import (
	"fmt"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/silh/trakind/pkg/db"
)

type WhichLocationState struct {
}

func (s *WhichLocationState) String() string {
	return "WhichLocationState"
}

func (s *WhichLocationState) To(fsm *FSM, msg *tg.Message, bot *Bot) {
	toSend := newMessage(fsm.chatID, "Which location?")
	toSend.ReplyMarkup = s.makeReplyKeyboard()
	if _, err := bot.API.Send(toSend); err != nil {
		fsm.log.Warnw("Failed to send message", "msg", toSend.Text, "err", err)
		fsm.To(doneState, msg)
	}
}

func (s *WhichLocationState) Do(fsm *FSM, msg *tg.Message, bot *Bot) error {
	if msg.IsCommand() {
		fsm.To(commandHandlingState, msg)
		return nil
	}
	location, ok := db.NameToLocation[msg.Text] // TODO should be more tolerable to case.
	if !ok {
		toSend := newMessage(
			fsm.chatID,
			fmt.Sprintf(
				"Location %s is incorrect, please click on a button with one of the available locations.",
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
	nextState := &HowManyPeopleState{location: location}
	fsm.To(nextState, msg)
	return nil
}

func (s *WhichLocationState) makeReplyKeyboard() tg.ReplyKeyboardMarkup {
	rows := make([][]tg.KeyboardButton, 0, len(db.LocationToName)/2)
	row := make([]tg.KeyboardButton, 0, 2)
	for i, location := range db.Locations {
		row = append(row, tg.NewKeyboardButton(location.Name))
		if len(row) == 2 {
			rows = append(rows, row)
			if i < len(db.Locations)-1 {
				row = make([]tg.KeyboardButton, 0, 2)
			}
		}
	}
	return tg.NewOneTimeReplyKeyboard(rows...)
}

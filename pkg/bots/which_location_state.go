package bots

import (
	"fmt"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/silh/trakind/pkg/db"
)

type WhichLocationState struct {
}

func (w *WhichLocationState) String() string {
	return "WhichLocationState"
}

func (w *WhichLocationState) To(fsm *FSM, msg *tg.Message, bot *Bot) {
	toSend := newMessage(fsm.chatID, "Which location?")
	toSend.ReplyMarkup = makeWhichLocationReplyMarkup()
	if _, err := bot.API.Send(toSend); err != nil {
		fsm.log.Warnw("Failed to send message", "msg", toSend.Text, "err", err)
		fsm.To(doneState, msg)
	}
}

func (w *WhichLocationState) Do(fsm *FSM, msg *tg.Message, bot *Bot) error {
	if msg.IsCommand() {
		fsm.To(commandHandlingState, msg)
		return nil
	}
	if location, ok := db.NameToLocation[msg.Text]; ok {
		fsm.To(&HowManyPeopleState{location: location}, msg)
		return nil
	}
	toSend := newMessage(
		fsm.chatID,
		fmt.Sprintf("Locations %s is incorrect, please select click on a button with one of the available locations.", msg.Text),
	)
	toSend.ReplyMarkup = makeWhichLocationReplyMarkup()
	if _, err := bot.API.Send(toSend); err != nil {
		fsm.log.Warnw("Failed to send message", "msg", toSend.Text, "err", err)
		fsm.To(doneState, msg)
	}
	return nil
}

func makeWhichLocationReplyMarkup() tg.ReplyKeyboardMarkup {
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

package bots

import (
	"errors"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/silh/trakind/pkg/db"
)

type StopTrackCommandState struct {
}

func (s StopTrackCommandState) String() string {
	return "StopTrackCommandState"
}

func (s StopTrackCommandState) To(fsm *FSM, msg *tg.Message, bot *Bot) {
	for _, location := range db.Locations {
		// TODO this should be improved
		subscriptions, err := db.Subscriptions.GetForLocation(location.Code)
		if err != nil {
			fsm.log.Warnw("Failed to get subscriptions for delete", "location", location, "err", err)
			continue
		}
		for _, subscription := range subscriptions {
			if subscription.ChatID == fsm.chatID {
				if err := db.Subscriptions.RemoveFromLocation(location.Code, subscription); err != nil {
					fsm.log.Warnw("Failed to delete subscription", "subscription", subscription, "err", err)
				} else {
					fsm.log.Infow("One less follower", "location", location.Code)
				}
			}
		}
	}
	toSend := newMessage(fsm.chatID, "You won't receive new notifications.")
	if _, err := bot.API.Send(toSend); err != nil {
		fsm.log.Warnw("Failed to send message", "msg", toSend.Text, "err", err)
	}
	fsm.To(doneState, msg)
}

func (s StopTrackCommandState) Do(*FSM, *tg.Message, *Bot) error {
	panic(errors.New("should never be called"))
}

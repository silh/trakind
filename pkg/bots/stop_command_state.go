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

func (s StopCommandState) To(fsm *FSM, _ *tg.Message, _ *Bot) {
	for _, location := range db.DocPickupLocations {
		subscriptions, err := db.Subscriptions.GetForLocation(location.Code)
		if err != nil {
			fsm.log.Warnw("Failed to get subscriptions", "location", location.Code, "err", err)
			continue
		}
		for _, subscription := range subscriptions {
			if subscription.ChatID == fsm.chatID {
				if err := db.Subscriptions.RemoveFromLocation(location.Code, subscription); err != nil {
					log.Warnw("Failed to delete subscription", "err", err)
					continue
				}
				fsm.log.Infow("Unsubscribed", "location", location.Code)
			}
		}
	}
	db.Users.Decrement()
	fsm.log.Info("Stopped")
	fsm.To(doneState, nil)
}

func (s StopCommandState) Do(*FSM, *tg.Message, *Bot) error {
	return errors.New("should never be called")
}

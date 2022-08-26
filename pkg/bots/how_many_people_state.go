package bots

import (
	"fmt"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/silh/trakind/pkg/db"
	"github.com/silh/trakind/pkg/domain"
	"strconv"
)

// Those are the limits specified on website
const (
	minPeople = 1
	maxPeople = 6
)

type HowManyPeopleState struct {
	location string
}

func (h *HowManyPeopleState) String() string {
	return "HowManyPeopleState"
}

func (h *HowManyPeopleState) To(fsm *FSM, msg *tg.Message, bot *Bot) {
	keyboard := makeHowManyPeopleReplyKeyboard()
	toSend := newMessage(fsm.chatID, "How many people?")
	toSend.ReplyMarkup = keyboard
	if _, err := bot.API.Send(toSend); err != nil {
		fsm.log.Warnw("Failed to send message", "msg", toSend.Text, "err", err)
		fsm.To(doneState, msg)
	}
}

func (h *HowManyPeopleState) Do(fsm *FSM, msg *tg.Message, bot *Bot) error {
	if msg.IsCommand() {
		fsm.To(commandHandlingState, msg)
		return nil
	}
	peopleCount, err := strconv.Atoi(msg.Text) // TODO should handle incorrect input
	if err != nil {
		toSend := newMessage(
			fsm.chatID,
			fmt.Sprintf("Please reply with number between %d and %d or click one of the buttons.", minPeople, maxPeople),
		)
		toSend.ReplyMarkup = makeWhichLocationReplyMarkup()
		if _, err = bot.API.Send(toSend); err != nil {
			fsm.log.Warnw("Failed to send message", "msg", toSend.Text, "err", err)
		}
		return nil
	}
	if peopleCount < minPeople || peopleCount > maxPeople {
		toSend := newMessage(
			fsm.chatID,
			fmt.Sprintf(
				"Incorrect number of people %d, please select between %d and %d or click one of the buttons",
				peopleCount, minPeople, maxPeople,
			),
		)
		toSend.ReplyMarkup = makeWhichLocationReplyMarkup()
		if _, err := bot.API.Send(toSend); err != nil {
			fsm.log.Warnw("Failed to send message", "msg", toSend.Text, "err", err)
			fsm.To(doneState, msg)
		}
		return nil
	}
	subscription := domain.Subscription{ChatID: fsm.chatID, PeopleCount: peopleCount}
	if err := db.Subscriptions.AddToLocation(h.location, subscription); err != nil {
		fsm.log.Warnw("Failed to store subscription", "subscription", subscription, "err", err)
		toSend := newMessage(fsm.chatID, "Failed to create subscription. Please try again.")
		if _, err = bot.API.Send(toSend); err != nil {
			fsm.log.Warnw("Failed to send message", "msg", toSend.Text, "err", err)
		}
		fsm.To(doneState, msg)
		return nil
	}
	toSend := newMessage(fsm.chatID, fmt.Sprintf("You will now get a notification when an open time window"+
		" found for the location %s for %d people", h.location, peopleCount))
	if _, err := bot.API.Send(toSend); err != nil {
		fsm.log.Warnw("Failed to send message", "msg", toSend.Text, "err", err)
	}
	fsm.To(doneState, msg)
	return nil
}

func makeHowManyPeopleReplyKeyboard() tg.ReplyKeyboardMarkup {
	row := make([]tg.KeyboardButton, 6)
	for i := 0; i < 6; i++ {
		row[i] = tg.NewKeyboardButton(strconv.Itoa(i + 1))
	}
	return tg.NewOneTimeReplyKeyboard(row)
}

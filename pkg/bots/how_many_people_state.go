package bots

import (
	"fmt"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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

func (s *HowManyPeopleState) String() string {
	return "HowManyPeopleState"
}

func (s *HowManyPeopleState) To(fsm *FSM, msg *tg.Message, bot *Bot) {
	toSend := newMessage(fsm.chatID, "How many people?")
	toSend.ReplyMarkup = s.makeReplyKeyboard()
	if _, err := bot.API.Send(toSend); err != nil {
		fsm.log.Warnw("Failed to send message", "msg", toSend.Text, "err", err)
		fsm.To(doneState, msg)
	}
}

func (s *HowManyPeopleState) Do(fsm *FSM, msg *tg.Message, bot *Bot) error {
	if msg.IsCommand() {
		fsm.To(commandHandlingState, msg)
		return nil
	}
	peopleCount, replyText, ok := s.getPeopleCount(msg)
	if !ok {
		toSend := newMessage(fsm.chatID, replyText)
		toSend.ReplyMarkup = s.makeReplyKeyboard()
		if _, err := bot.API.Send(toSend); err != nil {
			fsm.log.Warnw("Failed to send message", "msg", toSend.Text, "err", err)
			fsm.To(doneState, msg)
		}
		return nil
	}
	nextState := &BeforeDateState{location: s.location, peopleCount: peopleCount}
	fsm.To(nextState, msg)
	return nil
}

// getPeopleCount returns number of people from the message if it is valid. If it's not - return a message describing
// the problem and false as third value.
func (s *HowManyPeopleState) getPeopleCount(msg *tg.Message) (int, string, bool) {
	peopleCount, err := strconv.Atoi(msg.Text)
	if err != nil {
		return 0, fmt.Sprintf(
			"Please reply with a number between %d and %d or click one of the buttons.",
			minPeople, maxPeople,
		), false
	}
	if peopleCount < minPeople || peopleCount > maxPeople {
		return 0,
			fmt.Sprintf(
				"Incorrect number of people %d, please select between %d and %d or click one of the buttons",
				peopleCount, minPeople, maxPeople,
			), false
	}
	return peopleCount, "", true
}

func (s *HowManyPeopleState) makeReplyKeyboard() tg.ReplyKeyboardMarkup {
	row := make([]tg.KeyboardButton, 6)
	for i := 0; i < 6; i++ {
		row[i] = tg.NewKeyboardButton(strconv.Itoa(i + 1))
	}
	return tg.NewOneTimeReplyKeyboard(row)
}

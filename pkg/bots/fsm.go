package bots

import (
	"errors"
	"fmt"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/silh/trakind/pkg/db"
	"github.com/silh/trakind/pkg/domain"
	"go.uber.org/zap"
	"strconv"
)

type FSM struct {
	chatID domain.ChatID
	log    *zap.SugaredLogger
	bot    *Bot

	state State
}

func NewFSM(chatID domain.ChatID, bot *Bot) *FSM {
	return &FSM{
		chatID: chatID,
		log:    log.With("chat", chatID),
		bot:    bot,
		state:  initialState,
	}
}

func (fsm *FSM) To(newState State, msg *tg.Message) {
	fsm.log.Infow("State transition", "from", fsm.state, "to", newState)
	fsm.state = newState
	fsm.state.To(fsm, msg, fsm.bot)
}

func (fsm *FSM) Do(msg *tg.Message) error {
	return fsm.state.Do(fsm, msg, fsm.bot)
}

type State interface {
	fmt.Stringer
	To(fsm *FSM, msg *tg.Message, bot *Bot)       // TODO maybe return new state at the end?
	Do(fsm *FSM, msg *tg.Message, bot *Bot) error // TODO maybe return new state at the end?
}

type InitialState struct {
	commands map[string]State
}

var initialState = &InitialState{commands: map[string]State{
	"start":     startCommandState,
	"stop":      stopCommandState,
	"track":     whichLocationState,
	"stoptrack": stopTrackCommandState,
}}

func (s *InitialState) String() string {
	return "InitialState"
}

func (s *InitialState) To(*FSM, *tg.Message, *Bot) {
	panic(errors.New("should not be called")) // TODO or should it?
}

func (s *InitialState) Do(fsm *FSM, msg *tg.Message, bot *Bot) error {
	if msg == nil || !msg.IsCommand() {
		reply := newMessage(fsm.chatID, "Please select a command")
		bot.sendAndForget(reply, fsm.log)
		fsm.To(doneState, msg)
		return nil
	}
	fsm.To(commandHandlingState, msg)
	return nil
}

type CommandHandlingState struct {
	commands map[string]State
}

var commandHandlingState = &CommandHandlingState{commands: map[string]State{
	"start":     startCommandState,
	"stop":      stopCommandState,
	"track":     whichLocationState,
	"stoptrack": stopTrackCommandState,
}}

func (s *CommandHandlingState) String() string {
	return "CommandHandlingState"
}

func (s *CommandHandlingState) To(fsm *FSM, msg *tg.Message, bot *Bot) {
	fsm.log.Infow("New command", "text", msg.Command())
	if state, ok := s.commands[msg.Command()]; ok {
		fsm.To(state, msg)
		return
	}
	reply := newMessage(
		fsm.chatID,
		fmt.Sprintf("No such command %q, please select one of the available commands", msg.Command()),
	)
	bot.sendAndForget(reply, fsm.log)
	fsm.To(doneState, msg) // Just to not store it in memory indefinably
}

func (s *CommandHandlingState) Do(*FSM, *tg.Message, *Bot) error {
	panic(errors.New("should not be called"))
}

type StartCommandState struct {
}

var startCommandState = &StartCommandState{}

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

type StopCommandState struct {
}

func (s StopCommandState) String() string {
	return "StopCommandState"
}

var stopCommandState = &StopCommandState{}

func (s StopCommandState) To(*FSM, *tg.Message, *Bot) {
	db.Users.Decrement()
}

func (s StopCommandState) Do(*FSM, *tg.Message, *Bot) error {
	return errors.New("should never be called")
}

type DoneState struct {
}

func (s DoneState) String() string {
	return "DoneState"
}

var doneState = &DoneState{}

func (s DoneState) To(fsm *FSM, _ *tg.Message, _ *Bot) {
	delete(chatFSMs, fsm.chatID) // TODO this is ugly
	fsm.log.Infow("We are done")
}

func (s DoneState) Do(*FSM, *tg.Message, *Bot) error {
	return errors.New("should never be called")
}

type WhichLocationState struct {
}

func (w *WhichLocationState) String() string {
	return "WhichLocationState"
}

var whichLocationState = &WhichLocationState{}

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
	rows := make([][]tg.KeyboardButton, 0, len(db.LocationToName))
	for _, locationName := range db.LocationToName {
		rows = append(rows, []tg.KeyboardButton{tg.NewKeyboardButton(locationName)})
	}
	return tg.NewOneTimeReplyKeyboard(rows...)
}

type BeforeDateState struct {
}

type HowManyPeopleState struct {
	location string
}

func (h *HowManyPeopleState) String() string {
	return "HowManyPeopleState"
}

func (h *HowManyPeopleState) To(fsm *FSM, msg *tg.Message, bot *Bot) {
	row := make([][]tg.KeyboardButton, 6)
	for i := 0; i < 6; i++ {
		row[i] = []tg.KeyboardButton{tg.NewKeyboardButton(strconv.Itoa(i + 1))}
	}
	toSend := newMessage(fsm.chatID, "How many people?")
	toSend.ReplyMarkup = tg.NewOneTimeReplyKeyboard(row...)
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
		toSend := newMessage(fsm.chatID, "Failed to create subscription. Please try again.")
		if _, err = bot.API.Send(toSend); err != nil {
			fsm.log.Warnw("Failed to send message", "msg", toSend.Text, "err", err)
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

type StopTrackCommandState struct {
}

func (s StopTrackCommandState) String() string {
	return "StopTrackCommandState"
}

var stopTrackCommandState = &StopTrackCommandState{}

func (s StopTrackCommandState) To(fsm *FSM, msg *tg.Message, bot *Bot) {
	for location := range db.LocationToName {
		// TODO this should be improved
		subscriptions, err := db.Subscriptions.GetForLocation(location)
		if err != nil {
			fsm.log.Warnw("Failed to get subscriptions for delete", "location", location, "err", err)
			continue
		}
		for _, subscription := range subscriptions {
			if subscription.ChatID == fsm.chatID {
				if err := db.Subscriptions.RemoveFromLocation(location, subscription); err != nil {
					fsm.log.Warnw("Failed to delete subscription", "subscription", subscription, "err", err)
				} else {
					fsm.log.Debugw("One less follower", "location", location)
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

func newMessage(chatId domain.ChatID, text string) tg.MessageConfig {
	message := tg.NewMessage(int64(chatId), text)
	message.ReplyMarkup = tg.ReplyKeyboardRemove{RemoveKeyboard: true}
	return message
}

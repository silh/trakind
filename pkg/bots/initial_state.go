package bots

import (
	"errors"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type InitialState struct {
	commands map[string]State
}

func (s *InitialState) String() string {
	return "InitialState"
}

func (s *InitialState) To(*FSM, *tg.Message, *Bot) {
	panic(errors.New("should not be called")) // TODO or should it?
}

func (s *InitialState) Do(fsm *FSM, msg *tg.Message, bot *Bot) error {
	if msg == nil || !msg.IsCommand() {
		reply := newMessage(fsm.chatID, "Please select a command")
		bot.SendAndForget(reply, fsm.log)
		fsm.To(doneState, msg)
		return nil
	}
	fsm.To(commandHandlingState, msg)
	return nil
}

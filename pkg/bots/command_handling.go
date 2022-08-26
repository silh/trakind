package bots

import (
	"errors"
	"fmt"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type CommandHandlingState struct {
	commands map[string]State
}

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

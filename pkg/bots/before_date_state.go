package bots

import (
	"fmt"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/silh/trakind/pkg/db"
	"github.com/silh/trakind/pkg/domain"
	"strings"
)

type BeforeDateState struct {
	action      domain.Action
	location    domain.Location
	peopleCount int
}

func (s *BeforeDateState) String() string {
	return "BeforeDateState"
}

func (s *BeforeDateState) To(fsm *FSM, msg *tg.Message, bot *Bot) {
	toSend := newMessage(
		fsm.chatID,
		"Are you interested in time slots before certain date or all? "+
			"Please reply with a date in format YYYY-MM-DD or a word \"all\".",
	)
	if _, err := bot.API.Send(toSend); err != nil {
		fsm.log.Warnw("Failed to send message", "msg", toSend.Text, "err", err)
		fsm.To(doneState, msg)
	}
}

func (s *BeforeDateState) Do(fsm *FSM, msg *tg.Message, bot *Bot) error {
	if msg.IsCommand() {
		fsm.To(commandHandlingState, msg)
		return nil
	}
	subscription := domain.Subscription{
		ChatID:      fsm.chatID,
		PeopleCount: s.peopleCount,
		Action:      s.action.Code,
	}
	var err error
	if !strings.EqualFold(msg.Text, "all") {
		subscription.TrackBefore, err = domain.ParseWindowDate(msg.Text)
		if err != nil {
			fsm.log.Debugw("Could not parse windowDate", "err", err)
			toSend := newMessage(
				fsm.chatID,
				fmt.Sprintf(
					"Incorrect response %q. Please reply with a date in format YYYY-MM-DD or a word \"all\".",
					msg.Text,
				),
			)
			if _, err := bot.API.Send(toSend); err != nil {
				fsm.log.Warnw("Failed to send message", "msg", toSend.Text, "err", err)
				fsm.To(doneState, msg)
			}
			return nil
		}
	}
	// Actually save subscription
	if err := db.Subscriptions.AddToLocation(s.location.Code, subscription); err != nil {
		fsm.log.Warnw("Failed to store subscription", "subscription", subscription, "err", err)
		toSend := newMessage(fsm.chatID, "Failed to create subscription. Please try again.")
		if _, err = bot.API.Send(toSend); err != nil {
			fsm.log.Warnw("Failed to send message", "msg", toSend.Text, "err", err)
		}
		fsm.To(doneState, msg)
		return nil
	}
	s.sendSubscribedNotification(fsm, subscription, bot)
	fsm.log.Infow("One more follower", "location", s.location.Code)
	fsm.To(doneState, msg)
	return nil
}

func (s *BeforeDateState) sendSubscribedNotification(fsm *FSM, subscription domain.Subscription, bot *Bot) {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(
		"You will now get a notification when an open time window found for %s at the location %s for %d people",
		s.action.Name,
		s.location.Name,
		s.peopleCount,
	))
	if (subscription.TrackBefore != domain.Date{}) {
		sb.WriteString(fmt.Sprintf(" before %s", &subscription.TrackBefore))
	}
	sb.WriteRune('.')
	toSend := newMessage(fsm.chatID, sb.String())
	if _, err := bot.API.Send(toSend); err != nil {
		fsm.log.Warnw("Failed to send message", "msg", toSend.Text, "err", err)
	}
}

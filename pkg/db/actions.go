package db

import "github.com/silh/trakind/pkg/domain"

// Actions is a set of all supported actions.
var Actions = map[domain.Action]struct{}{
	domain.DocumentPickup: {},
	domain.Biometrics:     {},
}

// ActionForName returns action by its name.
func ActionForName(name string) (domain.Action, bool) {
	for action := range Actions {
		if action.Name == name {
			return action, true
		}
	}
	return domain.Action{}, false
}

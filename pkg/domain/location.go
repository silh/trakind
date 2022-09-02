package domain

type Location struct {
	Name             string
	Code             string
	AvailableActions map[Action]struct{}
}

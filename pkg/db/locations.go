package db

import (
	"github.com/silh/trakind/pkg/domain"
	"github.com/silh/trakind/pkg/sets"
)

// LocationToChats contain a set of keys which are all possible location for an appointment. Values are sets of
// subscribers for each location. TODO this should be stored on disk.
var LocationToChats = map[string]sets.Set[domain.Subscription]{
	"AM": sets.NewConcurrent[domain.Subscription](),
	"DH": sets.NewConcurrent[domain.Subscription](),
	"ZW": sets.NewConcurrent[domain.Subscription](),
	"DB": sets.NewConcurrent[domain.Subscription](),
}

// LocationToName location code to a proper name.
var LocationToName = map[string]string{
	"AM": "IND Amsterdam",
	"DH": "IND Den Haag",
	"ZW": "IND Zwolle",
	"DB": "IND Den Bosch",
}

package db

import (
	"github.com/silh/trakind/pkg/domain"
	"github.com/silh/trakind/pkg/loggers"
	"strings"
)

var log = loggers.Logger()

const locationsBucket = "locations"

var DocPickupLocations = []domain.Location{
	{Name: "IND Amsterdam", Code: "AM"},
	{Name: "IND Den Haag", Code: "DH"},
	{Name: "IND Zwolle", Code: "ZW"},
	{Name: "IND Den Bosch", Code: "DB"},
}

func LocationForName(name string) (domain.Location, bool) {
	for _, location := range DocPickupLocations {
		if strings.EqualFold(location.Name, name) {
			return location, true
		}
	}
	return domain.Location{}, false
}

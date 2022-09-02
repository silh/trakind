package db

import (
	"github.com/silh/trakind/pkg/domain"
	"github.com/silh/trakind/pkg/loggers"
	"strings"
)

var log = loggers.Logger()

const locationsBucket = "locations"

var allActionsSet = map[domain.Action]struct{}{
	domain.DocumentPickup: {},
	domain.Biometrics:     {},
}
var onlyBioSet = map[domain.Action]struct{}{
	domain.Biometrics: {},
}

var Locations = []domain.Location{
	{Name: "IND Amsterdam", Code: "AM", AvailableActions: allActionsSet},
	{Name: "IND Den Haag", Code: "DH", AvailableActions: allActionsSet},
	{Name: "IND Zwolle", Code: "ZW", AvailableActions: allActionsSet},
	{Name: "IND Den Bosch", Code: "DB", AvailableActions: allActionsSet},
	{Name: "IND Haarlem", Code: "6b425ff9f87de136a36b813cccf26e23", AvailableActions: onlyBioSet},
	{Name: "Expatcenter Groningen", Code: "0c127eb6d9fe1ced413d2112305e75f6", AvailableActions: onlyBioSet},
	{Name: "Expatcenter Maastricht", Code: "6c5280823686521552efe85094e607cf", AvailableActions: onlyBioSet},
	{Name: "Expatcenter Wageningen", Code: "b084907207cfeea941cd9698821fd894", AvailableActions: onlyBioSet},
	{Name: "Expatcenter Eindhoven", Code: "0588ef4088c08f53294eb60bab55c81e", AvailableActions: onlyBioSet},
	{Name: "Expatcenter Den Haag", Code: "5e325f444aeb56bb0270a61b4a0403eb", AvailableActions: onlyBioSet},
	{Name: "Expatcenter Rotterdam", Code: "f0ef3c8f0973875936329d713a68c5f3", AvailableActions: onlyBioSet},
	{Name: "Expatcenter Enschede", Code: "3535aca0fb9a2e8e8015f768fb3fa69d", AvailableActions: onlyBioSet},
	{Name: "Expatcenter Utrecht", Code: "fa24ccf0acbc76a7793765937eaee440", AvailableActions: onlyBioSet},
	{Name: "Expatcenter Amsterdam", Code: "284b189314071dcd571df5bb262a31db", AvailableActions: onlyBioSet},
}

func LocationForName(name string) (domain.Location, bool) {
	for _, location := range Locations {
		if strings.EqualFold(location.Name, name) {
			return location, true
		}
	}
	return domain.Location{}, false
}

func LocationsForAction(action domain.Action) []domain.Location {
	locations := make([]domain.Location, 0)
	for _, location := range Locations {
		if _, ok := location.AvailableActions[action]; ok {
			locations = append(locations, location)
		}
	}
	return locations
}

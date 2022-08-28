package db

import (
	"github.com/silh/trakind/pkg/loggers"
)

var log = loggers.Logger()

const locationsBucket = "locations"

type Location struct {
	Name string
	Code string
}

var Locations = []Location{
	{Name: "IND Amsterdam", Code: "AM"},
	{Name: "IND Den Haag", Code: "DH"},
	{Name: "IND Zwolle", Code: "ZW"},
	{Name: "IND Den Bosch", Code: "DB"},
}

// LocationToName location code to a proper name.
var LocationToName = map[string]string{}

// NameToLocation reverse of LocationToName.
var NameToLocation = map[string]string{}

func init() {
	for _, location := range Locations {
		LocationToName[location.Code] = location.Name
		NameToLocation[location.Name] = location.Code
	}
}

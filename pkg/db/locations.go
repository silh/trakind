package db

import (
	"encoding/json"
	"github.com/silh/trakind/pkg/domain"
	"github.com/silh/trakind/pkg/loggers"
	"github.com/xujiajun/nutsdb"
)

var log = loggers.Logger()

const locationsBucket = "locations"

var db *nutsdb.DB

var Subs *SubscriptionsDB

func init() {
	opt := nutsdb.DefaultOptions
	opt.Dir = "./db"
	var err error
	db, err = nutsdb.Open(opt)
	if err != nil {
		log.Fatal("Failed to open DB", "err", err)
	}
	Subs = &SubscriptionsDB{storage: db}
}

type SubscriptionsDB struct {
	storage *nutsdb.DB
}

func (db *SubscriptionsDB) AddToLocation(location string, subscription domain.Subscription) error {
	return db.storage.Update(func(tx *nutsdb.Tx) error {
		data, err := json.Marshal(&subscription)
		if err != nil {
			return err
		}
		return tx.SAdd(locationsBucket, []byte(location), data)
	})
}

func (db *SubscriptionsDB) RemoveFromLocation(location string, subscription domain.Subscription) error {
	return db.storage.Update(func(tx *nutsdb.Tx) error {
		value, err := json.Marshal(&subscription)
		if err != nil {
			return err
		}
		return tx.SRem(locationsBucket, []byte(location), value)
	})
}

// GetForLocation returns a list of subscriptions for given location.
// We don't expect that many of them, should be fine keeping all in-memory.
func (db *SubscriptionsDB) GetForLocation(location string) ([]domain.Subscription, error) {
	var result []domain.Subscription
	return result, db.storage.View(func(tx *nutsdb.Tx) error {
		locationKey := []byte(location)
		if ok, err := tx.SHasKey(locationsBucket, locationKey); !ok {
			return err // might be nil, we don't really care
		}
		list, err := tx.SMembers(locationsBucket, locationKey)
		if err != nil {
			return err
		}
		result = make([]domain.Subscription, len(list))
		var subscription domain.Subscription
		for i, item := range list {
			if err := json.Unmarshal(item, &subscription); err != nil {
				return err
			}
			result[i] = subscription
		}
		return nil
	})
}

// LocationToName location code to a proper name.
var LocationToName = map[string]string{
	"AM": "IND Amsterdam",
	"DH": "IND Den Haag",
	"ZW": "IND Zwolle",
	"DB": "IND Den Bosch",
}
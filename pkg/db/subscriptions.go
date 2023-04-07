package db

import (
	"encoding/json"
	"github.com/silh/trakind/pkg/domain"
	"github.com/xujiajun/nutsdb"
)

var Subscriptions *SubscriptionsDB

func init() {
	Subscriptions = &SubscriptionsDB{storage: db}
}

type SubscriptionsDB struct {
	storage *nutsdb.DB
}

func (db *SubscriptionsDB) AddToLocation(locationCode string, subscription domain.Subscription) error {
	return db.storage.Update(func(tx *nutsdb.Tx) error {
		data, err := json.Marshal(&subscription)
		if err != nil {
			return err
		}
		return tx.SAdd(locationsBucket, []byte(locationCode), data)
	})
}

func (db *SubscriptionsDB) RemoveFromLocation(locationCode string, subscription domain.Subscription) error {
	return db.storage.Update(func(tx *nutsdb.Tx) error {
		value, err := json.Marshal(&subscription)
		if err != nil {
			return err
		}
		return tx.SRem(locationsBucket, []byte(locationCode), value)
	})
}

// GetForLocation returns a list of subscriptions for given location.
// We don't expect that many of them, should be fine keeping all in-memory.
func (db *SubscriptionsDB) GetForLocation(locationCode string) ([]domain.Subscription, error) {
	var result []domain.Subscription
	return result, db.storage.View(func(tx *nutsdb.Tx) error {
		locationKey := []byte(locationCode)
		ok, err := tx.SHasKey(locationsBucket, locationKey)
		if !ok || err == nutsdb.ErrBucketNotFound {
			return nil
		}
		if err != nil {
			return err
		}
		list, err := tx.SMembers(locationsBucket, locationKey)
		if err != nil {
			return err
		}
		result = make([]domain.Subscription, len(list))
		for i, item := range list {
			var subscription domain.Subscription
			if err := json.Unmarshal(item, &subscription); err != nil {
				return err
			}
			result[i] = subscription
		}
		return nil
	})
}

// CountForLocation returns number of subscribers for a location.
func (db *SubscriptionsDB) CountForLocation(locationCode string) (int, error) {
	var result int
	return result, db.storage.View(func(tx *nutsdb.Tx) error {
		locationKey := []byte(locationCode)
		ok, err := tx.SHasKey(locationsBucket, locationKey)
		if !ok || err == nutsdb.ErrBucketNotFound {
			return nil
		}
		if err != nil {
			return err
		}
		result, err = tx.SCard(locationsBucket, locationKey)
		return err
	})
}

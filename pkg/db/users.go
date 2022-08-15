package db

import (
	"errors"
	"github.com/xujiajun/nutsdb"
	"strconv"
)

const TTLInfinite = 0
const usersBucket = "users"

var usersCounterKey = []byte("usersCounter")

var Users *UsersCounterDB

func init() {
	Users = &UsersCounterDB{storage: db}
}

type UsersCounterDB struct {
	storage *nutsdb.DB
}

func (db *UsersCounterDB) Increment() {
	db.add(1)
}

func (db *UsersCounterDB) Decrement() {
	db.add(-1)
}

func (db *UsersCounterDB) add(value int64) {
	err := db.storage.Update(func(tx *nutsdb.Tx) error {
		entry, err := tx.Get(usersBucket, usersCounterKey)
		if err != nil &&
			(!errors.Is(err, nutsdb.ErrBucketNotFound) &&
				!errors.Is(err, nutsdb.ErrKeyNotFound)) {
			log.Errorw("Users counter checked", "err", err)
			return nil
		}
		counter, err := strconv.ParseInt(string(entry.Value), 10, 0)
		if err != nil {
			log.Errorw("Users counter cannot be parsed", "err", err)
			return nil
		}
		if counter > 0 { // we have
			counter += value
		}
		if err = tx.Put(usersBucket, usersCounterKey, []byte(strconv.FormatInt(counter, 10)), TTLInfinite); err != nil {
			log.Errorw("Users counter be updated", "err", err)
			return nil
		}
		log.Info("New user", "count", counter) // TODO this should be moved to bot? or command?
		return nil
	})
	log.Errorw("Unexpected adjusting users counter error", "value", value, "err", err)
}

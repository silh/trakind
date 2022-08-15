package db

import "github.com/xujiajun/nutsdb"

var db *nutsdb.DB

func init() {
	opt := nutsdb.DefaultOptions
	opt.Dir = "./db"
	var err error
	if db, err = nutsdb.Open(opt); err != nil {
		log.Fatal("Failed to open DB", "err", err)
	}
}

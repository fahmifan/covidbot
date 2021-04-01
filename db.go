package covidbot

import (
	"time"

	"github.com/sirupsen/logrus"
	bolt "go.etcd.io/bbolt"
)

// MustOpenBoltDB open boltdb database and fatal on error
func MustOpenBoltDB() *bolt.DB {
	boltDB, err := bolt.Open("./covidbot.bolt.db", 0600, &bolt.Options{
		Timeout: time.Second * 30,
	})
	if err != nil {
		logrus.Fatal(err)
	}
	return boltDB
}

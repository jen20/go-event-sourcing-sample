package bbolt_test

import (
	"os"
	"testing"

	"github.com/hallgren/eventsourcing/base"
	"github.com/hallgren/eventsourcing/base/suite"
	"github.com/hallgren/eventsourcing/eventstore/bbolt"
)

func TestSuite(t *testing.T) {
	f := func() (base.EventStore, func(), error) {
		dbFile := "bolt.db"
		es := bbolt.MustOpenBBolt(dbFile)
		return es, func() {
			es.Close()
			os.Remove(dbFile)
		}, nil
	}
	suite.Test(t, f)
}

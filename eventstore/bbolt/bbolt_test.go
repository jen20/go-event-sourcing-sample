package bbolt_test

import (
	"os"
	"testing"

	"github.com/hallgren/eventsourcing/base"
	eventstore "github.com/hallgren/eventsourcing/eventstore"
	"github.com/hallgren/eventsourcing/eventstore/bbolt"
	"github.com/hallgren/eventsourcing/eventstore/suite"
)

func TestSuite(t *testing.T) {
	f := func(ser eventstore.Serializer) (base.EventStore, func(), error) {
		dbFile := "bolt.db"
		es := bbolt.MustOpenBBolt(dbFile, ser)
		return es, func() {
			es.Close()
			os.Remove(dbFile)
		}, nil
	}
	suite.Test(t, f)
}

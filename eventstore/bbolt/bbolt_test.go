package bbolt_test

import (
	"os"
	"testing"

	eventstore "github.com/hallgren/eventsourcing/eventstore"
	"github.com/hallgren/eventsourcing/eventstore/bbolt"
	"github.com/hallgren/eventsourcing/eventstore/suite"
)

func TestSuite(t *testing.T) {
	f := func(ser eventstore.Serializer) (eventstore.EventStore, func(), error) {
		dbFile := "bolt.db"
		es := bbolt.MustOpenBBolt(dbFile, ser)
		return es, func() {
			es.Close()
			os.Remove(dbFile)
		}, nil
	}
	suite.Test(t, f)
}

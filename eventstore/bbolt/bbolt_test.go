package bbolt_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/hallgren/eventsourcing"
	"github.com/hallgren/eventsourcing/eventstore/bbolt"
	"github.com/hallgren/eventsourcing/eventstore/suite"
)

func TestSuite(t *testing.T) {
	f := func() (eventsourcing.EventStore, func(), error) {
		dbFile := "bolt.db"
		ser := eventsourcing.NewSerializer(json.Marshal, json.Unmarshal)

		ser.Register(&suite.FrequentFlierAccount{},
			ser.Events(
				&suite.FrequentFlierAccountCreated{},
				&suite.FlightTaken{},
				&suite.StatusMatched{},
			),
		)
		es := bbolt.MustOpenBBolt(dbFile, *ser)
		return es, func() {
			es.Close()
			os.Remove(dbFile)
		}, nil
	}
	suite.Test(t, f)
}

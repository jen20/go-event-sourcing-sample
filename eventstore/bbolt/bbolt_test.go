package bbolt_test

import (
	"encoding/json"
	"github.com/hallgren/eventsourcing"
	"github.com/hallgren/eventsourcing/eventstore/bbolt"
	"github.com/hallgren/eventsourcing/eventstore/suite"
	"os"
	"testing"
)

func TestSuite(t *testing.T) {
	f := func() (eventsourcing.EventStore, func(), error) {
		dbFile := "bolt.db"
		ser := eventsourcing.NewSerializer(json.Marshal, json.Unmarshal)

		ser.RegisterTypes(&suite.FrequentFlierAccount{},
			func() interface{} { return &suite.FrequentFlierAccountCreated{}},
			func() interface{} { return &suite.FlightTaken{}},
			func() interface{} { return &suite.StatusMatched{}},
		)
		es := bbolt.MustOpenBBolt(dbFile, *ser)
		return es, func(){
			es.Close()
			os.Remove(dbFile)
		}, nil
	}
	suite.Test(t, f)
}

package esdb_test

import (
	"encoding/json"
	"testing"

	"github.com/EventStore/EventStore-Client-Go/esdb"
	"github.com/hallgren/eventsourcing"
	es "github.com/hallgren/eventsourcing/eventstore/esdb"
	"github.com/hallgren/eventsourcing/eventstore/suite"
)

func TestSuite(t *testing.T) {
	f := func() (eventsourcing.EventStore, func(), error) {
		// region createClient
		settings, err := esdb.ParseConnectionString("esdb://localhost:2113?tls=false")
		if err != nil {
			return nil, nil, err
		}

		db, err := esdb.NewClient(settings)
		if err != nil {
			return nil, nil, err
		}

		ser := eventsourcing.NewSerializer(json.Marshal, json.Unmarshal)

		ser.Register(&suite.FrequentFlierAccount{},
			ser.Events(
				&suite.FrequentFlierAccountCreated{},
				&suite.FlightTaken{},
				&suite.StatusMatched{},
			),
		)

		es := es.Open(db, *ser)
		return es, func() {
		}, nil
	}
	suite.Test(t, f)
}

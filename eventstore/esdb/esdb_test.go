//go:build manual
// +build manual

// make these tests manual as they are dependent on a running event store db.

package esdb_test

import (
	"testing"

	"github.com/EventStore/EventStore-Client-Go/v3/esdb"

	"github.com/hallgren/eventsourcing/core"
	"github.com/hallgren/eventsourcing/core/testsuite"
	es "github.com/hallgren/eventsourcing/eventstore/esdb"
)

func TestSuite(t *testing.T) {
	f := func() (core.EventStore, func(), error) {
		// region createClient
		settings, err := esdb.ParseConnectionString("esdb://localhost:2113?tls=false")
		if err != nil {
			return nil, nil, err
		}

		db, err := esdb.NewClient(settings)
		if err != nil {
			return nil, nil, err
		}

		es := es.Open(db, true)
		return es, func() {
		}, nil
	}
	testsuite.Test(t, f)
}

package memory_test

import (
	"testing"

	eventstore "github.com/hallgren/eventsourcing/eventstore"
	"github.com/hallgren/eventsourcing/eventstore/memory"
	"github.com/hallgren/eventsourcing/eventstore/suite"
)

func TestSuite(t *testing.T) {
	f := func(ser eventstore.Serializer) (eventstore.EventStore, func(), error) {
		es := memory.Create()
		return es, func() { es.Close() }, nil
	}
	suite.Test(t, f)
}

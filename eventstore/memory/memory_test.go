package memory_test

import (
	"testing"

	"github.com/hallgren/eventsourcing"
	"github.com/hallgren/eventsourcing/eventstore/memory"
	"github.com/hallgren/eventsourcing/eventstore/suite"
)

func TestSuite(t *testing.T) {
	f := func(ser eventsourcing.Serializer) (eventsourcing.EventStore, func(), error) {
		es := memory.Create()
		return es, func() { es.Close() }, nil
	}
	suite.Test(t, f)
}

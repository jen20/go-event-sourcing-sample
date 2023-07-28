package memory_test

import (
	"testing"

	"github.com/hallgren/eventsourcing/base"
	"github.com/hallgren/eventsourcing/base/suite"
	"github.com/hallgren/eventsourcing/eventstore/memory"
)

func TestSuite(t *testing.T) {
	f := func() (base.EventStore, func(), error) {
		es := memory.Create()
		return es, func() { es.Close() }, nil
	}
	suite.Test(t, f)
}

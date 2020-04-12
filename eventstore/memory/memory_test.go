package memory_test

import (
	"github.com/hallgren/eventsourcing/eventstore/memory"
	"github.com/hallgren/eventsourcing/eventstore/suite"
	"github.com/hallgren/eventsourcing/serializer/unsafe"
	"testing"
)

func TestSuite(t *testing.T) {
	f := func() (suite.Eventstore, func(), error) {
		es := memory.Create(unsafe.New())
		return es, func(){es.Close()}, nil
	}
	suite.Test(t, f)
}

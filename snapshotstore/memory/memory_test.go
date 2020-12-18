package memory_test

import (
	"encoding/xml"
	"testing"

	"github.com/hallgren/eventsourcing"
	"github.com/hallgren/eventsourcing/snapshotstore/memory"
	"github.com/hallgren/eventsourcing/snapshotstore/suite"
)

func TestMemorySnapshot(t *testing.T) {
	f := func() (suite.SnapshotStore, func(), error) {
		store := memory.New(*eventsourcing.NewSerializer(xml.Marshal, xml.Unmarshal))
		return store, func() {}, nil
	}
	suite.Test(t, f)
}

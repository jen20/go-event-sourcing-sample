package memory_test

import (
	"testing"

	"github.com/hallgren/eventsourcing"
	"github.com/hallgren/eventsourcing/snapshotstore/memory"
	"github.com/hallgren/eventsourcing/snapshotstore/suite"
)

type provider struct{}

func (p *provider) Setup() (eventsourcing.SnapshotStore, error) {
	return memory.New(), nil
}

func (p *provider) Cleanup() {}

func (p *provider) Teardown() {}

func TestMemorySnapshot(t *testing.T) {
	suite.Test(t, new(provider))
}

package memory

import (
	"context"

	"github.com/hallgren/eventsourcing/core"
)

type Memory struct {
	snapshots map[string]core.Snapshot
}

// Create in memory snapshot store
func Create() *Memory {
	return &Memory{
		snapshots: make(map[string]core.Snapshot),
	}
}

func (m *Memory) Get(ctx context.Context, aggregateID, aggregateType string) (core.Snapshot, error) {
	snapshot, ok := m.snapshots[aggregateID+"_"+aggregateType]
	if !ok {
		return core.Snapshot{}, core.ErrSnapshotNotFound
	}
	return snapshot, nil
}

func (m *Memory) Save(aggregateID, aggregateType string, snapshot core.Snapshot) error {
	m.snapshots[aggregateID+"_"+aggregateType] = snapshot
	return nil
}

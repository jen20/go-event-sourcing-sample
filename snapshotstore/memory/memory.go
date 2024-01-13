package memory

import (
	"context"
	"fmt"
)

type Memory struct {
	snapshots map[string][]byte
}

// Create in memory snapshot store
func Create() *Memory {
	return &Memory{
		snapshots: make(map[string][]byte),
	}
}

func (m *Memory) Get(ctx context.Context, aggregateID, aggregateType string) ([]byte, error) {
	snapshot, ok := m.snapshots[aggregateID+"_"+aggregateType]
	if !ok {
		return nil, fmt.Errorf("could not find snapshot")
	}
	return snapshot, nil
}

func (m *Memory) Save(aggregateID, aggregateType string, snapshot []byte) error {
	m.snapshots[aggregateID+"_"+aggregateType] = snapshot
	return nil
}

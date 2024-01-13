package core

import "context"

// SnapshotStore interface expose the methods an snapshot store must uphold
type SnapshotStore interface {
	Save(id, aggregateType string, snapshot []byte) error
	Get(ctx context.Context, id, aggregateType string) ([]byte, error)
}

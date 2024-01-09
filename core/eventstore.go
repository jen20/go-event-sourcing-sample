package core

import (
	"context"
	"errors"
)

// ErrConcurrency when the currently saved version of the aggregate differs from the new ones
var ErrConcurrency = errors.New("concurrency error")

// Iterator is the interface an event store Get needs to return
type Iterator interface {
	Next() bool
	Value() (Event, error)
	Close()
}

// EventStore interface expose the methods an event store must uphold
type EventStore interface {
	Save(events []Event) error
	Get(ctx context.Context, id string, aggregateType string, afterVersion Version) (Iterator, error)
}

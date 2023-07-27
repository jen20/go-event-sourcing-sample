package base

import (
	"context"
	"errors"
)

// ErrNoEvents when there is no events to get
var ErrNoEvents = errors.New("no events")

// ErrNoMoreEvents when iterator has no more events to deliver
var ErrNoMoreEvents = errors.New("no more events")

// Iterator is the interface an event store Get needs to return
type Iterator interface {
	Next() (Event, error)
	Close()
}

// EventStore interface expose the methods an event store must uphold
type EventStore interface {
	Save(events []Event) error
	Get(ctx context.Context, id string, aggregateType string, afterVersion Version) (Iterator, error)
}

package eventsourcing

import (
	"context"

	"github.com/hallgren/eventsourcing/base"
)

// EventStore interface expose the methods an event store must uphold
type EventStore interface {
	Save(events []base.Event) error
	Get(ctx context.Context, id string, aggregateType string, afterVersion base.Version) (base.EventIterator, error)
}

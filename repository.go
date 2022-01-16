package eventsourcing

import (
	"context"
	"errors"
	"reflect"
)

// EventIterator is the interface an event store Get needs to return
type EventIterator interface {
	Next() (Event, error)
	Close()
}

// EventStore interface expose the methods an event store must uphold
type EventStore interface {
	Save(events []Event) error
	Get(id string, aggregateType string, afterVersion Version) (EventIterator, error)
}

// SnapshotStore interface expose the methods an snapshot store must uphold
type SnapshotStore interface {
	Get(id, typ string) (Snapshot, error)
	Save(s Snapshot) error
}

// Aggregate interface to use the aggregate root specific methods
type Aggregate interface {
	Root() *AggregateRoot
	Transition(event Event)
}

// ErrSnapshotNotFound returns if snapshot not found
var ErrSnapshotNotFound = errors.New("snapshot not found")

// ErrAggregateNotFound returns if snapshot or event not found for aggregate
var ErrAggregateNotFound = errors.New("aggregate not found")

// Repository is the returned instance from the factory function
type Repository struct {
	*EventStream
	eventStore EventStore
	snapshot   *SnapshotHandler
}

// NewRepository factory function
func NewRepository(eventStore EventStore, snapshot *SnapshotHandler) *Repository {
	return &Repository{
		eventStore:  eventStore,
		snapshot:    snapshot,
		EventStream: NewEventStream(),
	}
}

// Save an aggregates events
func (r *Repository) Save(aggregate Aggregate) error {
	root := aggregate.Root()
	// use underlaying event slice to set GlobalVersion
	err := r.eventStore.Save(root.aggregateEvents)
	if err != nil {
		return err
	}
	// publish the saved events to subscribers
	r.Update(*root, root.Events())

	// update the internal aggregate state
	root.update()
	return nil
}

// SaveSnapshot saves the current state of the aggregate but only if it has no unsaved events
func (r *Repository) SaveSnapshot(aggregate Aggregate) error {
	if r.snapshot == nil {
		return errors.New("no snapshot store has been initialized")
	}
	return r.snapshot.Save(aggregate)
}

// GetWithContext fetches the aggregates event and build up the aggregate
// If there is a snapshot store try fetch a snapshot of the aggregate and fetch event after the
// version of the aggregate if any
// The event fetching can be cancled from the outside.
func (r *Repository) GetWithContext(ctx context.Context, id string, aggregate Aggregate) error {
	if reflect.ValueOf(aggregate).Kind() != reflect.Ptr {
		return errors.New("aggregate needs to be a pointer")
	}
	// if there is a snapshot store try fetch aggregate snapshot
	if r.snapshot != nil {
		err := r.snapshot.Get(id, aggregate)
		if err != nil && !errors.Is(err, ErrSnapshotNotFound) {
			return err
		}
	}
	root := aggregate.Root()
	aggregateType := reflect.TypeOf(aggregate).Elem().Name()
	// fetch events after the current version of the aggregate that could be fetched from the snapshot store
	eventIterator, err := r.eventStore.Get(id, aggregateType, root.Version())
	if err != nil && !errors.Is(err, ErrNoEvents) {
		return err
	} else if errors.Is(err, ErrNoEvents) && root.Version() == 0 {
		// no events and no snapshot
		return ErrAggregateNotFound
	}
	defer eventIterator.Close()
DONE:
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			event, err := eventIterator.Next()
			if err != nil && !errors.Is(err, ErrNoMoreEvents) {
				return err
			} else if errors.Is(err, ErrNoMoreEvents) && root.Version() == 0 {
				// no events and no snapshot (some eventstore will not return the error ErrNoEvent on Get())
				return ErrAggregateNotFound
			} else if errors.Is(err, ErrNoMoreEvents) {
				break DONE
			}
			// apply the event on the aggregate
			root.BuildFromHistory(aggregate, []Event{event})
		}
	}
	return nil
}

// Get fetches the aggregates event and build up the aggregate
// If there is a snapshot store try fetch a snapshot of the aggregate and fetch event after the
// version of the aggregate if any
func (r *Repository) Get(id string, aggregate Aggregate) error {
	return r.GetWithContext(context.Background(), id, aggregate)
}

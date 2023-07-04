package eventsourcing

import (
	"context"
	"errors"
	"reflect"

	"github.com/hallgren/eventsourcing/base"
)

// Aggregate interface to use the aggregate root specific methods
type Aggregate interface {
	Root() *AggregateRoot
	Transition(event Event)
}

type EventSubscribers interface {
	All(f func(e Event)) *subscription
	AggregateID(f func(e Event), aggregates ...Aggregate) *subscription
	Aggregate(f func(e Event), aggregates ...Aggregate) *subscription
	Event(f func(e Event), events ...interface{}) *subscription
	Name(f func(e Event), aggregate string, events ...string) *subscription
}

// ErrAggregateNotFound returns if snapshot or event not found for aggregate
var ErrAggregateNotFound = errors.New("aggregate not found")

// Repository is the returned instance from the factory function
type Repository struct {
	eventStream *EventStream
	eventStore  base.EventStore
}

// NewRepository factory function
func NewRepository(eventStore base.EventStore) *Repository {
	return &Repository{
		eventStore:  eventStore,
		eventStream: NewEventStream(),
	}
}

// Subscribers returns an interface with all event subscribers
func (r *Repository) Subscribers() EventSubscribers {
	return r.eventStream
}

// Save an aggregates events
func (r *Repository) Save(aggregate Aggregate) error {
	root := aggregate.Root()
	// use under laying event slice to set GlobalVersion
	err := r.eventStore.Save(root.aggregateEvents)
	if err != nil {
		return err
	}
	// publish the saved events to subscribers
	r.eventStream.Publish(*root, root.Events())

	// update the internal aggregate state
	root.update()
	return nil
}

// GetWithContext fetches the aggregates event and build up the aggregate based on it's current version.
// The event fetching can be canceled from the outside.
func (r *Repository) GetWithContext(ctx context.Context, id string, aggregate Aggregate) error {
	if reflect.ValueOf(aggregate).Kind() != reflect.Ptr {
		return errors.New("aggregate needs to be a pointer")
	}

	root := aggregate.Root()
	aggregateType := reflect.TypeOf(aggregate).Elem().Name()
	// fetch events after the current version of the aggregate that could be fetched from the snapshot store
	eventIterator, err := r.eventStore.Get(ctx, id, aggregateType, root.aggregateVersion)
	if err != nil && !errors.Is(err, base.ErrNoEvents) {
		return err
	} else if errors.Is(err, base.ErrNoEvents) && root.Version() == 0 {
		// no events and no snapshot
		return ErrAggregateNotFound
	} else if ctx.Err() != nil {
		return ctx.Err()
	}
	defer eventIterator.Close()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			event, err := eventIterator.Next()
			if err != nil && !errors.Is(err, base.ErrNoMoreEvents) {
				return err
			} else if errors.Is(err, base.ErrNoMoreEvents) && root.Version() == 0 {
				// no events and no snapshot (some eventstore will not return the error ErrNoEvent on Get())
				return ErrAggregateNotFound
			} else if errors.Is(err, base.ErrNoMoreEvents) {
				return nil
			}
			// apply the event on the aggregate
			root.BuildFromHistory(aggregate, []Event{EventConvert(event)})
		}
	}
}

// Get fetches the aggregates event and build up the aggregate
// If there is a snapshot store try fetch a snapshot of the aggregate and fetch event after the
// version of the aggregate if any
func (r *Repository) Get(id string, aggregate Aggregate) error {
	return r.GetWithContext(context.Background(), id, aggregate)
}

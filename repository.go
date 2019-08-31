package eventsourcing

import (
	"github.com/imkira/go-observer"
	"reflect"
)

// EventStore interface expose the methods an event store must uphold
type EventStore interface {
	Save(events []Event) error
	Get(id string, aggregateType string) ([]Event, error)
	EventStream() observer.Stream
}

// AggregateRooter interface to use the aggregate root specific methods
type AggregateRooter interface {
	Changes() []Event
	BuildFromHistory(events []Event)
	Parent() aggregate
	Transition(event Event)
	SetParent(a aggregate)
}

// Repository is the returned instance from the factory function
type Repository struct {
	eventStore EventStore
}

// NewRepository factory function
func NewRepository(eventStore EventStore) *Repository {
	return &Repository{
		eventStore: eventStore,
	}
}

// Save an aggregates events
func (r *Repository) Save(aggregate AggregateRooter) error {
	return r.eventStore.Save(aggregate.Changes())
}

// Get fetches the aggregates event and build up the aggregate
func (r *Repository) Get(id string, aggregate AggregateRooter) error {
	InitAggregate(aggregate)
	// TODO error handle the incoming aggregate to make sure its a pointer
	aggregateType := reflect.TypeOf(aggregate.Parent()).Elem().Name()
	events, err := r.eventStore.Get(id, aggregateType)
	if err != nil {
		return err
	}
	aggregate.BuildFromHistory(events)
	return nil
}

// EventStream returns a stream with all saved events
func (r *Repository) EventStream() observer.Stream {
	return r.eventStore.EventStream()
}
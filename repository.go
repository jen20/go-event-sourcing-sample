package eventsourcing

import (
	"fmt"
	"github.com/imkira/go-observer"
	"reflect"
)

// eventStore interface expose the methods an event store must uphold
type eventStore interface {
	Save(events []Event) error
	Get(id string, aggregateType string) ([]Event, error)
	EventStream() observer.Stream
}

// aggregateRooter interface to use the aggregate root specific methods
type aggregateRooter interface {
	Changes() []Event
	BuildFromHistory(a aggregate, events []Event)
	Transition(event Event)
}

// Repository is the returned instance from the factory function
type Repository struct {
	eventStore eventStore
}

// NewRepository factory function
func NewRepository(eventStore eventStore) *Repository {
	return &Repository{
		eventStore: eventStore,
	}
}

// Save an aggregates events
func (r *Repository) Save(aggregate aggregateRooter) error {
	return r.eventStore.Save(aggregate.Changes())
}

// Get fetches the aggregates event and build up the aggregate
func (r *Repository) Get(id string, aggregate aggregateRooter) error {
	if reflect.ValueOf(aggregate).Kind() != reflect.Ptr {
		return fmt.Errorf("aggregate needs to be a pointer")
	}
	aggregateType := reflect.TypeOf(aggregate).Elem().Name()
	events, err := r.eventStore.Get(id, aggregateType)
	if err != nil {
		return err
	}
	aggregate.BuildFromHistory(aggregate, events)
	return nil
}

// EventStream returns a stream with all saved events
func (r *Repository) EventStream() observer.Stream {
	return r.eventStore.EventStream()
}

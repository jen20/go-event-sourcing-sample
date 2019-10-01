package eventsourcing

import (
	"fmt"
	"github.com/imkira/go-observer"
	"reflect"
)

// eventStore interface expose the methods an event store must uphold
type eventStore interface {
	Save(events []Event) error
	Get(id string, aggregateType string, afterVersion Version) ([]Event, error)
	EventStream() observer.Stream
}

type snapshotStore interface {
	Get(id AggregateRootID, a interface{}) error
	Save(id AggregateRootID, a interface{}) error
}


// aggregate interface to use the aggregate root specific methods
type aggregate interface {
	BuildFromHistory(a aggregate, events []Event)
	Transition(event Event)
	changes() []Event
	updateVersion()
	version() Version
}

// Repository is the returned instance from the factory function
type Repository struct {
	eventStore eventStore
	snapshotStore snapshotStore
}

// NewRepository factory function
func NewRepository(eventStore eventStore, snapshotStore snapshotStore) *Repository {
	return &Repository{
		eventStore: eventStore,
		snapshotStore: snapshotStore,
	}
}

// Save an aggregates events
func (r *Repository) Save(aggregate aggregate) error {
	err := r.eventStore.Save(aggregate.changes())
	if err != nil {
		return err
	}
	aggregate.updateVersion()
	return nil
}

// Get fetches the aggregates event and build up the aggregate
func (r *Repository) Get(id string, aggregate aggregate) error {
	if reflect.ValueOf(aggregate).Kind() != reflect.Ptr {
		return fmt.Errorf("aggregate needs to be a pointer")
	}
	aggregateType := reflect.TypeOf(aggregate).Elem().Name()
	if r.snapshotStore != nil {
		err := r.snapshotStore.Get(AggregateRootID(id), aggregate)
		if err != nil {
			fmt.Println("Could not find snapshot")
		}
	}

	events, err := r.eventStore.Get(id, aggregateType, aggregate.version())
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

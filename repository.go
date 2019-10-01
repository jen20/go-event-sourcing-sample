package eventsourcing

import (
	"fmt"
	"github.com/hallgren/eventsourcing/snapshotstore"
	"github.com/imkira/go-observer"
	"reflect"
)

// eventStore interface expose the methods an event store must uphold
type eventStore interface {
	Save(events []Event) error
	Get(id string, aggregateType string, afterVersion Version) ([]Event, error)
}

type snapshotStore interface {
	Get(id string, a interface{}) error
	Save(id string, a interface{}) error
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
	events observer.Property  // A property to which all event changes for all event types are published

}

// NewRepository factory function
func NewRepository(eventStore eventStore, snapshotStore snapshotStore) *Repository {
	return &Repository{
		eventStore: eventStore,
		snapshotStore: snapshotStore,
		events: observer.NewProperty(nil),
	}
}

// Save an aggregates events
func (r *Repository) Save(aggregate aggregate) error {
	err := r.eventStore.Save(aggregate.changes())
	if err != nil {
		return err
	}
	// publish the saved events to the events stream
	for _, event := range aggregate.changes() {
		r.events.Update(event)
	}
	// aggregate are saved to the event store now its safe to update the internal aggregate state
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
		err := r.snapshotStore.Get(id, aggregate)
		if err != nil && err != snapshotstore.SnapshotNotFoundError {
			return err
		}
	}

	events, err := r.eventStore.Get(id, aggregateType, aggregate.version())
	if err != nil {
		return err
	}
	aggregate.BuildFromHistory(aggregate, events)
	return nil
}

// EventStream returns a stream where all saved event will be published
func (r *Repository) EventStream() observer.Stream {
	return r.events.Observe()
}

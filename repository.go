package eventsourcing

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"

	"github.com/hallgren/eventsourcing/core"
)

// Aggregate interface to use the aggregate root specific methods
type aggregate interface {
	Root() *AggregateRoot
	Transition(event Event)
	Register(RegisterFunc)
}

type EventSubscribers interface {
	All(f func(e Event)) *subscription
	AggregateID(f func(e Event), aggregates ...aggregate) *subscription
	Aggregate(f func(e Event), aggregates ...aggregate) *subscription
	Event(f func(e Event), events ...interface{}) *subscription
	Name(f func(e Event), aggregate string, events ...string) *subscription
}

var (
	// ErrAggregateNotFound returns if events not found for aggregate or aggregate was not based on snapshot from the outside
	ErrAggregateNotFound = errors.New("aggregate not found")

	// ErrAggregateNotRegistered when saving aggregate when it's not registered in the repository
	ErrAggregateNotRegistered = errors.New("aggregate not registered")

	// ErrEventNotRegistered when saving aggregate and one event is not registered in the repository
	ErrEventNotRegistered = errors.New("event not registered")

	// ErrConcurrency when the currently saved version of the aggregate differs from the new events
	ErrConcurrency = errors.New("concurrency error")
)

type MarshalFunc func(v interface{}) ([]byte, error)
type UnmarshalFunc func(data []byte, v interface{}) error

// Repository is the returned instance from the factory function
type Repository struct {
	eventStream *EventStream
	eventStore  core.EventStore
	// register that convert the Data []byte to correct type
	register *register
	// serializer / deserializer
	Serializer   MarshalFunc
	Deserializer UnmarshalFunc
}

// NewRepository factory function
func NewRepository(eventStore core.EventStore) *Repository {
	return &Repository{
		eventStore:   eventStore,
		eventStream:  NewEventStream(),
		Serializer:   json.Marshal,
		Deserializer: json.Unmarshal,
		register:     newRegister(),
	}
}

func (r *Repository) Register(a aggregate) {
	r.register.Register(a)
}

// Subscribers returns an interface with all event subscribers
func (r *Repository) Subscribers() EventSubscribers {
	return r.eventStream
}

// Save an aggregates events
func (r *Repository) Save(a aggregate) error {
	// TODO: check that the aggregate is registered before saving it?
	if !r.register.AggregateRegistered(a) {
		return ErrAggregateNotRegistered
	}

	root := a.Root()
	// use under laying event slice to set GlobalVersion

	var esEvents = make([]core.Event, 0)

	// serialize the data and meta data into []byte
	for _, event := range root.aggregateEvents {
		data, err := r.Serializer(event.Data())
		if err != nil {
			return err
		}
		metadata, err := r.Serializer(event.Metadata())
		if err != nil {
			return err
		}

		esEvent := core.Event{
			AggregateID:   event.AggregateID(),
			Version:       core.Version(event.Version()),
			AggregateType: event.AggregateType(),
			Timestamp:     event.Timestamp(),
			Data:          data,
			Metadata:      metadata,
			Reason:        event.Reason(),
		}
		_, ok := r.register.EventRegistered(esEvent)
		if !ok {
			return ErrEventNotRegistered
		}
		esEvents = append(esEvents, esEvent)
	}

	err := r.eventStore.Save(esEvents)
	if err != nil {
		if errors.Is(err, core.ErrConcurrency) {
			return ErrConcurrency
		}
		return err
	}

	// update the global version on event bound to the aggregate
	for i, event := range esEvents {
		root.aggregateEvents[i].event.GlobalVersion = event.GlobalVersion
	}

	// publish the saved events to subscribers
	r.eventStream.Publish(*root, root.Events())

	// update the internal aggregate state
	root.update()
	return nil
}

// GetWithContext fetches the aggregates event and build up the aggregate based on it's current version.
// The event fetching can be canceled from the outside.
func (r *Repository) GetWithContext(ctx context.Context, id string, a aggregate) error {
	if reflect.ValueOf(a).Kind() != reflect.Ptr {
		return errors.New("aggregate needs to be a pointer")
	}

	root := a.Root()
	aggregateType := aggregateType(a)
	// fetch events after the current version of the aggregate that could be fetched from the snapshot store
	eventIterator, err := r.eventStore.Get(ctx, id, aggregateType, core.Version(root.aggregateVersion))
	if err != nil && !errors.Is(err, core.ErrNoEvents) {
		return err
	} else if errors.Is(err, core.ErrNoEvents) && root.Version() == 0 {
		// no events and not based on a snapshot
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
			if err != nil && !errors.Is(err, core.ErrNoMoreEvents) {
				return err
			} else if errors.Is(err, core.ErrNoMoreEvents) && root.Version() == 0 {
				// no events and no snapshot (some eventstore will not return the error ErrNoEvent on Get())
				return ErrAggregateNotFound
			} else if errors.Is(err, core.ErrNoMoreEvents) {
				return nil
			}
			// apply the event to the aggregate
			f, found := r.register.EventRegistered(event)
			if !found {
				continue
			}
			data := f()
			err = r.Deserializer(event.Data, &data)
			if err != nil {
				return err
			}
			metadata := make(map[string]interface{})
			err = r.Deserializer(event.Metadata, &metadata)
			if err != nil {
				return err
			}

			e := NewEvent(event, data, metadata)
			root.BuildFromHistory(a, []Event{e})
		}
	}
}

// Get fetches the aggregates event and build up the aggregate.
// If the aggregate is based on a snapshot it fetches event after the
// version of the aggregate.
func (r *Repository) Get(id string, a aggregate) error {
	return r.GetWithContext(context.Background(), id, a)
}

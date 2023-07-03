package base

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"time"
)

// Version is the event version used in event.Version, event.GlobalVersion and aggregateRoot
type Version uint64

// ErrNoEvents when there is no events to get
var ErrNoEvents = errors.New("no events")

// ErrNoMoreEvents when iterator has no more events to deliver
var ErrNoMoreEvents = errors.New("no more events")

// EventIterator is the interface an event store Get needs to return
type EventIterator interface {
	Next() (Event, error)
	Close()
}

// EventStore interface expose the methods an event store must uphold
type EventStore interface {
	Save(events []Event) error
	Get(ctx context.Context, id string, aggregateType string, afterVersion Version) (EventIterator, error)
}

// Event holding meta data and the application specific event in the Data property
type Event struct {
	AggregateID   string
	Version       Version
	GlobalVersion Version
	AggregateType string
	Timestamp     time.Time
	Data          interface{}
	Metadata      map[string]interface{}
}

// Reason returns the name of the data struct
func (e Event) Reason() string {
	if e.Data == nil {
		return ""
	}
	return reflect.TypeOf(e.Data).Elem().Name()
}

// DataAs convert the event.Data to the supplied type.
func (e Event) DataAs(i interface{}) error {
	b, err := json.Marshal(e.Data)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, i)
}

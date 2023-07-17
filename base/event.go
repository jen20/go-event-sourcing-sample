package base

import (
	"encoding/json"
	"errors"
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

// Event holding meta data and the application specific event in the Data property
type Event struct {
	AggregateID   string
	Version       Version
	GlobalVersion Version
	AggregateType string
	Timestamp     time.Time
	Reason        string // based on the Data type
	Data          []byte // interface{}
	Metadata      []byte // map[string]interface{}
}

// Reason returns the name of the data struct
// TODO: Not sure this should be in this struct.
/*
func (e Event) Reason() string {
	if e.Data == nil {
		return ""
	}
	return reflect.TypeOf(e.Data).Elem().Name()
}
*/

// DataAs convert the event.Data to the supplied type.
// TODO: Not sure this should be in this struct.
func (e Event) DataAs(i interface{}) error {
	b, err := json.Marshal(e.Data)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, i)
}

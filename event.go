package eventsourcing

import (
	"errors"
	"reflect"
	"time"
)

// ErrNoEvents when there is no events to get
var ErrNoEvents = errors.New("no events")

// ErrNoMoreEvents when iterator has no more events to deliver
var ErrNoMoreEvents = errors.New("no more events")

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

// EventUntyped holding meta data and the application specific event in the Data property
type EventUntyped struct {
	AggregateID   string
	Version       Version
	GlobalVersion Version
	AggregateType string
	Timestamp     time.Time
	Data          map[string]interface{}
	Metadata      map[string]interface{}
}

// Reason returns the name of the data struct
func (e Event) Reason() string {
	if e.Data == nil {
		return ""
	}
	return reflect.TypeOf(e.Data).Elem().Name()
}

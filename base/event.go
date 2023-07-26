package base

import (
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
	Data          []byte // interface{} on the external Event type
	Metadata      []byte // map[string]interface{} on the external Event type
}

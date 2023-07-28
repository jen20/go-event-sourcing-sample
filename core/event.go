package core

import (
	"time"
)

// Version is the event version used in event.Version and event.GlobalVersio
type Version uint64

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

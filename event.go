package eventsourcing

import (
	"reflect"
	"time"

	"github.com/hallgren/eventsourcing/core"
)

// Version is the event version used in event.Version and event.GlobalVersion
type Version core.Version

type Event struct {
	event    core.Event // internal event
	data     interface{}
	metadata map[string]interface{}
}

func NewEvent(e core.Event, data interface{}, metadata map[string]interface{}) Event {
	return Event{event: e, data: data, metadata: metadata}
}

func (e Event) Data() interface{} {
	return e.data
}

func (e Event) Metadata() map[string]interface{} {
	return e.metadata
}

func (e Event) AggregateType() string {
	return e.event.AggregateType
}

func (e Event) AggregateID() string {
	return e.event.AggregateID
}

func (e Event) Reason() string {
	if e.data == nil {
		return ""
	}
	return reflect.TypeOf(e.data).Elem().Name()
}

func (e Event) Version() Version {
	return Version(e.event.Version)
}

func (e Event) Timestamp() time.Time {
	return e.event.Timestamp
}

func (e Event) GlobalVersion() Version {
	return Version(e.event.GlobalVersion)
}

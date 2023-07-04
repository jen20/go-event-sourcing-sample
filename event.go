package eventsourcing

import (
	"time"

	"github.com/hallgren/eventsourcing/base"
)

type Event struct {
	event base.Event // internal event
}

func EventConvert(e base.Event) Event {
	return Event{e}
}

func (e Event) Data() interface{} {
	return e.event.Data
}

func (e Event) Metadata() map[string]interface{} {
	return e.event.Metadata
}

func (e Event) AggregateType() string {
	return e.event.AggregateType
}

func (e Event) AggregateID() string {
	return e.event.AggregateID
}

func (e Event) Reason() string {
	return e.event.Reason()
}

func (e Event) Version() uint64 {
	return uint64(e.event.Version)
}

func (e Event) Timestamp() time.Time {
	return e.event.Timestamp
}

func (e Event) GlobalVersion() uint64 {
	return uint64(e.event.GlobalVersion)
}

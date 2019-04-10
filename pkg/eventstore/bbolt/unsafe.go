// +build bbolt_unsafe

package bbolt

import (
	"go-event-sourcing-sample/pkg/eventsourcing"
	"unsafe"
)

func Serialize(event eventsourcing.Event) []byte {
	value := make([]byte, unsafe.Sizeof(eventsourcing.Event{}))
	t := (*eventsourcing.Event)(unsafe.Pointer(&value[0]))

	// Sets the properties on the event
	t.AggregateRootID = event.AggregateRootID
	t.AggregateType = event.AggregateType
	t.Data = event.Data
	t.MetaData = event.MetaData
	t.Reason = event.Reason
	t.Version = event.Version

	return value
}

func Deserialize(obj []byte) eventsourcing.Event {
	var event = &eventsourcing.Event{}
	event = (*eventsourcing.Event)(unsafe.Pointer(&obj[0]))
	return event
}

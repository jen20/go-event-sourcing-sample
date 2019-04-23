// +build bbolt_unsafe

package bbolt

import (
	"go-event-sourcing-sample"
	"unsafe"
)

func Serialize(event main.Event) []byte {
	value := make([]byte, unsafe.Sizeof(main.Event{}))
	t := (*main.Event)(unsafe.Pointer(&value[0]))

	// Sets the properties on the event
	t.AggregateRootID = event.AggregateRootID
	t.AggregateType = event.AggregateType
	t.Data = event.Data
	t.MetaData = event.MetaData
	t.Reason = event.Reason
	t.Version = event.Version

	return value
}

func Deserialize(obj []byte) main.Event {
	var event = &main.Event{}
	event = (*main.Event)(unsafe.Pointer(&obj[0]))
	return event
}

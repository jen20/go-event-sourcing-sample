package unsafe

import (
	"github.com/hallgren/eventsourcing"
	"unsafe"
)

type Handler struct{}

// New returns a json Handle
func New() *Handler {
	return &Handler{}
}

func (h *Handler) SerializeEvent(event eventsourcing.Event) ([]byte, error) {
	value := make([]byte, unsafe.Sizeof(eventsourcing.Event{}))
	t := (*eventsourcing.Event)(unsafe.Pointer(&value[0]))

	// Sets the properties on the event
	t.AggregateRootID = event.AggregateRootID
	t.AggregateType = event.AggregateType
	t.Data = event.Data
	t.MetaData = event.MetaData
	t.Reason = event.Reason
	t.Version = event.Version

	return value, nil
}

func (h *Handler) DeserializeEvent(obj []byte) (eventsourcing.Event, error) {
	var event = &eventsourcing.Event{}
	event = (*eventsourcing.Event)(unsafe.Pointer(&obj[0]))
	return *event, nil
}

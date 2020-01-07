package unsafe

import (
	"unsafe"

	"github.com/hallgren/eventsourcing"
)

// Handler of unsafe
type Handler struct{}

// New returns a json Handle
func New() *Handler {
	return &Handler{}
}

// SerializeEvent serializes an event to byte
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
	t.Timestamp = event.Timestamp

	return value, nil
}

// DeserializeEvent deserializes from byte to event
func (h *Handler) DeserializeEvent(obj []byte) (eventsourcing.Event, error) {
	event := (*eventsourcing.Event)(unsafe.Pointer(&obj[0]))
	return *event, nil
}

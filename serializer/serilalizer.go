package serializer

import (
	"errors"
	"github.com/hallgren/eventsourcing"
	"reflect"
)

// Handler for json serializes
type Handler struct {
	eventRegister map[string]interface{}
	serializer serializer
}

// New returns a json Handle
func New(serializer serializer) *Handler {
	return &Handler{
		eventRegister: make(map[string]interface{}),
		serializer: serializer,
	}
}

type aggregate interface {
	Transition(event eventsourcing.Event)
}

type serializer interface {
	Marshal(v interface{}) ([]byte, error)
	Unmarshal(data []byte, v interface{}) error
}

var (
	// ErrAggregateNameMissing return if aggregate name is missing
	ErrAggregateNameMissing = errors.New("missing aggregate name")

	// ErrNoEventsToRegister return if no events to register
	ErrNoEventsToRegister = errors.New("no events to register")

	// ErrEventNameMissing return if event name is missing
	ErrEventNameMissing = errors.New("missing event name")
)

// Register events aggregate
func (h *Handler) Register(aggregate aggregate, events ...interface{}) error {
	typ := reflect.TypeOf(aggregate).Elem().Name()
	if typ == "" {
		return ErrAggregateNameMissing
	}
	if len(events) == 0 {
		return ErrNoEventsToRegister
	}
	for _, event := range events {
		reason := reflect.TypeOf(event).Elem().Name()
		if reason == "" {
			return ErrEventNameMissing
		}
		h.eventRegister[typ+"_"+reason] = event
	}
	return nil
}

// EventStruct return a struct from the registry
func (h *Handler) EventStruct(typ, reason string) interface{} {
	return h.eventRegister[typ+"_"+reason]
}

// Marshal pass the request to the under laying Marshal method
func (h *Handler) Marshal(v interface{}) ([]byte, error) {
	return h.serializer.Marshal(v)
}

// Unmarshal pass the request to the under laying Unmarshal method
func (h *Handler) Unmarshal(data []byte, v interface{}) error {
	return h.serializer.Unmarshal(data, v)
}
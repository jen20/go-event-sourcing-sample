package serializer

import (
	"errors"
	"reflect"
	"sync"
)

type aggregate interface {}

type marshal func (v interface{}) ([]byte, error)
type unmarshal func(data []byte, v interface{}) error

// Handler for json serializes
type Handler struct {
	eventRegister map[string]interface{}
	marshal marshal
	unmarshal unmarshal
	lock sync.Mutex
}

// New returns a json Handle
func New(marshalF marshal, unmarshalF unmarshal) *Handler {
	return &Handler{
		eventRegister: make(map[string]interface{}),
		marshal: marshalF,
		unmarshal: unmarshalF,
	}
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
	h.lock.Lock()
	defer h.lock.Unlock()
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
	h.lock.Lock()
	defer h.lock.Unlock()
	return h.eventRegister[typ+"_"+reason]
}

// Marshal pass the request to the under laying Marshal method
func (h *Handler) Marshal(v interface{}) ([]byte, error) {
	h.lock.Lock()
	defer h.lock.Unlock()
	return h.marshal(v)
}

// Unmarshal pass the request to the under laying Unmarshal method
func (h *Handler) Unmarshal(data []byte, v interface{}) error {
	h.lock.Lock()
	defer h.lock.Unlock()
	return h.unmarshal(data, v)
}
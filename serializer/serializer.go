package serializer

import (
	"errors"
	"reflect"
)

type aggregate interface {}

type marshal func (v interface{}) ([]byte, error)
type unmarshal func(data []byte, v interface{}) error

// Handler for json serializes
type Handler struct {
	eventRegister map[string]func() interface{}
	f func(event interface{})
	marshal       marshal
	unmarshal     unmarshal
}

// EventFunc holds the event and a func how to construct a new instance of the event.
type EventFunc struct {
	Event interface{}
	F     EventF
}

type EventF = func() interface{}

// New returns a json Handle
func New(marshalF marshal, unmarshalF unmarshal) *Handler {
	return &Handler{
		eventRegister: make(map[string]EventF),
		marshal: marshalF,
		unmarshal: unmarshalF,
	}
}

var (
	// ErrAggregateNameMissing return if aggregate name is missing
	ErrAggregateNameMissing = errors.New("missing aggregate name")

	// ErrNoEventsToRegister return if no events to register
	ErrNoEventsToRegister = errors.New("no events to register")

	// ErrEventNameMissing return if Event name is missing
	ErrEventNameMissing = errors.New("missing Event name")
)


// RegisterTypes events aggregate
func (h *Handler) RegisterTypes(aggregate aggregate, events ...EventFunc) error {
	typ := reflect.TypeOf(aggregate).Elem().Name()
	if typ == "" {
		return ErrAggregateNameMissing
	}
	if len(events) == 0 {
		return ErrNoEventsToRegister
	}

	for _, event := range events {
		reason := reflect.TypeOf(event.Event).Elem().Name()
		if reason == "" {
			return ErrEventNameMissing
		}
		h.eventRegister[typ+"_"+reason] = event.F
	}
	return nil
}

// Type return a struct from the registry
func (h *Handler) Type(typ, reason string) EventF {
	return h.eventRegister[typ+"_"+reason]
}

// Marshal pass the request to the under laying Marshal method
func (h *Handler) Marshal(v interface{}) ([]byte, error) {
	return h.marshal(v)
}

// Unmarshal pass the request to the under laying Unmarshal method
func (h *Handler) Unmarshal(data []byte, v interface{}) error {
	return h.unmarshal(data, v)
}
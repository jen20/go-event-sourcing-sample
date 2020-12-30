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

type EventFunc = func() interface{}

// New returns a json Handle
func New(marshalF marshal, unmarshalF unmarshal) *Handler {
	return &Handler{
		eventRegister: make(map[string]EventFunc),
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

	for _, f := range events {
		event := f()
		reason := reflect.TypeOf(event).Elem().Name()
		if reason == "" {
			return ErrEventNameMissing
		}
		h.eventRegister[typ+"_"+reason] = f
	}
	return nil
}

// Type return a struct from the registry
func (h *Handler) Type(typ, reason string) EventFunc {
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
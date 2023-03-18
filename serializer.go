package eventsourcing

import (
	"errors"
	"reflect"
)

type eventFunc = func() interface{}
type EventsFunc = func(events ...interface{}) error
type MarshalSnapshotFunc func(v interface{}) ([]byte, error)
type UnmarshalSnapshotFunc func(data []byte, v interface{}) error

type aggregate interface {
	RegisterEvents(EventsFunc) error
}

// Serializer for json serializes
type Serializer struct {
	eventRegister map[string]eventFunc
	marshal       MarshalSnapshotFunc
	unmarshal     UnmarshalSnapshotFunc
}

// NewSerializer returns a json Handle
func NewSerializer(marshalF MarshalSnapshotFunc, unmarshalF UnmarshalSnapshotFunc) *Serializer {
	return &Serializer{
		eventRegister: make(map[string]eventFunc),
		marshal:       marshalF,
		unmarshal:     unmarshalF,
	}
}

var (
	// ErrAggregateNameMissing return if aggregate name is missing
	ErrAggregateNameMissing = errors.New("missing aggregate name")

	// ErrNoEventsToRegister return if no events to register
	ErrNoEventsToRegister = errors.New("no events to register")

	// ErrEventNameMissing return if Event name is missing
	ErrEventNameMissing = errors.New("missing event name")
)

func eventToFunc(event interface{}) eventFunc {
	return func() interface{} { return event }
}

// Events is a helper function to make the event type registration simpler
func (h *Serializer) Events(events ...interface{}) []eventFunc {
	res := []eventFunc{}
	for _, e := range events {
		res = append(res, eventToFunc(e))
	}
	return res
}

func (h *Serializer) RegisterAggregate(a aggregate) error {
	typ := reflect.TypeOf(a).Elem().Name()
	if typ == "" {
		return ErrAggregateNameMissing
	}

	fu := func(events ...interface{}) error {
		eventsF := h.Events(events...)
		for _, f := range eventsF {
			event := f()
			reason := reflect.TypeOf(event).Elem().Name()
			if reason == "" {
				return ErrEventNameMissing
			}
			h.eventRegister[typ+"_"+reason] = f
		}
		return nil
	}

	return a.RegisterEvents(fu)
}

// Register will hold a map of aggregate_event to be able to set the currect type when
// the data is unmarhaled.
func (h *Serializer) Register(aggregate Aggregate, events []eventFunc) error {
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

// RegisterTypes events aggregate
func (h *Serializer) RegisterTypes(aggregate Aggregate, events ...eventFunc) error {
	return h.Register(aggregate, events)
}

// Type return a struct from the registry
func (h *Serializer) Type(typ, reason string) (eventFunc, bool) {
	d, ok := h.eventRegister[typ+"_"+reason]
	return d, ok
}

// Marshal pass the request to the under laying Marshal method
func (h *Serializer) Marshal(v interface{}) ([]byte, error) {
	return h.marshal(v)
}

// Unmarshal pass the request to the under laying Unmarshal method
func (h *Serializer) Unmarshal(data []byte, v interface{}) error {
	return h.unmarshal(data, v)
}

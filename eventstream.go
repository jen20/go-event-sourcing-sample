package eventsourcing

import (
	"github.com/imkira/go-observer"
	"reflect"
)

type EventStream struct {
	eventRegister map[reflect.Type][]observer.Property
	global        observer.Property
}

// NewEventStream factory function
func NewEventStream() *EventStream {
	return &EventStream{
		eventRegister: make(map[reflect.Type][]observer.Property),
		global:        observer.NewProperty(nil),
	}
}

func (e *EventStream) Update(event Event) {
	// update streams that subscribe to the event
	t := reflect.TypeOf(event.Data)
	if props, ok := e.eventRegister[t]; ok {
		for _, prop := range props {
			prop.Update(event)
		}
	}
	// update the global stream
	e.global.Update(event)
}

func (e *EventStream) Subscribe(events ...interface{}) observer.Stream {
	// subscribe to global event changes
	if events == nil {
		return e.global.Observe()
	}

	// new property
	prop := observer.NewProperty(nil)

	// add prop to global events
	for _, event := range events {
		t := reflect.TypeOf(event)
		if e.eventRegister[t] == nil {
			// add the event type and prop to the empty register key
			e.eventRegister[t] = []observer.Property{prop}
		} else {
			// adds one more subscriber to the event type
			e.eventRegister[t] = append(e.eventRegister[t], prop)
		}
	}
	return prop.Observe()
}

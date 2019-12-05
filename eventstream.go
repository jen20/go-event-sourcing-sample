package eventsourcing

import (
	"reflect"
	"sync"
)

type EventStream struct {
	specificEvents map[reflect.Type][]func(e Event)
	allEvents      []func(e Event)
	publishLock    sync.Mutex
}

// NewEventStream factory function
func NewEventStream() *EventStream {
	return &EventStream{
		specificEvents: make(map[reflect.Type][]func(e Event)),
		allEvents:      []func(e Event){},
	}
}

// Update calls the functions that are subscribing to event
func (e *EventStream) Update(events []Event) {
	// the lock prevent other event updates get mixed with this update
	e.publishLock.Lock()
	for _, event := range events {

		// call all functions that has registered for the specific event
		t := reflect.TypeOf(event.Data)
		if functions, ok := e.specificEvents[t]; ok {
			for _, f := range functions {
				f(event)
			}
		}

		// call all functions that has registered for all events
		for _, f := range e.allEvents {
			f(event)
		}
	}
	e.publishLock.Unlock()
}

// Subscribe bind the f function to be called either when all or specific events are created in the system
func (e *EventStream) Subscribe(f func(e Event), events ...interface{}) {
	// subscribe to all event changes
	if events == nil {
		e.allEvents = append(e.allEvents, f)
		return
	}
	// subscribe to specified events
	for _, event := range events {
		t := reflect.TypeOf(event)
		if e.specificEvents[t] == nil {
			// add the event type and prop to the empty register key
			e.specificEvents[t] = []func(e Event){f}
		} else {
			// adds one more property to the event type
			e.specificEvents[t] = append(e.specificEvents[t], f)
		}
	}
}

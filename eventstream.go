package eventsourcing

import (
	"reflect"
	"sync"
)

// EventStream struct what handles event subscription
type EventStream struct {
	aggregateEvents map[string][]func(e Event)
	specificEvents  map[reflect.Type][]func(e Event)
	allEvents       []func(e Event)
	publishLock     sync.Mutex
}

// NewEventStream factory function
func NewEventStream() *EventStream {
	return &EventStream{
		aggregateEvents: make(map[string][]func(e Event)),
		specificEvents:  make(map[reflect.Type][]func(e Event)),
		allEvents:       []func(e Event){},
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

		// call all functions that has registered for the aggregate events
		if functions, ok := e.aggregateEvents[event.AggregateType]; ok {
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

// SubscribeAll bind the f function to be called on all events independent on aggregate or event type
func (e *EventStream) SubscribeAll(f func(e Event)) {
	e.allEvents = append(e.allEvents, f)
}

// SubscribeAggregate bind the f function to be called on events on the aggregate type
func (e *EventStream) SubscribeAggregate(f func(e Event), a aggregate) {
	aggregateType := reflect.TypeOf(a).Elem().Name()
	if e.aggregateEvents[aggregateType] == nil {
		// add the event type and prop to the empty register key
		e.aggregateEvents[aggregateType] = []func(e Event){f}
	} else {
		// adds one more property to the event type
		e.aggregateEvents[aggregateType] = append(e.aggregateEvents[aggregateType], f)
	}
}

// SubscribeSpecific bind the f function to be called on specific events
func (e *EventStream) SubscribeSpecific(f func(e Event), events ...interface{}) {
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

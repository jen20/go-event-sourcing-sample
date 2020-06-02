package eventsourcing

import (
	"fmt"
	"reflect"
	"sync"
)

// EventStream struct what handles event subscription
type EventStream struct {
	// holds subscribers of aggregate types events
	aggregateTypes map[string][]func(e Event)
	// holds subscribers of specific aggregates (type and identifier)
	specificAggregates map[string][]func(e Event)
	// holds subscribers of specific events
	specificEvents map[reflect.Type][]func(e Event)
	// holds subscribers of all events
	allEvents []func(e Event)
	// makes sure events are delivered in order
	publishLock sync.Mutex
}

// NewEventStream factory function
func NewEventStream() *EventStream {
	return &EventStream{
		aggregateTypes:     make(map[string][]func(e Event)),
		specificAggregates: make(map[string][]func(e Event)),
		specificEvents:     make(map[reflect.Type][]func(e Event)),
		allEvents:          []func(e Event){},
	}
}

// Update calls the functions that are subscribing to event
func (e *EventStream) Update(agg aggregate, events []Event) {
	// the lock prevent other event updates get mixed with this update
	e.publishLock.Lock()
	for _, event := range events {
		// call all functions that has registered for all events
		for _, f := range e.allEvents {
			f(event)
		}

		// call all functions that has registered for the specific event
		t := reflect.TypeOf(event.Data)

		if functions, ok := e.specificEvents[t]; ok {
			for _, f := range functions {
				f(event)
			}
		}

		var pkgPath string
		if agg != nil {
			pkgPath = reflect.TypeOf(agg).Elem().PkgPath()
		}

		ref := fmt.Sprintf("%s_%s",pkgPath, event.AggregateType)
		// call all functions that has registered for the aggregate type events
		if functions, ok := e.aggregateTypes[ref]; ok {
			for _, f := range functions {
				f(event)
			}
		}

		// call all functions that has registered for the aggregate type and id events
		// ref also include the package name ensuring that Aggregate Types can have the same name.
		ref = fmt.Sprintf("%s_%s", ref, agg.id())
		if functions, ok := e.specificAggregates[ref]; ok {
			for _, f := range functions {
				f(event)
			}
		}


	}
	e.publishLock.Unlock()
}

// SubscribeAll bind the f function to be called on all events independent on aggregate or event type
func (e *EventStream) SubscribeAll(f func(e Event)) {
	e.allEvents = append(e.allEvents, f)
}

// SubscribeSpecificAggregate bind the f function to be called on events that belongs to aggregate based on type and id
func (e *EventStream) SubscribeSpecificAggregate(f func(e Event), aggregates ...aggregate) {
	for _, a := range aggregates {
		pkgPath := reflect.TypeOf(a).Elem().PkgPath()
		aggregateType := reflect.TypeOf(a).Elem().Name()
		ref := fmt.Sprintf("%s_%s_%s", pkgPath, aggregateType, a.id())
		if e.specificAggregates[ref] == nil {
			// add the name and id of the aggregate and function to call to the empty register key
			e.specificAggregates[ref] = []func(e Event){f}
		} else {
			// adds one more function to the aggregate
			e.specificAggregates[ref] = append(e.specificAggregates[ref], f)
		}
	}
}

// SubscribeAggregateTypes bind the f function to be called on events on the aggregate type
func (e *EventStream) SubscribeAggregateTypes(f func(e Event), aggregates ...aggregate) {
	for _, a := range aggregates {
		pkgPath := reflect.TypeOf(a).Elem().PkgPath()
		name := reflect.TypeOf(a).Elem().Name()
		ref := fmt.Sprintf("%s_%s",pkgPath, name)

		if e.aggregateTypes[ref] == nil {
			// add the name of the aggregate and function to call to the empty register key
			e.aggregateTypes[ref] = []func(e Event){f}
		} else {
			// adds one more function to the aggregate
			e.aggregateTypes[ref] = append(e.aggregateTypes[ref], f)
		}
	}
}

// SubscribeSpecificEvent bind the f function to be called on specific events
func (e *EventStream) SubscribeSpecificEvent(f func(e Event), events ...interface{}) {
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

package eventsourcing

import (
	"fmt"
	"reflect"
	"sync"
)

// EventStream struct what handles event subscription
type EventStream struct {
	// holds subscribers of aggregate types events
	aggregateTypes map[string][]*subscription
	// holds subscribers of specific aggregates (type and identifier)
	specificAggregates map[string][]*subscription
	// holds subscribers of specific events
	specificEvents map[reflect.Type][]*subscription
	// holds subscribers of all events
	allEvents []*subscription
	// makes sure events are delivered in order and subscriptions are persistent
	lock sync.Mutex
}

// subscription holding the subscribe / unsubscribe / and func to be called when
// event matches the subscription
type subscription struct {
	funcEvent        func(e Event)
	funcMapInterface func(e map[string]interface{})
	unsubF           func()
	subF             func()
}

// Unsubscribe stops the subscription
func (s *subscription) Unsubscribe() {
	s.unsubF()
}

// Subscribe starts the subscription
func (s *subscription) Subscribe() {
	s.subF()
}

// NewEventStream factory function
func NewEventStream() *EventStream {
	return &EventStream{
		aggregateTypes:     make(map[string][]*subscription),
		specificAggregates: make(map[string][]*subscription),
		specificEvents:     make(map[reflect.Type][]*subscription),
		allEvents:          []*subscription{},
	}
}

// Publish calls the functions that are subscribing to the event stream
func (e *EventStream) Publish(agg AggregateRoot, events []Event) {
	// the lock prevent other event updates get mixed with this update
	e.lock.Lock()
	defer e.lock.Unlock()

	for _, event := range events {
		e.allEventPublisher(event)
		e.specificEventPublisher(event)
		e.aggregateTypePublisher(agg, event)
		e.specificAggregatesPublisher(agg, event)
	}
}

// call all functions that has registered for all events
func (e *EventStream) allEventPublisher(event Event) {
	for _, s := range e.allEvents {
		s.funcEvent(event)
	}
}

// call all functions that has registered for the specific event
func (e *EventStream) specificEventPublisher(event Event) {
	t := reflect.TypeOf(event.Data)
	if subs, ok := e.specificEvents[t]; ok {
		for _, s := range subs {
			s.funcEvent(event)
		}
	}
}

// call all functions that has registered for the aggregate type events
func (e *EventStream) aggregateTypePublisher(agg AggregateRoot, event Event) {
	ref := fmt.Sprintf("%s_%s", agg.path(), event.AggregateType)
	if subs, ok := e.aggregateTypes[ref]; ok {
		for _, s := range subs {
			s.funcEvent(event)
		}
	}
}

// call all functions that has registered for the aggregate type and ID events
func (e *EventStream) specificAggregatesPublisher(agg AggregateRoot, event Event) {
	// ref also include the package name ensuring that Aggregate Types can have the same name.
	ref := fmt.Sprintf("%s_%s", agg.path(), event.AggregateType)
	ref = fmt.Sprintf("%s_%s", ref, agg.ID())
	if subs, ok := e.specificAggregates[ref]; ok {
		for _, s := range subs {
			s.funcEvent(event)
		}
	}
}

// SubscriberAll bind the funcEvent function to be called on all events independent on aggregate or event type
func (e *EventStream) SubscriberAll(f func(e Event)) *subscription {
	s := subscription{
		funcEvent: f,
	}
	s.unsubF = func() {
		e.lock.Lock()
		defer e.lock.Unlock()

		for i, sub := range e.allEvents {
			if &s == sub {
				e.allEvents = append(e.allEvents[:i], e.allEvents[i+1:]...)
				break
			}
		}
	}
	s.subF = func() {
		e.lock.Lock()
		defer e.lock.Unlock()

		e.allEvents = append(e.allEvents, &s)
	}
	return &s
}

// SubscriberSpecificAggregate bind the funcEvent function to be called on events that belongs to aggregate based on type and ID
func (e *EventStream) SubscriberSpecificAggregate(f func(e Event), aggregates ...Aggregate) *subscription {
	s := subscription{
		funcEvent: f,
	}
	s.unsubF = func() {
		e.lock.Lock()
		defer e.lock.Unlock()

		for _, a := range aggregates {
			name := reflect.TypeOf(a).Elem().Name()
			root := a.Root()
			ref := fmt.Sprintf("%s_%s_%s", root.path(), name, root.ID())
			for i, sub := range e.specificAggregates[ref] {
				if &s == sub {
					e.specificAggregates[ref] = append(e.specificAggregates[ref][:i], e.specificAggregates[ref][i+1:]...)
					break
				}
			}
		}
	}
	s.subF = func() {
		e.lock.Lock()
		defer e.lock.Unlock()

		for _, a := range aggregates {
			name := reflect.TypeOf(a).Elem().Name()
			root := a.Root()
			ref := fmt.Sprintf("%s_%s_%s", root.path(), name, root.ID())

			// adds one more function to the aggregate
			e.specificAggregates[ref] = append(e.specificAggregates[ref], &s)
		}
	}
	return &s
}

// SubscriberAggregateType bind the funcEvent function to be called on events on the aggregate type
func (e *EventStream) SubscriberAggregateType(f func(e Event), aggregates ...Aggregate) *subscription {
	s := subscription{
		funcEvent: f,
	}
	s.unsubF = func() {
		e.lock.Lock()
		defer e.lock.Unlock()

		for _, a := range aggregates {
			name := reflect.TypeOf(a).Elem().Name()
			root := a.Root()
			ref := fmt.Sprintf("%s_%s", root.path(), name)
			for i, sub := range e.aggregateTypes[ref] {
				if &s == sub {
					e.aggregateTypes[ref] = append(e.aggregateTypes[ref][:i], e.aggregateTypes[ref][i+1:]...)
					break
				}
			}
		}
	}
	s.subF = func() {
		e.lock.Lock()
		defer e.lock.Unlock()

		for _, a := range aggregates {
			name := reflect.TypeOf(a).Elem().Name()
			root := a.Root()
			ref := fmt.Sprintf("%s_%s", root.path(), name)

			// adds one more function to the aggregate
			e.aggregateTypes[ref] = append(e.aggregateTypes[ref], &s)
		}
	}
	return &s
}

// SubscriberSpecificEvent bind the funcEvent function to be called on specific events
func (e *EventStream) SubscriberSpecificEvent(f func(e Event), events ...interface{}) *subscription {
	s := subscription{
		funcEvent: f,
	}
	s.unsubF = func() {
		e.lock.Lock()
		defer e.lock.Unlock()

		for _, event := range events {
			t := reflect.TypeOf(event)
			for i, sub := range e.specificEvents[t] {
				if &s == sub {
					e.specificEvents[t] = append(e.specificEvents[t][:i], e.specificEvents[t][i+1:]...)
					break
				}
			}
		}
	}
	s.subF = func() {
		e.lock.Lock()
		defer e.lock.Unlock()

		// subscribe to specified events
		for _, event := range events {
			t := reflect.TypeOf(event)
			// adds one more property to the event type
			e.specificEvents[t] = append(e.specificEvents[t], &s)
		}
	}
	return &s
}

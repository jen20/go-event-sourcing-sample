package eventsourcing

import (
	"fmt"
	"reflect"
	"sync"
)

// EventStream struct that handles event subscription
type EventStream struct {
	// makes sure events are delivered in order and subscriptions are persistent
	lock sync.Mutex

	// subscribers that get types Events
	// holds subscribers of aggregate types events
	aggregateTypes map[string][]*subscription
	// holds subscribers of specific aggregates (type and identifier)
	specificAggregates map[string][]*subscription
	// holds subscribers of specific events
	specificEvents map[reflect.Type][]*subscription
	// holds subscribers of all events
	allEvents []*subscription
	// holds subscribers of aggregate types by name
	names map[string][]*subscription
}

// subscription holding the subscribe / unsubscribe / and func to be called when
// event matches the subscription
type subscription struct {
	eventF func(e Event)
	unsubF func()
	subF   func()
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
		allEvents:          make([]*subscription, 0),
		names:              make(map[string][]*subscription),
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
		e.aggregateTypePublisherUntyped(event)
	}
}

// call all functions that has registered for all events
func (e *EventStream) allEventPublisher(event Event) {
	for _, s := range e.allEvents {
		s.eventF(event)
	}
}

// call all functions that has registered for the specific event
func (e *EventStream) specificEventPublisher(event Event) {
	t := reflect.TypeOf(event.Data)
	if subs, ok := e.specificEvents[t]; ok {
		for _, s := range subs {
			s.eventF(event)
		}
	}
}

// call all functions that has registered for the aggregate type events
func (e *EventStream) aggregateTypePublisher(agg AggregateRoot, event Event) {
	ref := fmt.Sprintf("%s_%s", agg.path(), event.AggregateType)
	if subs, ok := e.aggregateTypes[ref]; ok {
		for _, s := range subs {
			s.eventF(event)
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
			s.eventF(event)
		}
	}
}

// call all functions that has registered for the aggregate type events
func (e *EventStream) aggregateTypePublisherUntyped(event Event) {
	ref := event.AggregateType+"_"+event.Reason()
	if subs, ok := e.names[ref]; ok {
		for _, s := range subs {
			s.eventF(event)
		}
	}
}

// All subscribe to all events that is stored in the repository
func (e *EventStream) All(f func(e Event)) *subscription {
	s := subscription{
		eventF: f,
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

// Aggregate subscribe to events that belongs to aggregates based on its type and ID
func (e *EventStream) Aggregate(f func(e Event), aggregates ...Aggregate) *subscription {
	s := subscription{
		eventF: f,
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

// AggregateType subscribe to events based on the aggregate type
func (e *EventStream) AggregateType(f func(e Event), aggregates ...Aggregate) *subscription {
	s := subscription{
		eventF: f,
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

// SpecificEvent subscribe on specific events where the interface is a pointer to the event struct
func (e *EventStream) SpecificEvent(f func(e Event), events ...interface{}) *subscription {
	s := subscription{
		eventF: f,
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

// Name subscribe to aggregate and events names
func (e *EventStream) Name(f func(e Event), aggregate string, events ...string) *subscription {
	s := subscription{
		eventF: f,
	}
	s.unsubF = func() {
		e.lock.Lock()
		defer e.lock.Unlock()

		for _, event := range events {
			ref := aggregate+"_"+event
			for i, sub := range e.names[ref] {
				if &s == sub {
					e.names[ref] = append(e.names[ref][:i], e.names[ref][i+1:]...)
					break
				}
			}
		}
	}
	s.subF = func() {
		e.lock.Lock()
		defer e.lock.Unlock()

		for _, event := range events {
			ref := aggregate+"_"+event
			e.names[ref] = append(e.names[ref], &s)
		}
	}
	return &s
}

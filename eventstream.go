package eventsourcing

import (
	"fmt"
	"reflect"
	"sync"
)

// EventStream struct what handles event subscription
type EventStream struct {
	// holds subscribers of aggregate types events
	aggregateTypes map[string][]*Subscription
	// holds subscribers of specific aggregates (type and identifier)
	specificAggregates map[string][]*Subscription
	// holds subscribers of specific events
	specificEvents map[reflect.Type][]*Subscription
	// holds subscribers of all events
	allEvents []*Subscription
	// makes sure events are delivered in order
	publishLock sync.Mutex
}

type Subscription struct {
	f func(e Event)
	closeF func()
}

func (s *Subscription) Close() {
	s.closeF()
}

// NewEventStream factory function
func NewEventStream() *EventStream {
	return &EventStream{
		aggregateTypes:     make(map[string][]*Subscription),
		specificAggregates: make(map[string][]*Subscription),
		specificEvents:     make(map[reflect.Type][]*Subscription),
		allEvents:          []*Subscription{},
	}
}

// Update calls the functions that are subscribing to event
func (e *EventStream) Update(agg aggregate, events []Event) {
	// the lock prevent other event updates get mixed with this update
	e.publishLock.Lock()
	defer e.publishLock.Unlock()

	for _, event := range events {
		// call all functions that has registered for all events
		for _, s := range e.allEvents {
			s.f(event)
		}

		// call all functions that has registered for the specific event
		t := reflect.TypeOf(event.Data)
		if subs, ok := e.specificEvents[t]; ok {
			for _, s := range subs {
				s.f(event)
			}
		}

		ref := fmt.Sprintf("%s_%s", agg.path(), event.AggregateType)
		// call all functions that has registered for the aggregate type events
		if subs, ok := e.aggregateTypes[ref]; ok {
			for _, s := range subs {
				s.f(event)
			}
		}

		// call all functions that has registered for the aggregate type and id events
		// ref also include the package name ensuring that Aggregate Types can have the same name.
		ref = fmt.Sprintf("%s_%s", ref, agg.id())
		if subs, ok := e.specificAggregates[ref]; ok {
			for _, s := range subs {
				s.f(event)
			}
		}
	}
}

// SubscribeAll bind the f function to be called on all events independent on aggregate or event type
func (e *EventStream) SubscribeAll(f func(e Event)) *Subscription {
	s := Subscription{
		f : f,
	}
	s.closeF = func() {
		for i, sub := range e.allEvents {
			if &s == sub {
				e.allEvents = append(e.allEvents[:i], e.allEvents[i+1:]...)
				break
			}
		}
	}
	e.allEvents = append(e.allEvents, &s)
	return &s
}

// SubscribeSpecificAggregate bind the f function to be called on events that belongs to aggregate based on type and id
func (e *EventStream) SubscribeSpecificAggregate(f func(e Event), aggregates ...aggregate) *Subscription {
	s := Subscription{
		f : f,
	}
	s.closeF = func() {
		for _, a := range aggregates {
			name := reflect.TypeOf(a).Elem().Name()
			ref := fmt.Sprintf("%s_%s_%s", a.path(), name, a.id())
			for i, sub := range e.specificAggregates[ref] {
				if &s == sub {
					e.specificAggregates[ref] = append(e.specificAggregates[ref][:i], e.specificAggregates[ref][i+1:]...)
					break
				}
			}
		}
	}
	for _, a := range aggregates {
		name := reflect.TypeOf(a).Elem().Name()
		ref := fmt.Sprintf("%s_%s_%s", a.path(), name, a.id())

		// adds one more function to the aggregate
		e.specificAggregates[ref] = append(e.specificAggregates[ref], &s)
	}
	return &s
}

// SubscribeAggregateType bind the f function to be called on events on the aggregate type
func (e *EventStream) SubscribeAggregateType(f func(e Event), aggregates ...aggregate) *Subscription {
	s := Subscription{
		f : f,
	}
	s.closeF = func() {
		for _, a := range aggregates {
			name := reflect.TypeOf(a).Elem().Name()
			ref := fmt.Sprintf("%s_%s", a.path(), name)
			for i, sub := range e.aggregateTypes[ref] {
				if &s == sub {
					e.aggregateTypes[ref] = append(e.aggregateTypes[ref][:i], e.aggregateTypes[ref][i+1:]...)
					break
				}
			}
		}
	}
	for _, a := range aggregates {
		name := reflect.TypeOf(a).Elem().Name()
		ref := fmt.Sprintf("%s_%s", a.path(), name)

		// adds one more function to the aggregate
		e.aggregateTypes[ref] = append(e.aggregateTypes[ref], &s)
	}
	return &s
}

// SubscribeSpecificEvent bind the f function to be called on specific events
func (e *EventStream) SubscribeSpecificEvent(f func(e Event), events ...interface{}) *Subscription {
	s := Subscription{
		f : f,
	}
	s.closeF = func() {
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

	// subscribe to specified events
	for _, event := range events {
		t := reflect.TypeOf(event)
		// adds one more property to the event type
		e.specificEvents[t] = append(e.specificEvents[t], &s)
	}
	return &s
}

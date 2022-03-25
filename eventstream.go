package eventsourcing

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
)

// EventStream struct what handles event subscriptionEvent
type EventStream struct {
	// makes sure events are delivered in order and subscriptions are persistent
	lock sync.Mutex

	// subscribers that get types Events
	// holds subscribers of aggregate types events
	aggregateTypes map[string][]*subscriptionEvent
	// holds subscribers of specific aggregates (type and identifier)
	specificAggregates map[string][]*subscriptionEvent
	// holds subscribers of specific events
	specificEvents map[reflect.Type][]*subscriptionEvent
	// holds subscribers of all events
	allEvents []*subscriptionEvent

	// subscribers that get map[string]interface{} events
	// holds subscribers of all events as maps
	allEventsAsMaps []*subscriptionEventAsMaps
}

// subscriptionEvent holding the subscribe / unsubscribe / and func to be called when
// event matches the subscriptionEvent
type subscriptionEvent struct {
	eventF func(e Event)
	unsubF func()
	subF   func()
}

// Unsubscribe stops the subscriptionEvent
func (s *subscriptionEvent) Unsubscribe() {
	s.unsubF()
}

// Subscribe starts the subscriptionEvent
func (s *subscriptionEvent) Subscribe() {
	s.subF()
}

// subscriptionEventAsMaps holding the subscribe / unsubscribe / and func to be called when
// event matches the subscriptionEvent
type subscriptionEventAsMaps struct {
	eventF func(e EventUntyped)
	unsubF func()
	subF   func()
}

// Unsubscribe stops the subscriptionEvent
func (s *subscriptionEventAsMaps) Unsubscribe() {
	s.unsubF()
}

// Subscribe starts the subscriptionEvent
func (s *subscriptionEventAsMaps) Subscribe() {
	s.subF()
}

// NewEventStream factory function
func NewEventStream() *EventStream {
	return &EventStream{
		aggregateTypes:     make(map[string][]*subscriptionEvent),
		specificAggregates: make(map[string][]*subscriptionEvent),
		specificEvents:     make(map[reflect.Type][]*subscriptionEvent),
		allEvents:          []*subscriptionEvent{},
	}
}

// Publish calls the functions that are subscribing to the event stream
func (e *EventStream) Publish(agg AggregateRoot, events []Event) {
	// the lock prevent other event updates get mixed with this update
	e.lock.Lock()
	defer e.lock.Unlock()

	for _, event := range events {
		// types events
		e.allEventPublisher(event)
		e.specificEventPublisher(event)
		e.aggregateTypePublisher(agg, event)
		e.specificAggregatesPublisher(agg, event)

		// map[string]interface{} events
		e.allEventPublisherAsMap(event)
	}
}

// call all functions that has registered for all events
func (e *EventStream) allEventPublisher(event Event) {
	for _, s := range e.allEvents {
		s.eventF(event)
	}
}

// call all functions that has registered for all events
func (e *EventStream) allEventPublisherAsMap(event Event) {
	var m EventUntyped

	// convert the typed event to map[string]interface{}
	b, err := json.Marshal(event)
	if err != nil {
		return
	}
	err = json.Unmarshal(b, &m)
	if err != nil {
		return
	}

	for _, s := range e.allEventsAsMaps {
		s.eventF(m)
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

// SubscriberAll bind the eventF function to be called on all events independent on aggregate or event type
func (e *EventStream) SubscriberAll(f func(e Event)) *subscriptionEvent {
	s := subscriptionEvent{
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

// SubscriberSpecificAggregate bind the eventF function to be called on events that belongs to aggregate based on type and ID
func (e *EventStream) SubscriberSpecificAggregate(f func(e Event), aggregates ...Aggregate) *subscriptionEvent {
	s := subscriptionEvent{
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

// SubscriberAggregateType bind the eventF function to be called on events on the aggregate type
func (e *EventStream) SubscriberAggregateType(f func(e Event), aggregates ...Aggregate) *subscriptionEvent {
	s := subscriptionEvent{
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

// SubscriberSpecificEvent bind the eventF function to be called on specific events
func (e *EventStream) SubscriberSpecificEvent(f func(e Event), events ...interface{}) *subscriptionEvent {
	s := subscriptionEvent{
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

// SubscriberAllAsMap bind the eventF function to be called on all events independent on aggregate or event type
func (e *EventStream) SubscriberAllAsMap(f func(e EventUntyped)) *subscriptionEventAsMaps {
	s := subscriptionEventAsMaps{
		eventF: f,
	}
	s.unsubF = func() {
		e.lock.Lock()
		defer e.lock.Unlock()

		for i, sub := range e.allEventsAsMaps {
			if &s == sub {
				e.allEventsAsMaps = append(e.allEventsAsMaps[:i], e.allEventsAsMaps[i+1:]...)
				break
			}
		}
	}
	s.subF = func() {
		e.lock.Lock()
		defer e.lock.Unlock()

		e.allEventsAsMaps = append(e.allEventsAsMaps, &s)
	}
	return &s
}

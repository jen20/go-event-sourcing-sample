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

	// subscribers that get untyped events -> events.Data as map[string]interface{}
	// holds subscribers of all untyped events
	allEventsUntyped []*subscriptionEventUntyped
	// holds subscribers of aggregate types
	aggregateTypesUntyped map[string][]*subscriptionEventUntyped
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

// subscriptionEventUntyped holding the subscribe / unsubscribe / and func to be called when
// event matches the subscriptionEvent
type subscriptionEventUntyped struct {
	eventF func(e EventUntyped)
	unsubF func()
	subF   func()
}

// Unsubscribe stops the subscriptionEvent
func (s *subscriptionEventUntyped) Unsubscribe() {
	s.unsubF()
}

// Subscribe starts the subscriptionEvent
func (s *subscriptionEventUntyped) Subscribe() {
	s.subF()
}

// NewEventStream factory function
func NewEventStream() *EventStream {
	return &EventStream{
		aggregateTypes:        make(map[string][]*subscriptionEvent),
		specificAggregates:    make(map[string][]*subscriptionEvent),
		specificEvents:        make(map[reflect.Type][]*subscriptionEvent),
		allEvents:             make([]*subscriptionEvent, 0),
		allEventsUntyped:      make([]*subscriptionEventUntyped, 0),
		aggregateTypesUntyped: make(map[string][]*subscriptionEventUntyped),
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

		// untyped events
		e.allEventPublisherUntyped(event)
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

// call all functions that has registered for all events
func (e *EventStream) allEventPublisherUntyped(event Event) {
	var m EventUntyped

	if len(e.allEventsUntyped) > 0 {
		// convert the typed event.Data to map[string]interface{}
		b, err := json.Marshal(event)
		if err != nil {
			return
		}
		err = json.Unmarshal(b, &m)
		if err != nil {
			return
		}

		for _, s := range e.allEventsUntyped {
			s.eventF(m)
		}
	}
}

// call all functions that has registered for the aggregate type events
func (e *EventStream) aggregateTypePublisherUntyped(event Event) {
	var m EventUntyped

	if subs, ok := e.aggregateTypesUntyped[event.AggregateType]; ok {
		b, err := json.Marshal(event)
		if err != nil {
			return
		}
		err = json.Unmarshal(b, &m)
		if err != nil {
			return
		}

		for _, s := range subs {
			s.eventF(m)
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

// SubscriberAllUntyped bind the eventF function to be called on all events independent on aggregate or event type
func (e *EventStream) SubscriberAllUntyped(f func(e EventUntyped)) *subscriptionEventUntyped {
	s := subscriptionEventUntyped{
		eventF: f,
	}
	s.unsubF = func() {
		e.lock.Lock()
		defer e.lock.Unlock()

		for i, sub := range e.allEventsUntyped {
			if &s == sub {
				e.allEventsUntyped = append(e.allEventsUntyped[:i], e.allEventsUntyped[i+1:]...)
				break
			}
		}
	}
	s.subF = func() {
		e.lock.Lock()
		defer e.lock.Unlock()

		e.allEventsUntyped = append(e.allEventsUntyped, &s)
	}
	return &s
}

// SubscriberAggregateTypeUntyped bind the eventF function to be called on events on the aggregate type
func (e *EventStream) SubscriberAggregateTypeUntyped(f func(e EventUntyped), aggregateTypes ...string) *subscriptionEventUntyped {
	s := subscriptionEventUntyped{
		eventF: f,
	}
	s.unsubF = func() {
		e.lock.Lock()
		defer e.lock.Unlock()

		for _, ref := range aggregateTypes {
			for i, sub := range e.aggregateTypesUntyped[ref] {
				if &s == sub {
					e.aggregateTypesUntyped[ref] = append(e.aggregateTypesUntyped[ref][:i], e.aggregateTypesUntyped[ref][i+1:]...)
					break
				}
			}
		}
	}
	s.subF = func() {
		e.lock.Lock()
		defer e.lock.Unlock()

		for _, ref := range aggregateTypes {
			// adds one more function to the aggregate
			e.aggregateTypesUntyped[ref] = append(e.aggregateTypesUntyped[ref], &s)
		}
	}
	return &s
}

package eventsourcing_test

import (
	"sync"
	"testing"

	"github.com/hallgren/eventsourcing"
)

type AnAggregate struct {
	eventsourcing.AggregateRoot
}

func (a *AnAggregate) Transition(e eventsourcing.Event) {}

type AnEvent struct {
	Name string
}

type AnotherAggregate struct {
	eventsourcing.AggregateRoot
}

func (a *AnotherAggregate) Transition(e eventsourcing.Event) {}

type AnotherEvent struct{}

var event = eventsourcing.Event{Version: 123, Data: &AnEvent{Name: "123"}, Reason: "AnEvent", AggregateType: "AnAggregate"}
var otherEvent = eventsourcing.Event{Version: 456, Data: &AnotherEvent{}, Reason: "AnotherEvent", AggregateType: "AnotherAggregate"}

func TestAll(t *testing.T) {
	var streamEvent *eventsourcing.Event
	e := eventsourcing.NewEventStream()
	f := func(e eventsourcing.Event) {
		streamEvent = &e
	}
	s := e.SubscribeAll(f)
	defer s.Close()
	e.Update(&AnAggregate{}, []eventsourcing.Event{event})

	if streamEvent == nil {
		t.Fatalf("should have received event")
	}
	if streamEvent.Version != event.Version {
		t.Fatalf("wrong info in event got %q expected %q", streamEvent.Version, event.Version)
	}
}

func TestSubscribeOneEvent(t *testing.T) {
	var streamEvent *eventsourcing.Event
	e := eventsourcing.NewEventStream()
	f := func(e eventsourcing.Event) {
		streamEvent = &e
	}
	s := e.SubscribeSpecificEvent(f, &AnEvent{})
	defer s.Close()
	e.Update(&AnAggregate{}, []eventsourcing.Event{event})

	if streamEvent == nil {
		t.Fatalf("should have received event")
	}

	if streamEvent.Version != event.Version {
		t.Fatalf("wrong info in event got %q expected %q", streamEvent.Version, event.Version)
	}
}

func TestSubscribeSpecificAggregate(t *testing.T) {
	// setup aggregates with identifiers
	anAggregate := AnAggregate{eventsourcing.AggregateRoot{AggregateID: "123"}}
	anOtherAggregate := AnotherAggregate{eventsourcing.AggregateRoot{AggregateID: "456"}}

	var streamEvent *eventsourcing.Event
	e := eventsourcing.NewEventStream()
	f := func(e eventsourcing.Event) {
		streamEvent = &e
	}
	s := e.SubscribeSpecificAggregate(f, &anAggregate, &anOtherAggregate)
	defer s.Close()
	// update with event from the AnAggregate aggregate
	e.Update(&anAggregate, []eventsourcing.Event{event})
	if streamEvent == nil {
		t.Fatalf("should have received event")
	}
	if streamEvent.Version != event.Version {
		t.Fatalf("wrong info in event got %q expected %q", streamEvent.Version, event.Version)
	}

	// update with event from the AnotherAggregate aggregate
	e.Update(&anOtherAggregate, []eventsourcing.Event{otherEvent})
	if streamEvent.Version != otherEvent.Version {
		t.Fatalf("wrong info in event got %q expected %q", streamEvent.Version, otherEvent.Version)
	}
}

func TestSubscribeAggregateType(t *testing.T) {
	var streamEvent *eventsourcing.Event
	e := eventsourcing.NewEventStream()
	f := func(e eventsourcing.Event) {
		streamEvent = &e
	}
	s := e.SubscribeAggregateType(f, &AnAggregate{}, &AnotherAggregate{})
	defer s.Close()

	// update with event from the AnAggregate aggregate
	e.Update(&AnAggregate{}, []eventsourcing.Event{event})
	if streamEvent == nil {
		t.Fatalf("should have received event")
	}
	if streamEvent.Version != event.Version {
		t.Fatalf("wrong info in event got %q expected %q", streamEvent.Version, event.Version)
	}

	// update with event from the AnotherAggregate aggregate
	e.Update(&AnotherAggregate{}, []eventsourcing.Event{otherEvent})
	if streamEvent.Version != otherEvent.Version {
		t.Fatalf("wrong info in event got %q expected %q", streamEvent.Version, otherEvent.Version)
	}
}

func TestSubscribeToManyEvents(t *testing.T) {
	var streamEvents []*eventsourcing.Event
	e := eventsourcing.NewEventStream()
	f := func(e eventsourcing.Event) {
		streamEvents = append(streamEvents, &e)
	}
	s := e.SubscribeSpecificEvent(f, &AnEvent{}, &AnotherEvent{})
	defer s.Close()
	e.Update(&AnAggregate{}, []eventsourcing.Event{event})
	e.Update(&AnotherAggregate{}, []eventsourcing.Event{otherEvent})

	if streamEvents == nil {
		t.Fatalf("should have received event")
	}

	if len(streamEvents) != 2 {
		t.Fatalf("should have received 2 events")
	}

	switch ev := streamEvents[0].Data.(type) {
	case *AnotherEvent:
		t.Fatalf("expecting AnEvent got %q", ev)
	}

	switch ev := streamEvents[1].Data.(type) {
	case *AnEvent:
		t.Fatalf("expecting OtherEvent got %q", ev)
	}

}

func TestUpdateNoneSubscribedEvent(t *testing.T) {
	var streamEvent *eventsourcing.Event
	e := eventsourcing.NewEventStream()
	f := func(e eventsourcing.Event) {
		streamEvent = &e
	}
	s := e.SubscribeSpecificEvent(f, &AnotherEvent{})
	defer s.Close()
	e.Update(&AnAggregate{}, []eventsourcing.Event{event})

	if streamEvent != nil {
		t.Fatalf("should not have received event %q", streamEvent)
	}
}

func TestManySubscribers(t *testing.T) {
	streamEvent1 := make([]eventsourcing.Event, 0)
	streamEvent2 := make([]eventsourcing.Event, 0)
	streamEvent3 := make([]eventsourcing.Event, 0)
	streamEvent4 := make([]eventsourcing.Event, 0)
	streamEvent5 := make([]eventsourcing.Event, 0)

	e := eventsourcing.NewEventStream()
	f1 := func(e eventsourcing.Event) {
		streamEvent1 = append(streamEvent1, e)
	}
	f2 := func(e eventsourcing.Event) {
		streamEvent2 = append(streamEvent2, e)
	}
	f3 := func(e eventsourcing.Event) {
		streamEvent3 = append(streamEvent3, e)
	}
	f4 := func(e eventsourcing.Event) {
		streamEvent4 = append(streamEvent4, e)
	}
	f5 := func(e eventsourcing.Event) {
		streamEvent5 = append(streamEvent5, e)
	}
	s := e.SubscribeSpecificEvent(f1, &AnotherEvent{})
	defer s.Close()
	s = e.SubscribeSpecificEvent(f2, &AnotherEvent{}, &AnEvent{})
	defer s.Close()
	s = e.SubscribeSpecificEvent(f3, &AnEvent{})
	defer s.Close()
	s = e.SubscribeAll(f4)
	defer s.Close()
	s = e.SubscribeAggregateType(f5, &AnAggregate{})
	defer s.Close()

	e.Update(&AnAggregate{}, []eventsourcing.Event{event})

	if len(streamEvent1) != 0 {
		t.Fatalf("stream1 should not have any events")
	}

	if len(streamEvent2) != 1 {
		t.Fatalf("stream2 should have one event")
	}

	if len(streamEvent3) != 1 {
		t.Fatalf("stream3 should have one event")
	}

	if len(streamEvent4) != 1 {
		t.Fatalf("stream4 should have one event")
	}

	if len(streamEvent5) != 1 {
		t.Fatalf("stream5 should have one event")
	}
}

func TestParallelUpdates(t *testing.T) {
	streamEvent := make([]eventsourcing.Event, 0)
	e := eventsourcing.NewEventStream()

	// functions to bind to event subscription
	f1 := func(e eventsourcing.Event) {
		streamEvent = append(streamEvent, e)
	}
	f2 := func(e eventsourcing.Event) {
		streamEvent = append(streamEvent, e)
	}
	f3 := func(e eventsourcing.Event) {
		streamEvent = append(streamEvent, e)
	}
	s := e.SubscribeSpecificEvent(f1, &AnEvent{})
	defer s.Close()
	s = e.SubscribeSpecificEvent(f2, &AnotherEvent{})
	defer s.Close()
	s = e.SubscribeAll(f3)
	defer s.Close()

	wg := sync.WaitGroup{}
	// concurrently update the event stream
	for i := 1; i < 1000; i++ {
		wg.Add(2)
		go func() {
			e.Update(&AnotherAggregate{}, []eventsourcing.Event{otherEvent, otherEvent})
			wg.Done()
		}()
		go func() {
			e.Update(&AnAggregate{}, []eventsourcing.Event{event, event})
			wg.Done()
		}()
	}
	wg.Wait()

	var lastEvent eventsourcing.Event
	// check that events comes coupled together in four due to the lock in the event stream that makes sure all registered
	// functions are called together and that is not mixed with other events
	for j, event := range streamEvent {
		if j%4 == 0 {
			lastEvent = event
		} else {
			if lastEvent.Reason != event.Reason {
				t.Fatal("same event should come in couple of four")
			}
		}
	}
}

func TestClose(t *testing.T) {
	count := 0
	e := eventsourcing.NewEventStream()
	f := func(e eventsourcing.Event) {
		count++
	}
	s1 := e.SubscribeAll(f)
	s2 := e.SubscribeSpecificEvent(f, &AnEvent{})
	s3 := e.SubscribeAggregateType(f, &AnAggregate{})
	s4 := e.SubscribeSpecificAggregate(f, &AnAggregate{})

	// trigger all 4 subscriptions
	e.Update(&AnAggregate{}, []eventsourcing.Event{event})
	if count != 4 {
		t.Fatalf("should have received four event")
	}
	// close all subscriptions
	s1.Close()
	s2.Close()
	s3.Close()
	s4.Close()

	// new event should not trigger closed subscriptions
	e.Update(&AnAggregate{}, []eventsourcing.Event{event})
	if count != 4 {
		t.Fatalf("should not have received event after subscriptions are closed")
	}
}

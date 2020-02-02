package eventsourcing_test

import (
	"sync"
	"testing"

	"github.com/hallgren/eventsourcing"
)

type AnAggregate struct {
	eventsourcing.AggregateRoot
}

func (a *AnAggregate) Transition(e eventsourcing.Event) {
}

type AnEvent struct {
	Name string
}

type AnotherEvent struct{}

var event = eventsourcing.Event{Version: 123, Data: &AnEvent{Name: "123"}, Reason: "AnEvent", AggregateType: "AnAggregate"}
var otherEvent = eventsourcing.Event{Version: 123, Data: &AnotherEvent{}, Reason: "AnotherEvent", AggregateType: "AnotherAggregate"}

func TestAll(t *testing.T) {
	var streamEvent *eventsourcing.Event
	e := eventsourcing.NewEventStream()
	f := func(e eventsourcing.Event) {
		streamEvent = &e
	}
	e.SubscribeAll(f)
	e.Update([]eventsourcing.Event{event})

	if streamEvent == nil {
		t.Fatalf("should have received event")
	}
	if streamEvent.Version != event.Version {
		t.Fatalf("wrong info in event got %q expected %q", streamEvent.Version, event.Version)
	}
}

func TestSpecific(t *testing.T) {
	var streamEvent *eventsourcing.Event
	e := eventsourcing.NewEventStream()
	f := func(e eventsourcing.Event) {
		streamEvent = &e
	}
	e.SubscribeSpecific(f, &AnEvent{})
	e.Update([]eventsourcing.Event{event})

	if streamEvent == nil {
		t.Fatalf("should have received event")
	}

	if streamEvent.Version != event.Version {
		t.Fatalf("wrong info in event got %q expected %q", streamEvent.Version, event.Version)
	}
}

func TestSubscribeAggregate(t *testing.T) {
	var streamEvent *eventsourcing.Event
	e := eventsourcing.NewEventStream()
	f := func(e eventsourcing.Event) {
		streamEvent = &e
	}
	e.SubscribeAggregate(f, &AnAggregate{})
	e.Update([]eventsourcing.Event{event})

	if streamEvent == nil {
		t.Fatalf("should have received event")
	}

	if streamEvent.Version != event.Version {
		t.Fatalf("wrong info in event got %q expected %q", streamEvent.Version, event.Version)
	}
}

func TestManySpecific(t *testing.T) {
	var streamEvents []*eventsourcing.Event
	e := eventsourcing.NewEventStream()
	f := func(e eventsourcing.Event) {
		streamEvents = append(streamEvents, &e)
	}
	e.SubscribeSpecific(f, &AnEvent{}, &AnotherEvent{})
	e.Update([]eventsourcing.Event{event})
	e.Update([]eventsourcing.Event{otherEvent})

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
	e.SubscribeSpecific(f, &AnotherEvent{})
	e.Update([]eventsourcing.Event{event})

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
	e.SubscribeSpecific(f1, &AnotherEvent{})
	e.SubscribeSpecific(f2, &AnotherEvent{}, &AnEvent{})
	e.SubscribeSpecific(f3, &AnEvent{})
	e.SubscribeAll(f4)
	e.SubscribeAggregate(f5, &AnAggregate{})


	e.Update([]eventsourcing.Event{event})

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
	e.SubscribeSpecific(f1, &AnEvent{})
	e.SubscribeSpecific(f2, &AnotherEvent{})
	e.SubscribeAll(f3)

	wg := sync.WaitGroup{}
	// concurrently update the event stream
	for i := 1; i < 1000; i++ {
		wg.Add(2)
		go func() {
			e.Update([]eventsourcing.Event{otherEvent, otherEvent})
			wg.Done()
		}()
		go func() {
			e.Update([]eventsourcing.Event{event, event})
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

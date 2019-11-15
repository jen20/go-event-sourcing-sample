package eventsourcing_test

import (
	"github.com/hallgren/eventsourcing"
	"testing"
)

type AnEvent struct {
	Name string
}

type AnotherEvent struct {}

var event = eventsourcing.Event{Version:123,Data:&AnEvent{Name:"123"}}

func TestAllEvents(t *testing.T) {
	e := eventsourcing.NewEventStream()
	stream := e.Subscribe()
	e.Update(event)

	if !stream.HasNext() {
		t.Fatalf("should have received event")
	}

	streamEvent := stream.WaitNext().(eventsourcing.Event)
	if streamEvent.Version != event.Version {
		t.Fatalf("wrong info in event got %q expected %q", streamEvent.Version, event.Version)
	}
}

func TestSpecificEvent(t *testing.T) {
	e := eventsourcing.NewEventStream()
	stream := e.Subscribe(&AnEvent{})
	e.Update(event)

	if !stream.HasNext() {
		t.Fatalf("should have received event")
	}

	streamEvent := stream.WaitNext().(eventsourcing.Event)
	if streamEvent.Version != event.Version {
		t.Fatalf("wrong info in event got %q expected %q", streamEvent.Version, event.Version)
	}
}

func TestUpdateNoneSubscribedEvent(t *testing.T) {
	e := eventsourcing.NewEventStream()
	stream := e.Subscribe(&AnotherEvent{})
	e.Update(event)

	if stream.HasNext() {
		t.Fatalf("should not have received event %q", stream.WaitNext())
	}
}

func TestManySubscribers(t *testing.T) {
	e := eventsourcing.NewEventStream()
	stream1 := e.Subscribe(&AnotherEvent{})
	stream2 := e.Subscribe(&AnotherEvent{}, &AnEvent{})
	stream3 := e.Subscribe(&AnEvent{})
	stream4 := e.Subscribe()

	e.Update(event)

	if stream1.HasNext() {
		t.Fatalf("stream1 should not have any events")
	}

	if !stream2.HasNext() {
		t.Fatalf("stream2 should have one event")
	} else {
		stream2.Next()
		if stream2.HasNext() {
			t.Fatalf("stream2 should only have one event")
		}
	}

	if !stream3.HasNext() {
		t.Fatalf("stream3 should have one event")
	} else {
		stream3.Next()
		if stream3.HasNext() {
			t.Fatalf("stream3 should only have one event")
		}
	}

	if !stream4.HasNext(){
		t.Fatalf("stream4 should have one event")
	} else {
		stream4.Next()
		if stream4.HasNext() {
			t.Fatalf("stream4 should only have one event")
		}
	}
}
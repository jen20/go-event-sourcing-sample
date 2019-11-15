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
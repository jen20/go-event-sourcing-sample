package eventsourcing_test

import (
	"github.com/hallgren/eventsourcing"
	"testing"
)

func TestAllEvents(t *testing.T) {
	e := eventsourcing.NewEventStream()
	stream := e.Subscribe(nil)

	event := eventsourcing.Event{Version:123}
	e.Update(event)

	if !stream.HasNext() {
		t.Fatalf("should have received event")
	}

	streamEvent := stream.WaitNext().(eventsourcing.Event)
	if streamEvent.Version != event.Version {
		t.Fatalf("wrong info in event got %q expected %q", streamEvent.Version, event.Version)
	}

}

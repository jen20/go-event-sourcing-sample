package eventsourcing_test

import (
	"testing"

	"github.com/hallgren/eventsourcing"
)

func TestEvent(t *testing.T) {
	e := eventsourcing.Event{
		Data: &Born{},
	}
	if e.Reason() != "Born" {
		t.Fatalf("expected Born got %s", e.Reason())
	}
}

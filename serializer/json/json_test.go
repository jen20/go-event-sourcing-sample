package json_test

import (
	"github.com/hallgren/eventsourcing"
	"github.com/hallgren/eventsourcing/serializer/json"
	"testing"
)

func TestSerializeDeserialize(t *testing.T) {
	h := json.New()
	v, err := h.Serialize(eventsourcing.Event{AggregateRootID: "123"})
	if err != nil {
		t.Fatalf("could not serialize event, %v", err)
	}
	event, err := h.Deserialize(v)
	if err != nil {
		t.Fatalf("Could not deserialize event, %v", err)
	}

	if event.AggregateRootID != "123" {
		t.Fatalf("wrong value in aggregateID expected: 123, actual: %v", event.AggregateRootID)
	}
}

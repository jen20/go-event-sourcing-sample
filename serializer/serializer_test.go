package serializer_test

import (
	"fmt"
	"github.com/hallgren/eventsourcing"
	"github.com/hallgren/eventsourcing/serializer/json"
	"github.com/hallgren/eventsourcing/serializer/unsafe"
	"reflect"
	"testing"
)

type serializer interface {
	Serialize(event eventsourcing.Event) ([]byte, error)
	Deserialize(v []byte) (event eventsourcing.Event, err error)
}

func initSerializers() []serializer {
	return []serializer{json.New(), unsafe.New()}
}

// FrequentFlierAccount represents the state of an instance of the frequent flier
// account aggregate. It tracks changes on itself in the form of domain events.
type testEvent struct {
	b int
}

type testAggregate struct {
	eventsourcing.AggregateRoot
	a int
}

// Create constructor for the Person
func Create() *testAggregate {
	aggregate := testAggregate{}
	aggregate.TrackChange(&aggregate, testEvent{b: 1})
	return &aggregate
}
// Transition the testAggregate state dependent on the events
func (t *testAggregate) Transition(event eventsourcing.Event) {
	switch e := ev.Data.(type) {

	case testEvent:
		t.a = e.b
	}
}

var ev = eventsourcing.Event{AggregateRootID: "123", Version: 7, Reason: "FlightTaken", AggregateType: "nana", Data: testEvent{b: 1}}

func TestSerializeDeserialize(t *testing.T) {
	serializers := initSerializers()

	for _, s := range serializers {
		t.Run(reflect.TypeOf(s).Elem().Name(), func(t *testing.T) {
			b, err := s.Serialize(ev)
			if err != nil {
				t.Fatalf("could not serialize ev, %v", err)
			}
			event2, err := s.Deserialize(b)
			if err != nil {
				t.Fatalf("could not deserialize to ev, %v", err)
			}
			aggregate := Create()
			testAggregate.BuildFromHistory(&aggregate, []eventsourcing.Event{ev})
			if ev.AggregateRootID != event2.AggregateRootID {
				t.Fatalf("ev id %q and event2 id %q are not the same", ev.AggregateRootID, event2.AggregateRootID)
			}
			if ev.Version != event2.Version {
				t.Fatalf("ev version %q and event2 version %q are not the same", ev.Version, event2.Version)
			}
			if ev.Data != event2.Data {
				fmt.Println(ev.Data)
				fmt.Println(event2.Data)
				t.Fatalf("ev data %q and event2 data %q are not the same", ev.Data, event2.Data)
			}
		})
	}
}
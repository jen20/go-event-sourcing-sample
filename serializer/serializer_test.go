package serializer_test

import (
	"github.com/hallgren/eventsourcing"
	"github.com/hallgren/eventsourcing/serializer/json"
	"github.com/hallgren/eventsourcing/serializer/unsafe"
	"reflect"
	"testing"
	"time"
)

type serializer interface {
	SerializeEvent(event eventsourcing.Event) ([]byte, error)
	DeserializeEvent(v []byte) (event eventsourcing.Event, err error)
}

func initSerializers(t *testing.T) []serializer {
	j := json.New()
	err := j.Register(&SomeAggregate{}, &SomeData{}, &SomeData2{})
	if err != nil {
		t.Fatalf("could not register aggregate events %v", err)
	}
	return []serializer{j, unsafe.New()}
}

type SomeAggregate struct{}

func (s *SomeAggregate) Transition(event eventsourcing.Event) {}

type SomeData struct {
	A int
	B string
}

type SomeData2 struct {
	A int
	B string
}

var data = SomeData{
	1,
	"b",
}

var metaData = make(map[string]interface{})

func TestSerializeDeserialize(t *testing.T) {
	serializers := initSerializers(t)
	metaData["foo"] = "bar"
	timestamp := time.Now().UTC()
	for _, s := range serializers {
		t.Run(reflect.TypeOf(s).Elem().Name(), func(t *testing.T) {
			v, err := s.SerializeEvent(eventsourcing.Event{
				AggregateRootID: "123",
				Version:         1,
				Data:            data,
				AggregateType:   "SomeAggregate",
				Reason:          "SomeData",
				MetaData:        metaData,
				Timestamp: 		 timestamp,
			})
			if err != nil {
				t.Fatalf("could not serialize event, %v", err)
			}
			event, err := s.DeserializeEvent(v)
			if err != nil {
				t.Fatalf("Could not deserialize event, %v", err)
			}

			if event.AggregateRootID != "123" {
				t.Fatalf("wrong value in aggregateID expected: 123, actual: %v", event.AggregateRootID)
			}

			if event.Reason != "SomeData" {
				t.Fatalf("wrong reason")
			}

			if event.MetaData["foo"] != "bar" {
				t.Fatalf("wrong value in meta data")
			}

			if event.Version != 1 {
				t.Fatalf("wrong value in AggregateVersion")
			}

			if event.Timestamp != timestamp {
				t.Fatalf("timestamp expected: %v got: %v", timestamp, event.Timestamp)
			}

			switch event.AggregateType {
			case "SomeAggregate":
				switch d := event.Data.(type) {
				case *SomeData:
					if d.A != data.A {
						t.Fatalf("wrong value in event.Data.A")
					}
				}
			default:
				t.Error("wrong aggregate type")
			}
		})
	}
}

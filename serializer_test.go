package eventsourcing_test

import (
	"encoding/json"
	"fmt"
	"github.com/hallgren/eventsourcing"
	"reflect"
	"sync"
	"testing"
)

func initSerializers(t *testing.T) []*eventsourcing.Serializer {
	var result []*eventsourcing.Serializer
	s := eventsourcing.NewSerializer(json.Marshal, json.Unmarshal)
	err := s.RegisterTypes(&SomeAggregate{}, func() interface{} { return &SomeData{} }, func() interface{} { return &SomeData2{} })
	if err != nil {
		t.Fatalf("could not register aggregate events %v", err)
	}
	result = append(result, s)
	return result
}

type SomeAggregate struct{
	eventsourcing.AggregateRoot
}

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
	for _, s := range serializers {
		t.Run(reflect.TypeOf(s).Elem().Name(), func(t *testing.T) {
			d, err := s.Marshal(data)
			if err != nil {
				t.Fatalf("could not Marshal data, %v", err)
			}
			fmt.Println(string(d))

			f, ok := s.Type("SomeAggregate", "SomeData")
			if !ok {
				t.Fatal("could not find event type registered for SomeAggregate/SomeData")
			}
			data2 := f()
			err = s.Unmarshal(d, &data2)
			if err != nil {
				t.Fatalf("Could not Unmarshal data, %v", err)
			}

			/*
				if data2.A != data.A {
					t.Fatalf("wrong value in A expected: %d, actual: %d", data.A, data2.A)
				}
			*/
			m, err := s.Marshal(metaData)
			if err != nil {
				t.Fatalf("could not Marshal metadata, %v", err)
			}
			metaData2 := map[string]interface{}{}
			err = s.Unmarshal(m, &metaData2)
			if err != nil {
				t.Fatalf("Could not Unmarshal metadata, %v", err)
			}
			if metaData["foo"] != metaData2["foo"] {
				t.Fatalf("wrong value in metadata key foo expected: bar, actual: %v", metaData2["foo"])
			}
		})
	}
}

func TestConcurrentUnmarshal(t *testing.T) {
	serializers := initSerializers(t)
	metaData["foo"] = "bar"
	for _, s := range serializers {
		d, err := s.Marshal(data)
		if err != nil {
			t.Fatalf("could not Marshal data, %v", err)
		}
		wg := sync.WaitGroup{}
		wg.Add(10)
		for i := 0; i < 10; i++ {
			go func(j int) {
				defer wg.Done()
				f, ok := s.Type("SomeAggregate", "SomeData")
				if !ok {
					t.Fatal("could not find event type registered for SomeAggregate/SomeData")
				}
				dataOut := f()
				err2 := s.Unmarshal(d, &dataOut)
				if err2 != nil {
					t.Errorf("Could not Unmarshal data, %v", err2)
				}
				switch dataOut.(type) {
				case *SomeData:
				default:
					t.Errorf("wrong type")
				}
			}(i)
		}
		wg.Wait()
	}
}

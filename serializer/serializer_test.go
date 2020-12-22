package serializer_test

import (
	"encoding/json"
	"fmt"
	"github.com/hallgren/eventsourcing"
	"github.com/hallgren/eventsourcing/serializer"
	"reflect"
	"testing"
)

func initSerializers(t *testing.T) []*serializer.Handler {
	var result = []*serializer.Handler{}
	s := serializer.New(json.Marshal, json.Unmarshal)
	err := s.Register(&SomeAggregate{}, &SomeData{}, &SomeData2{})
	if err != nil {
		t.Fatalf("could not register aggregate events %v", err)
	}
	result = append(result, s)
	return result
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
	for _, s := range serializers {
		t.Run(reflect.TypeOf(s).Elem().Name(), func(t *testing.T) {
			d,err := s.Marshal(data)
			if err != nil {
				t.Fatalf("could not Marshal data, %v", err)
			}
			fmt.Println(string(d))

			data2 := s.EventStruct("SomeAggregate", "SomeDate")
			err = s.Unmarshal(d, &data2)
			if err != nil {
				t.Fatalf("Could not Unmarshal data, %v", err)
			}
			/*
			if data2.A != data.A {
				t.Fatalf("wrong value in A expected: %d, actual: %d", data.A, data2.A)
			}
			*/
			m,err:= s.Marshal(metaData)
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

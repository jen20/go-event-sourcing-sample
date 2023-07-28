package core_test

import (
	"errors"
	"testing"
	"time"

	"github.com/hallgren/eventsourcing/core"
)

var aggregateType = "FrequentFlierAccount"
var timestamp = time.Now()

func testEvents(aggregateID string) []core.Event {
	metadata := make(map[string]interface{})
	metadata["test"] = "hello"
	history := []core.Event{
		{AggregateID: aggregateID, Version: 1, AggregateType: aggregateType, Timestamp: timestamp, Reason: "FrequentFlierAccountCreated", Data: []byte{}, Metadata: []byte{}},
		{AggregateID: aggregateID, Version: 2, AggregateType: aggregateType, Timestamp: timestamp, Reason: "StatusMatched", Data: []byte{}, Metadata: []byte{}},
		{AggregateID: aggregateID, Version: 3, AggregateType: aggregateType, Timestamp: timestamp, Reason: "FlightTaken", Data: []byte{}, Metadata: []byte{}},
		{AggregateID: aggregateID, Version: 4, AggregateType: aggregateType, Timestamp: timestamp, Reason: "FlightTaken", Data: []byte{}, Metadata: []byte{}},
		{AggregateID: aggregateID, Version: 5, AggregateType: aggregateType, Timestamp: timestamp, Reason: "FlightTaken", Data: []byte{}, Metadata: []byte{}},
		{AggregateID: aggregateID, Version: 6, AggregateType: aggregateType, Timestamp: timestamp, Reason: "FlightTaken", Data: []byte{}, Metadata: []byte{}},
	}
	return history
}

func TestValidate(t *testing.T) {
	err := core.ValidateEvents("123", 0, testEvents("123"))
	if err != nil {
		t.Fatal(err)
	}
}

func TestVersionAlreadySaved(t *testing.T) {
	err := core.ValidateEvents("123", 1, testEvents("123"))
	if !errors.Is(err, core.ErrConcurrency) {
		t.Fatal(err)
	}
}

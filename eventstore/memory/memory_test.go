package memory_test

import (
	"github.com/hallgren/eventsourcing"
	"github.com/hallgren/eventsourcing/eventstore/memory"
	"testing"
	"time"
)

// Status represents the Red, Silver or Gold tier level of a FrequentFlierAccount
type Status int

const (
	StatusRed    Status = iota
	StatusSilver Status = iota
	StatusGold   Status = iota
)

type FrequentFlierAccountCreated struct {
	AccountId         string
	OpeningMiles      int
	OpeningTierPoints int
}

type StatusMatched struct {
	NewStatus Status
}

type FlightTaken struct {
	MilesAdded      int
	TierPointsAdded int
}

var aggregateID = eventsourcing.AggregateRootID("123")
var aggregateType = "FrequentFlierAccount"

func testEvents() []eventsourcing.Event {

	history := []eventsourcing.Event{
		{AggregateRootID: aggregateID, Version: 1, Reason: "FrequentFlierAccountCreated", AggregateType: aggregateType, Data: FrequentFlierAccountCreated{AccountId: "1234567", OpeningMiles: 10000, OpeningTierPoints: 0}},
		{AggregateRootID: aggregateID, Version: 2, Reason: "StatusMatched", AggregateType: aggregateType, Data: StatusMatched{NewStatus: StatusSilver}},
		{AggregateRootID: aggregateID, Version: 3, Reason: "FlightTaken", AggregateType: aggregateType, Data: FlightTaken{MilesAdded: 2525, TierPointsAdded: 5}},
		{AggregateRootID: aggregateID, Version: 4, Reason: "FlightTaken", AggregateType: aggregateType, Data: FlightTaken{MilesAdded: 2512, TierPointsAdded: 5}},
		{AggregateRootID: aggregateID, Version: 5, Reason: "FlightTaken", AggregateType: aggregateType, Data: FlightTaken{MilesAdded: 5600, TierPointsAdded: 5}},
		{AggregateRootID: aggregateID, Version: 6, Reason: "FlightTaken", AggregateType: aggregateType, Data: FlightTaken{MilesAdded: 3000, TierPointsAdded: 3}},
	}

	return history
}

func testEventsPartTwo() []eventsourcing.Event {

	history := []eventsourcing.Event{
		{AggregateRootID: aggregateID, Version: 7, Reason: "FlightTaken", AggregateType: aggregateType, Data: FlightTaken{MilesAdded: 5600, TierPointsAdded: 5}},
		{AggregateRootID: aggregateID, Version: 8, Reason: "FlightTaken", AggregateType: aggregateType, Data: FlightTaken{MilesAdded: 3000, TierPointsAdded: 3}},
	}

	return history
}

var aggregateIDOther = eventsourcing.AggregateRootID("666")

func testEventOtherAggregate() eventsourcing.Event {
	return eventsourcing.Event{AggregateRootID: aggregateIDOther, Version: 1, Reason: "FrequentFlierAccountCreated", AggregateType: aggregateType, Data: FrequentFlierAccountCreated{AccountId: "1234567", OpeningMiles: 10000, OpeningTierPoints: 0}}
}

func TestSaveAndGetEvents(t *testing.T) {
	eventStore := memory.Create()
	defer eventStore.Close()
	err := eventStore.Save(testEvents())
	if err != nil {
		t.Error(err)
	}

	fetchedEvents, _ := eventStore.Get(string(aggregateID), aggregateType)

	if len(fetchedEvents) != len(testEvents()) {
		t.Error("Wrong number of events returned")
	}

	if fetchedEvents[0].Version != testEvents()[0].Version {
		t.Error("Wrong events returned")
	}

	// Add more events to the same aggregate event stream
	err = eventStore.Save(testEventsPartTwo())
	if err != nil {
		t.Error(err)
	}

	fetchedEventsIncludingPartTwo, _ := eventStore.Get(string(aggregateID), aggregateType)

	if len(fetchedEventsIncludingPartTwo) != len(append(testEvents(), testEventsPartTwo()...)) {
		t.Error("Wrong number of events returned")
	}

	if fetchedEventsIncludingPartTwo[0].Version != testEvents()[0].Version {
		t.Error("Wrong events returned")
	}

}

func TestSaveEventsFromMoreThanOneAggregate(t *testing.T) {
	eventStore := memory.Create()
	defer eventStore.Close()

	invalidEvent := append(testEvents(), testEventOtherAggregate())

	err := eventStore.Save(invalidEvent)
	if err == nil {
		t.Error("Should not be able to save events that belongs to more than one aggregate")
	}
}

func TestSaveEventsFromMoreThanOneAggregateType(t *testing.T) {
	eventStore := memory.Create()
	defer eventStore.Close()

	events := testEvents()
	events[1].AggregateType = "OtherAggregateType"

	err := eventStore.Save(events)
	if err == nil {
		t.Error("Should not be able to save events that belongs to other aggregate type")
	}
}

func TestSaveEventsInWrongOrder(t *testing.T) {
	eventStore := memory.Create()
	defer eventStore.Close()

	events := append(testEvents(), testEvents()[0])

	err := eventStore.Save(events)
	if err == nil {
		t.Error("Should not be able to save events that are in wrong version order")
	}
}

func TestSaveEventsInWrongVersion(t *testing.T) {
	eventStore := memory.Create()
	defer eventStore.Close()

	events := testEventsPartTwo()

	err := eventStore.Save(events)
	if err == nil {
		t.Error("Should not be able to save events that are out of sync compared to the storage order")
	}
}

func TestSaveEventsWithEmptyReason(t *testing.T) {
	eventStore := memory.Create()
	defer eventStore.Close()

	events := testEvents()
	events[2].Reason = ""

	err := eventStore.Save(events)
	if err == nil {
		t.Error("Should not be able to save events with empty reason")
	}
}

func TestGetGlobalEvents(t *testing.T) {
	eventStore := memory.Create()
	defer eventStore.Close()

	events := testEvents()
	eventStore.Save(events)

	fetchedEvents := eventStore.GlobalGet(2, 2)

	if len(fetchedEvents) != 2 {
		t.Error("Fetched the wrong amount of events")
	}

	if fetchedEvents[0].Version != events[1].Version {
		t.Errorf("Fetched the wrong events %v %v", fetchedEvents, events[2].Version)
	}

}

func TestGetGlobalEventsNotExisting(t *testing.T) {
	eventStore := memory.Create()
	defer eventStore.Close()

	events := testEvents()
	eventStore.Save(events)

	fetchedEvents := eventStore.GlobalGet(100, 2)

	if len(fetchedEvents) != 0 {
		t.Error("Fetched none existing events")
	}

}

func TestEventStream(t *testing.T) {
	eventStore := memory.Create()
	defer eventStore.Close()
	stream := eventStore.EventStream()

	events := testEvents()
	err := eventStore.Save(events)
	if err != nil {
		t.Error("Could not save the events")
	}

	counter := 0
outer:
	for {
		select {
		// wait for changes
		case <-stream.Changes():
			// advance to next value
			stream.Next()
			counter++
		case <-time.After(10 * time.Millisecond):
			// The stream has 10 milli seconds to deliver the events
			break outer
		}
	}

	if counter != 6 {
		t.Errorf("Not all events was received from the stream, got %q", counter)
	}
}

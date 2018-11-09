package eventstore_test

import (
	"go-event-sourcing-sample/pkg/eventsourcing"
	"go-event-sourcing-sample/pkg/eventstore"

	"testing"
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
		eventsourcing.Event{AggregateRootID: aggregateID, Version: 1, Reason: "FrequentFlierAccountCreated", AggregateType: aggregateType, Data: FrequentFlierAccountCreated{AccountId: "1234567", OpeningMiles: 10000, OpeningTierPoints: 0}},
		eventsourcing.Event{AggregateRootID: aggregateID, Version: 2, Reason: "StatusMatched", AggregateType: aggregateType, Data: StatusMatched{NewStatus: StatusSilver}},
		eventsourcing.Event{AggregateRootID: aggregateID, Version: 3, Reason: "FlightTaken", AggregateType: aggregateType, Data: FlightTaken{MilesAdded: 2525, TierPointsAdded: 5}},
		eventsourcing.Event{AggregateRootID: aggregateID, Version: 4, Reason: "FlightTaken", AggregateType: aggregateType, Data: FlightTaken{MilesAdded: 2512, TierPointsAdded: 5}},
		eventsourcing.Event{AggregateRootID: aggregateID, Version: 5, Reason: "FlightTaken", AggregateType: aggregateType, Data: FlightTaken{MilesAdded: 5600, TierPointsAdded: 5}},
		eventsourcing.Event{AggregateRootID: aggregateID, Version: 6, Reason: "FlightTaken", AggregateType: aggregateType, Data: FlightTaken{MilesAdded: 3000, TierPointsAdded: 3}},
	}

	return history
}

func testEventsPartTwo() []eventsourcing.Event {

	history := []eventsourcing.Event{
		eventsourcing.Event{AggregateRootID: aggregateID, Version: 7, Reason: "FlightTaken", AggregateType: aggregateType, Data: FlightTaken{MilesAdded: 5600, TierPointsAdded: 5}},
		eventsourcing.Event{AggregateRootID: aggregateID, Version: 8, Reason: "FlightTaken", AggregateType: aggregateType, Data: FlightTaken{MilesAdded: 3000, TierPointsAdded: 3}},
	}

	return history
}

var aggregateIDOther = eventsourcing.AggregateRootID("666")

func testEventOtherAggregate() eventsourcing.Event {
	return eventsourcing.Event{AggregateRootID: aggregateIDOther, Version: 1, Reason: "FrequentFlierAccountCreated", AggregateType: aggregateType, Data: FrequentFlierAccountCreated{AccountId: "1234567", OpeningMiles: 10000, OpeningTierPoints: 0}}
}

func TestSaveAndGetEvents(t *testing.T) {
	eventstore := eventstore.CreateMemory()
	defer eventstore.Close()
	err := eventstore.Save(testEvents())
	if err != nil {
		t.Error(err)
	}

	fetchedEvents := eventstore.Get(string(aggregateID), aggregateType)

	if len(fetchedEvents) != len(testEvents()) {
		t.Error("Wrong number of events returned")
	}

	if fetchedEvents[0].Version != testEvents()[0].Version {
		t.Error("Wrong events returned")
	}

	// Add more events to the same aggregate event stream
	err = eventstore.Save(testEventsPartTwo())
	if err != nil {
		t.Error(err)
	}

	fetchedEventsIncludingPartTwo := eventstore.Get(string(aggregateID), aggregateType)

	if len(fetchedEventsIncludingPartTwo) != len(append(testEvents(), testEventsPartTwo()...)) {
		t.Error("Wrong number of events returned")
	}

	if fetchedEventsIncludingPartTwo[0].Version != testEvents()[0].Version {
		t.Error("Wrong events returned")
	}

}

func TestSaveEventsFromMoreThanOneAggregate(t *testing.T) {
	eventstore := eventstore.CreateMemory()
	defer eventstore.Close()

	invalidEvent := append(testEvents(), testEventOtherAggregate())

	err := eventstore.Save(invalidEvent)
	if err == nil {
		t.Error("Should not be able to save events that belongs to more than one aggregate")
	}
}

func TestSaveEventsFromMoreThanOneAggregateType(t *testing.T) {
	eventstore := eventstore.CreateMemory()
	defer eventstore.Close()

	events := testEvents()
	events[1].AggregateType = "OtherAggregateType"

	err := eventstore.Save(events)
	if err == nil {
		t.Error("Should not be able to save events that belongs to other aggregate type")
	}
}

func TestSaveEventsInWrongOrder(t *testing.T) {
	eventstore := eventstore.CreateMemory()
	defer eventstore.Close()

	events := append(testEvents(), testEvents()[0])

	err := eventstore.Save(events)
	if err == nil {
		t.Error("Should not be able to save events that are in wrong version order")
	}
}

func TestSaveEventsInWrongVersion(t *testing.T) {
	eventstore := eventstore.CreateMemory()
	defer eventstore.Close()

	events := testEventsPartTwo()

	err := eventstore.Save(events)
	if err == nil {
		t.Error("Should not be able to save events that are out of sync compared to the storage order")
	}
}

func TestSaveEventsWithEmptyReason(t *testing.T) {
	eventstore := eventstore.CreateMemory()
	defer eventstore.Close()

	events := testEvents()
	events[2].Reason = ""

	err := eventstore.Save(events)
	if err == nil {
		t.Error("Should not be able to save events with empty reason")
	}
}

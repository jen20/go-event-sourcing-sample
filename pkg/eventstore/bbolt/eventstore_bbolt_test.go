package bbolt_test

import (
	"go-event-sourcing-sample/pkg/eventsourcing"
	"go-event-sourcing-sample/pkg/eventstore/bbolt"
	"os"

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

var dbFile = "test.db"
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

func TestMain(m *testing.M) {
	_ = os.Remove(dbFile)
	errLevel := m.Run()
	os.Remove(dbFile)
	os.Exit(errLevel)
}

func TestSaveAndGetEvents(t *testing.T) {
	bolt := bbolt.MustOpenBBolt(dbFile)
	defer bolt.Close()
	err := bolt.Save(testEvents())
	if err != nil {
		t.Error(err)
	}

	fetchedEvents, err := bolt.Get(string(aggregateID), aggregateType)

	if len(fetchedEvents) != len(testEvents()) {
		t.Error("Wrong number of events returned")
	}

	if fetchedEvents[0].Version != testEvents()[0].Version {
		t.Error("Wrong events returned")
	}

	// Add more events to the same aggregate event stream
	err = bolt.Save(testEventsPartTwo())
	if err != nil {
		t.Error(err)
	}

	fetchedEventsIncludingPartTwo, err := bolt.Get(string(aggregateID), aggregateType)

	if len(fetchedEventsIncludingPartTwo) != len(append(testEvents(), testEventsPartTwo()...)) {
		t.Error("Wrong number of events returned")
	}

	if fetchedEventsIncludingPartTwo[0].Version != testEvents()[0].Version {
		t.Error("Wrong events returned")
	}

}

func TestSaveEventsFromMoreThanOneAggregate(t *testing.T) {
	eventstore := bbolt.MustOpenBBolt(dbFile)
	defer eventstore.Close()

	invalidEvent := append(testEvents(), testEventOtherAggregate())

	err := eventstore.Save(invalidEvent)
	if err == nil {
		t.Error("Should not be able to save events that belongs to more than one aggregate")
	}
}

func TestSaveEventsFromMoreThanOneAggregateType(t *testing.T) {
	eventstore := bbolt.MustOpenBBolt(dbFile)
	defer eventstore.Close()

	events := testEvents()
	events[1].AggregateType = "OtherAggregateType"

	err := eventstore.Save(events)
	if err == nil {
		t.Error("Should not be able to save events that belongs to other aggregate type")
	}
}

func TestSaveEventsInWrongOrder(t *testing.T) {
	eventstore := bbolt.MustOpenBBolt(dbFile)
	defer eventstore.Close()

	events := append(testEvents(), testEvents()[0])

	err := eventstore.Save(events)
	if err == nil {
		t.Error("Should not be able to save events that are in wrong version order")
	}
}

func TestSaveEventsInWrongVersion(t *testing.T) {
	eventstore := bbolt.MustOpenBBolt(dbFile)
	defer eventstore.Close()

	events := testEventsPartTwo()

	err := eventstore.Save(events)
	if err == nil {
		t.Error("Should not be able to save events that are out of sync compared to the storage order")
	}
}

func TestSaveEventsWithEmptyReason(t *testing.T) {
	eventstore := bbolt.MustOpenBBolt(dbFile)
	defer eventstore.Close()

	events := testEvents()
	events[2].Reason = ""

	err := eventstore.Save(events)
	if err == nil {
		t.Error("Should not be able to save events with empty reason")
	}
}

/*
func TestGetGlobalEvents(t *testing.T) {
	eventstore := bbolt.MustOpenBBolt(dbFile)
	defer eventstore.Close()

	events := testEvents()
	eventstore.Save(events)

	fetchedEvents := eventstore.GlobalGet(2, 2)

	if len(fetchedEvents) != 2 {
		t.Error("Fetched the wrong amount of events")
	}

	if fetchedEvents[0].Version != events[2].Version {
		t.Errorf("Fetched the wrong events %v %v", fetchedEvents, events[2].Version)
	}

}

func TestGetGlobalEventsNotExisting(t *testing.T) {
	eventstore := bbolt.MustOpenBBolt(dbFile)
	defer eventstore.Close()

	events := testEvents()
	eventstore.Save(events)

	fetchedEvents := eventstore.GlobalGet(100, 2)

	if len(fetchedEvents) != 0 {
		t.Error("Fetched none existing events")
	}

}
*/

package bbolt_test

import (
	"gitlab.se.axis.com/morganh/eventsourcing"
	"gitlab.se.axis.com/morganh/eventsourcing/eventstore/bbolt"
	"os"
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

var dbFile = "test.db"
var aggregateID = eventsourcing.AggregateRootID("123")
var aggregateType = "FrequentFlierAccount"

func testEventsWithID(aggregateID eventsourcing.AggregateRootID) []eventsourcing.Event {

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

func testEvents() []eventsourcing.Event {
	return testEventsWithID(aggregateID)
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

func TestMain(m *testing.M) {
	_ = os.Remove(dbFile)
	errLevel := m.Run()
	os.Remove(dbFile)
	os.Exit(errLevel)
}

func TestSaveAndGetEvents(t *testing.T) {
	os.Remove(dbFile)
	bolt := bbolt.MustOpenBBolt(dbFile)
	defer bolt.Close()
	defer os.Remove(dbFile)
	
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
	os.Remove(dbFile)
	eventStore := bbolt.MustOpenBBolt(dbFile)
	defer eventStore.Close()
	defer os.Remove(dbFile)

	invalidEvent := append(testEvents(), testEventOtherAggregate())

	err := eventStore.Save(invalidEvent)
	if err == nil {
		t.Error("Should not be able to save events that belongs to more than one aggregate")
	}
}

func TestSaveEventsFromMoreThanOneAggregateType(t *testing.T) {
	os.Remove(dbFile)
	eventStore := bbolt.MustOpenBBolt(dbFile)
	defer eventStore.Close()
	defer os.Remove(dbFile)

	events := testEvents()
	events[1].AggregateType = "OtherAggregateType"

	err := eventStore.Save(events)
	if err == nil {
		t.Error("Should not be able to save events that belongs to other aggregate type")
	}
}

func TestSaveEventsInWrongOrder(t *testing.T) {
	os.Remove(dbFile)
	eventStore := bbolt.MustOpenBBolt(dbFile)
	defer eventStore.Close()
	defer os.Remove(dbFile)

	events := append(testEvents(), testEvents()[0])

	err := eventStore.Save(events)
	if err == nil {
		t.Error("Should not be able to save events that are in wrong version order")
	}
}

func TestSaveEventsInWrongVersion(t *testing.T) {
	os.Remove(dbFile)
	eventStore := bbolt.MustOpenBBolt(dbFile)
	defer eventStore.Close()
	defer os.Remove(dbFile)

	events := testEventsPartTwo()

	err := eventStore.Save(events)
	if err == nil {
		t.Error("Should not be able to save events that are out of sync compared to the storage order")
	}
}

func TestSaveEventsWithEmptyReason(t *testing.T) {
	os.Remove(dbFile)
	eventStore := bbolt.MustOpenBBolt(dbFile)
	defer eventStore.Close()
	defer os.Remove(dbFile)

	events := testEvents()
	events[2].Reason = ""

	err := eventStore.Save(events)
	if err == nil {
		t.Error("Should not be able to save events with empty reason")
	}
}

func TestGetGlobalEvents(t *testing.T) {
	os.Remove(dbFile)
	eventStore := bbolt.MustOpenBBolt(dbFile)
	defer eventStore.Close()
	defer os.Remove(dbFile)

	events := testEvents()
	err := eventStore.Save(events)
	if err != nil {
		t.Error("Could not save the events")
	}

	fetchedEvents := eventStore.GlobalGet(2, 2)

	if len(fetchedEvents) != 2 {
		t.Error("Fetched the wrong amount of events")
	}

	if fetchedEvents[0].Version != events[1].Version {
		t.Errorf("Fetched the wrong events %v %v", fetchedEvents[0].Version, events[2].Version)
	}

}

func TestGetGlobalEventsNotExisting(t *testing.T) {
	os.Remove(dbFile)
	eventStore := bbolt.MustOpenBBolt(dbFile)
	defer eventStore.Close()
	defer os.Remove(dbFile)

	events := testEvents()
	err := eventStore.Save(events)
	if err != nil {
		t.Error("Could not save the events")
	}

	fetchedEvents := eventStore.GlobalGet(100, 2)

	if len(fetchedEvents) != 0 {
		t.Error("Fetched none existing events")
	}

}

func TestEventStream(t *testing.T) {
	os.Remove(dbFile)
	eventStore := bbolt.MustOpenBBolt(dbFile)
	defer eventStore.Close()
	defer os.Remove(dbFile)
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
		case <-time.After(10*time.Millisecond):
			// The stream has 10 milliseconds to deliver the events
			break outer
		}
	}

	if counter != 6 {
		t.Errorf("Not all events was received from the stream, got %q", counter)
	}
}
package sql_test

import (
	sqldriver "database/sql"
	"github.com/hallgren/eventsourcing"
	"github.com/hallgren/eventsourcing/eventstore/sql"
	"github.com/hallgren/eventsourcing/serializer/json"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
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

var dbFile = "test.sql"
var aggregateID = eventsourcing.AggregateRootID("123")
var aggregateType = "FrequentFlierAccount"
var jsonSerializer = json.New()

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

func TestSaveAndGetEvents(t *testing.T) {
	os.Remove(dbFile)
	db, err := sqldriver.Open("sqlite3", dbFile)
	if err != nil {
		t.Fatalf("Could not open sqlit3 database %v", err)
	}
	s := sql.Open(*db, jsonSerializer)
	defer s.Close()
	defer os.Remove(dbFile)

	err = s.Save(testEvents())
	if err != nil {
		t.Fatal(err)
	}

	fetchedEvents, err := s.Get(string(aggregateID), aggregateType)

	if len(fetchedEvents) != len(testEvents()) {
		t.Fatal("Wrong number of events returned")
	}

	if fetchedEvents[0].Version != testEvents()[0].Version {
		t.Fatal("Wrong events returned")
	}

	// Add more events to the same aggregate event stream
	err = s.Save(testEventsPartTwo())
	if err != nil {
		t.Fatal(err)
	}

	fetchedEventsIncludingPartTwo, err := s.Get(string(aggregateID), aggregateType)

	if len(fetchedEventsIncludingPartTwo) != len(append(testEvents(), testEventsPartTwo()...)) {
		t.Error("Wrong number of events returned")
	}

	if fetchedEventsIncludingPartTwo[0].Version != testEvents()[0].Version {
		t.Error("Wrong events returned")
	}
}

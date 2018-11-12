package bbolt_test

import (
	"fmt"
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

func testEventsWithID(aggregateID eventsourcing.AggregateRootID) []eventsourcing.Event {

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

func testEvents() []eventsourcing.Event {
	return testEventsWithID(aggregateID)
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
// Person aggregate
type Person struct {
	aggregateRoot eventsourcing.AggregateRoot
	name          string
	age           int
}

// PersonCreated event
type PersonCreated struct {
	name       string
	initialAge int
}

// AgedOneYear event
type AgedOneYear struct {
}

// transition the person state dependent on the events
func (person *Person) transition(event eventsourcing.Event) {
	switch e := event.Data.(type) {

	case PersonCreated:
		person.age = e.initialAge
		person.name = e.name

	case AgedOneYear:
		person.age += 1
	}
}

// CreatePersonWithID constructor for the Person that sets the aggregate id from the outside
func CreatePersonWithID(id, name string) (*Person, error) {

	if name == "" {
		return nil, fmt.Errorf("name can't be blank")
	}

	person := &Person{}
	person.aggregateRoot.SetID(id)
	person.aggregateRoot.TrackChange(*person, PersonCreated{name: name, initialAge: 0}, person.transition)
	return person, nil
}

// NewFrequentFlierAccountFromHistory creates a FrequentFlierAccount given a history
// of the changes which have occurred for that account.
func CreatePersonFromHistory(events []eventsourcing.Event) *Person {
	state := &Person{}
	state.aggregateRoot.BuildFromHistory(events, state.transition)
	return state
}

// GrowOlder command
func (person *Person) GrowOlder() {
	person.aggregateRoot.TrackChange(*person, AgedOneYear{}, person.transition)

}

// Benchmark the time it takes to retrieve a list of keys for an entity type
func BenchmarkFetch101Events(b *testing.B) {
	os.Remove(dbFile)
	eventstore := bbolt.MustOpenBBolt(dbFile)
	defer eventstore.Close()

	person, err := CreatePersonWithID("123", "kalle")
	if err != nil {
		b.Error(err)
	}

	for i := 0; i < 100; i++ {
		person.GrowOlder()
	}

	err = eventstore.Save(person.aggregateRoot.Changes())
	if err != nil {
		b.Error(err)
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		events, _ := eventstore.Get("123", person.aggregateRoot.Changes()[0].AggregateType)
		CreatePersonFromHistory(events)
	}
}

package eventstore_test

import (
	sqldriver "database/sql"
	"fmt"
	"github.com/hallgren/eventsourcing"
	"github.com/hallgren/eventsourcing/eventstore/bbolt"
	"github.com/hallgren/eventsourcing/eventstore/memory"
	s "github.com/hallgren/eventsourcing/eventstore/sql"
	"github.com/hallgren/eventsourcing/serializer/json"
	"github.com/hallgren/eventsourcing/serializer/unsafe"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"reflect"
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
var aggregateID2 = eventsourcing.AggregateRootID("321")
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

func sql() (*s.SQL, func(), error) {
	dbFile := "test.sql"
	os.Remove(dbFile)
	db, err := sqldriver.Open("sqlite3", dbFile)
	if err != nil {
		return nil, nil, fmt.Errorf("Could not open sqlit3 database %v", err)
	}
	err = db.Ping()
	if err != nil {
		return nil, nil, fmt.Errorf("could not ping database %v", err)
	}
	s := s.Open(*db, jsonSerializer)
	err = s.Migrate()
	if err != nil {
		return nil, nil, fmt.Errorf("could not migrate database %v", err)
	}
	return s, func() {
		s.Close()
		os.Remove(dbFile)
	}, nil
}

func bolt() (*bbolt.BBolt, func()) {
	dbFile := "bolt.db"
	b := bbolt.MustOpenBBolt(dbFile, jsonSerializer)
	return b, func() {
		b.Close()
		os.Remove(dbFile)
	}
}

func initEventStores() ([]eventstore, func(), error) {
	sqlEventStore, closer, err := sql()
	if err != nil {
		return nil, nil, err
	}
	boltEventStore, closerBolt := bolt()
	eventstores := []eventstore{sqlEventStore, boltEventStore, memory.Create(unsafe.New())}
	return eventstores, func() {
		closer()
		closerBolt()
	}, nil
}

type eventstore interface {
	Save(events []eventsourcing.Event) error
	Get(id string, aggregateType string, afterVersion eventsourcing.Version) ([]eventsourcing.Event, error)
	GlobalGet(start, count int) []eventsourcing.Event
}

func TestSaveAndGetEvents(t *testing.T) {
	stores, closer, err := initEventStores()
	if err != nil {
		t.Fatalf("Could not init event stores %v", err)
	}
	defer closer()

	for _, es := range stores {
		t.Run(reflect.TypeOf(es).Elem().Name(), func(t *testing.T) {
			err := es.Save(testEvents())
			if err != nil {
				t.Fatal(err)
			}

			fetchedEvents, err := es.Get(string(aggregateID), aggregateType, 0)

			if len(fetchedEvents) != len(testEvents()) {
				t.Fatal("Wrong number of events returned")
			}

			if fetchedEvents[0].Version != testEvents()[0].Version {
				t.Fatal("Wrong events returned")
			}

			// Add more events to the same aggregate event stream
			err = es.Save(testEventsPartTwo())
			if err != nil {
				t.Fatal(err)
			}

			fetchedEventsIncludingPartTwo, err := es.Get(string(aggregateID), aggregateType, 0)
			if err != nil {
				t.Fatalf("repository Get returned error: %v", err)
			}

			if len(fetchedEventsIncludingPartTwo) != len(append(testEvents(), testEventsPartTwo()...)) {
				t.Error("Wrong number of events returned")
			}

			if fetchedEventsIncludingPartTwo[0].Version != testEvents()[0].Version {
				t.Error("Wrong events returned")
			}
		})

	}
}

func TestGetEventsAfterVersion(t *testing.T) {
	stores, closer, err := initEventStores()
	if err != nil {
		t.Fatalf("Could not init event stores %v", err)
	}
	defer closer()

	for _, es := range stores {
		t.Run(reflect.TypeOf(es).Elem().Name(), func(t *testing.T) {
			err := es.Save(testEvents())
			if err != nil {
				t.Fatal(err)
			}

			fetchedEvents, err := es.Get(string(aggregateID), aggregateType, 1)

			// Should return one less event
			if len(fetchedEvents) != len(testEvents())-1 {
				t.Fatal("Wrong number of events returned")
			}

			// first event version should be 2
			if fetchedEvents[0].Version != 2 {
				t.Fatal("Wrong events returned")
			}
		})

	}
}

func TestSaveEventsFromMoreThanOneAggregate(t *testing.T) {
	stores, closer, err := initEventStores()
	if err != nil {
		t.Fatalf("Could not init event stores %v", err)
	}
	defer closer()

	for _, es := range stores {
		t.Run(reflect.TypeOf(es).Elem().Name(), func(t *testing.T) {
			invalidEvent := append(testEvents(), testEventOtherAggregate())

			err := es.Save(invalidEvent)
			if err == nil {
				t.Error("Should not be able to save events that belongs to more than one aggregate")
			}
		})
	}
}

func TestSaveEventsFromMoreThanOneAggregateType(t *testing.T) {
	stores, closer, err := initEventStores()
	if err != nil {
		t.Fatalf("Could not init event stores %v", err)
	}
	defer closer()

	events := testEvents()
	events[1].AggregateType = "OtherAggregateType"

	for _, es := range stores {
		t.Run(reflect.TypeOf(es).Elem().Name(), func(t *testing.T) {
			err := es.Save(events)
			if err == nil {
				t.Error("Should not be able to save events that belongs to other aggregate type")
			}
		})
	}
}

func TestSaveEventsInWrongOrder(t *testing.T) {
	stores, closer, err := initEventStores()
	if err != nil {
		t.Fatalf("Could not init event stores %v", err)
	}
	defer closer()

	events := append(testEvents(), testEvents()[0])
	for _, es := range stores {
		t.Run(reflect.TypeOf(es).Elem().Name(), func(t *testing.T) {
			err := es.Save(events)
			if err == nil {
				t.Error("Should not be able to save events that are in wrong version order")
			}
		})
	}
}

func TestSaveEventsInWrongVersion(t *testing.T) {
	stores, closer, err := initEventStores()
	if err != nil {
		t.Fatalf("Could not init event stores %v", err)
	}
	defer closer()

	events := testEventsPartTwo()
	for _, es := range stores {
		t.Run(reflect.TypeOf(es).Elem().Name(), func(t *testing.T) {
			err := es.Save(events)
			if err == nil {
				t.Error("Should not be able to save events that are out of sync compared to the storage order")
			}
		})
	}
}

func TestSaveEventsWithEmptyReason(t *testing.T) {
	stores, closer, err := initEventStores()
	if err != nil {
		t.Fatalf("Could not init event stores %v", err)
	}
	defer closer()

	events := testEvents()
	events[2].Reason = ""
	for _, es := range stores {
		t.Run(reflect.TypeOf(es).Elem().Name(), func(t *testing.T) {
			err := es.Save(events)
			if err == nil {
				t.Error("Should not be able to save events with empty reason")
			}
		})
	}
}

func TestGetGlobalEvents(t *testing.T) {
	stores, closer, err := initEventStores()
	if err != nil {
		t.Fatalf("Could not init event stores %v", err)
	}
	defer closer()

	events := testEvents()
	for _, es := range stores {
		t.Run(reflect.TypeOf(es).Elem().Name(), func(t *testing.T) {

			err := es.Save(events)
			if err != nil {
				t.Fatalf("%v could not save the events", err)
			}
			_ = es.Save([]eventsourcing.Event{{AggregateRootID: aggregateID2, Version: 1, Reason: "FrequentFlierAccountCreated", AggregateType: aggregateType, Data: FrequentFlierAccountCreated{AccountId: "1234567", OpeningMiles: 10000, OpeningTierPoints: 0}}})

			fetchedEvents := es.GlobalGet(6, 2)

			if len(fetchedEvents) != 2 {
				t.Fatalf("Fetched the wrong amount of events")
			}

			if fetchedEvents[0].Version != events[5].Version {
				t.Fatalf("%v fetched the wrong events %v %v", es, fetchedEvents[0].Version, events[2].Version)
			}
		})
	}
}

func TestGetGlobalEventsNotExisting(t *testing.T) {
	stores, closer, err := initEventStores()
	if err != nil {
		t.Fatalf("Could not init event stores %v", err)
	}
	defer closer()
	events := testEvents()
	for _, es := range stores {
		t.Run(reflect.TypeOf(es).Elem().Name(), func(t *testing.T) {
			err := es.Save(events)
			if err != nil {
				t.Error("Could not save the events")
			}

			fetchedEvents := es.GlobalGet(100, 2)

			if len(fetchedEvents) != 0 {
				t.Error("Fetched none existing events")
			}
		})
	}
}

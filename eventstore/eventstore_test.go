package eventstore_test

import (
	sqldriver "database/sql"
	"errors"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/hallgren/eventsourcing"
	"github.com/hallgren/eventsourcing/eventstore/bbolt"
	"github.com/hallgren/eventsourcing/eventstore/memory"
	s "github.com/hallgren/eventsourcing/eventstore/sql"
	"github.com/hallgren/eventsourcing/serializer/json"
	"github.com/hallgren/eventsourcing/serializer/unsafe"
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

var aggregateID = "123"
var aggregateID2 = "321"
var aggregateType = "FrequentFlierAccount"
var jsonSerializer = json.New()

func testEventsWithID(aggregateID string) []eventsourcing.Event {
	metaData := make(map[string]interface{})
	metaData["test"] = "hello"
	history := []eventsourcing.Event{
		{AggregateRootID: aggregateID, Version: 1, Reason: "FrequentFlierAccountCreated", AggregateType: aggregateType, Data: FrequentFlierAccountCreated{AccountId: "1234567", OpeningMiles: 10000, OpeningTierPoints: 0}, MetaData: metaData},
		{AggregateRootID: aggregateID, Version: 2, Reason: "StatusMatched", AggregateType: aggregateType, Data: StatusMatched{NewStatus: StatusSilver}, MetaData: metaData},
		{AggregateRootID: aggregateID, Version: 3, Reason: "FlightTaken", AggregateType: aggregateType, Data: FlightTaken{MilesAdded: 2525, TierPointsAdded: 5}, MetaData: metaData},
		{AggregateRootID: aggregateID, Version: 4, Reason: "FlightTaken", AggregateType: aggregateType, Data: FlightTaken{MilesAdded: 2512, TierPointsAdded: 5}, MetaData: metaData},
		{AggregateRootID: aggregateID, Version: 5, Reason: "FlightTaken", AggregateType: aggregateType, Data: FlightTaken{MilesAdded: 5600, TierPointsAdded: 5}, MetaData: metaData},
		{AggregateRootID: aggregateID, Version: 6, Reason: "FlightTaken", AggregateType: aggregateType, Data: FlightTaken{MilesAdded: 3000, TierPointsAdded: 3}, MetaData: metaData},
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

var aggregateIDOther = "666"

func testEventOtherAggregate() eventsourcing.Event {
	return eventsourcing.Event{AggregateRootID: aggregateIDOther, Version: 1, Reason: "FrequentFlierAccountCreated", AggregateType: aggregateType, Data: FrequentFlierAccountCreated{AccountId: "1234567", OpeningMiles: 10000, OpeningTierPoints: 0}}
}

func sql() (*s.SQL, func(), error) {
	dbFile := "test.sql"
	os.Remove(dbFile)
	db, err := sqldriver.Open("sqlite3", dbFile)
	if err != nil {
		return nil, nil, errors.New(fmt.Sprintf("could not open sqlit3 database %v", err))
	}
	err = db.Ping()
	if err != nil {
		return nil, nil, errors.New(fmt.Sprintf("could not ping database %v", err))
	}
	s := s.Open(*db, jsonSerializer)
	err = s.Migrate()
	if err != nil {
		return nil, nil, errors.New(fmt.Sprintf("could not migrate database %v", err))
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
}

func TestSaveAndGetEvents(t *testing.T) {
	stores, closer, err := initEventStores()
	if err != nil {
		t.Fatalf("could not init event stores %v", err)
	}
	defer closer()

	for _, es := range stores {
		t.Run(reflect.TypeOf(es).Elem().Name(), func(t *testing.T) {
			err := es.Save(testEvents())
			if err != nil {
				t.Fatal(err)
			}

			fetchedEvents, err := es.Get(string(aggregateID), aggregateType, 0)
			if err != nil {
				t.Fatal(err)
			}

			if len(fetchedEvents) != len(testEvents()) {
				t.Fatal("wrong number of events returned")
			}

			if fetchedEvents[0].Version != testEvents()[0].Version {
				t.Fatal("wrong events returned")
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
				t.Error("wrong number of events returned")
			}

			if fetchedEventsIncludingPartTwo[0].Version != testEvents()[0].Version {
				t.Error("wrong event version returned")
			}

			if fetchedEventsIncludingPartTwo[0].AggregateRootID != testEvents()[0].AggregateRootID {
				t.Error("wrong event aggregateID returned")
			}

			if fetchedEventsIncludingPartTwo[0].AggregateType != testEvents()[0].AggregateType {
				t.Error("wrong event aggregateType returned")
			}

			if fetchedEventsIncludingPartTwo[0].Reason != testEvents()[0].Reason {
				t.Error("wrong event aggregateType returned")
			}

			if fetchedEventsIncludingPartTwo[0].MetaData["test"] != "hello" {
				t.Error("wrong event meta data returned")
			}
		})

	}
}

func TestGetEventsAfterVersion(t *testing.T) {
	stores, closer, err := initEventStores()
	if err != nil {
		t.Fatalf("could not init event stores %v", err)
	}
	defer closer()

	for _, es := range stores {
		t.Run(reflect.TypeOf(es).Elem().Name(), func(t *testing.T) {
			err := es.Save(testEvents())
			if err != nil {
				t.Fatal(err)
			}

			fetchedEvents, err := es.Get(string(aggregateID), aggregateType, 1)
			if err != nil {
				t.Fatal(err)
			}

			// Should return one less event
			if len(fetchedEvents) != len(testEvents())-1 {
				t.Fatal("wrong number of events returned")
			}
			fmt.Println(fetchedEvents)
			// first event version should be 2
			if fetchedEvents[0].Version != 2 {
				t.Fatal("wrong events returned")
			}
		})

	}
}

func TestSaveEventsFromMoreThanOneAggregate(t *testing.T) {
	stores, closer, err := initEventStores()
	if err != nil {
		t.Fatalf("could not init event stores %v", err)
	}
	defer closer()

	for _, es := range stores {
		t.Run(reflect.TypeOf(es).Elem().Name(), func(t *testing.T) {
			invalidEvent := append(testEvents(), testEventOtherAggregate())

			err := es.Save(invalidEvent)
			if err == nil {
				t.Error("should not be able to save events that belongs to more than one aggregate")
			}
		})
	}
}

func TestSaveEventsFromMoreThanOneAggregateType(t *testing.T) {
	stores, closer, err := initEventStores()
	if err != nil {
		t.Fatalf("could not init event stores %v", err)
	}
	defer closer()

	events := testEvents()
	events[1].AggregateType = "OtherAggregateType"

	for _, es := range stores {
		t.Run(reflect.TypeOf(es).Elem().Name(), func(t *testing.T) {
			err := es.Save(events)
			if err == nil {
				t.Error("should not be able to save events that belongs to other aggregate type")
			}
		})
	}
}

func TestSaveEventsInWrongOrder(t *testing.T) {
	stores, closer, err := initEventStores()
	if err != nil {
		t.Fatalf("could not init event stores %v", err)
	}
	defer closer()

	events := append(testEvents(), testEvents()[0])
	for _, es := range stores {
		t.Run(reflect.TypeOf(es).Elem().Name(), func(t *testing.T) {
			err := es.Save(events)
			if err == nil {
				t.Error("should not be able to save events that are in wrong version order")
			}
		})
	}
}

func TestSaveEventsInWrongVersion(t *testing.T) {
	stores, closer, err := initEventStores()
	if err != nil {
		t.Fatalf("could not init event stores %v", err)
	}
	defer closer()

	events := testEventsPartTwo()
	for _, es := range stores {
		t.Run(reflect.TypeOf(es).Elem().Name(), func(t *testing.T) {
			err := es.Save(events)
			if err == nil {
				t.Error("should not be able to save events that are out of sync compared to the storage order")
			}
		})
	}
}

func TestSaveEventsWithEmptyReason(t *testing.T) {
	stores, closer, err := initEventStores()
	if err != nil {
		t.Fatalf("could not init event stores %v", err)
	}
	defer closer()

	events := testEvents()
	events[2].Reason = ""
	for _, es := range stores {
		t.Run(reflect.TypeOf(es).Elem().Name(), func(t *testing.T) {
			err := es.Save(events)
			if err == nil {
				t.Error("should not be able to save events with empty reason")
			}
		})
	}
}

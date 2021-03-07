package suite

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/hallgren/eventsourcing"
)

type eventstoreFunc = func() (eventsourcing.EventStore, func(), error)

func Test(t *testing.T, esFunc eventstoreFunc) {
	tests := []struct {
		title string
		run   func(t *testing.T, es eventsourcing.EventStore)
	}{
		{"should save and get events", saveAndGetEvents},
		{"should get events after version", getEventsAfterVersion},
		{"should not save events from different aggregates", saveEventsFromMoreThanOneAggregate},
		{"should not save events from different aggregate types", saveEventsFromMoreThanOneAggregateType},
		{"should not save events in wrong order", saveEventsInWrongOrder},
		{"should not save events in wrong version", saveEventsInWrongVersion},
		{"should not save event with no reason", saveEventsWithEmptyReason},
		{"should save and get event concurrently", saveAndGetEventsConcurrently},
		{"should return error when no events", getErrWhenNoEvents},
	}
	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			es, closeFunc, err := esFunc()
			if err != nil {
				t.Fatal(err)
			}
			test.run(t, es)
			closeFunc()
		})
	}
}

// Status represents the Red, Silver or Gold tier level of a FrequentFlierAccount
type Status int

const (
	StatusRed    Status = iota
	StatusSilver Status = iota
	StatusGold   Status = iota
)

type FrequentFlierAccount struct {
	eventsourcing.AggregateRoot
}

func (f *FrequentFlierAccount) Transition(e eventsourcing.Event) {}

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
var aggregateIDOther = "666"
var timestamp = time.Now()

func testEventsWithID(aggregateID string) []eventsourcing.Event {
	metaData := make(map[string]interface{})
	metaData["test"] = "hello"
	history := []eventsourcing.Event{
		{AggregateID: aggregateID, Version: 1, Reason: "FrequentFlierAccountCreated", AggregateType: aggregateType, Timestamp: timestamp, Data: &FrequentFlierAccountCreated{AccountId: "1234567", OpeningMiles: 10000, OpeningTierPoints: 0}, MetaData: metaData},
		{AggregateID: aggregateID, Version: 2, Reason: "StatusMatched", AggregateType: aggregateType, Timestamp: timestamp, Data: &StatusMatched{NewStatus: StatusSilver}, MetaData: metaData},
		{AggregateID: aggregateID, Version: 3, Reason: "FlightTaken", AggregateType: aggregateType, Timestamp: timestamp, Data: &FlightTaken{MilesAdded: 2525, TierPointsAdded: 5}, MetaData: metaData},
		{AggregateID: aggregateID, Version: 4, Reason: "FlightTaken", AggregateType: aggregateType, Timestamp: timestamp, Data: &FlightTaken{MilesAdded: 2512, TierPointsAdded: 5}, MetaData: metaData},
		{AggregateID: aggregateID, Version: 5, Reason: "FlightTaken", AggregateType: aggregateType, Timestamp: timestamp, Data: &FlightTaken{MilesAdded: 5600, TierPointsAdded: 5}, MetaData: metaData},
		{AggregateID: aggregateID, Version: 6, Reason: "FlightTaken", AggregateType: aggregateType, Timestamp: timestamp, Data: &FlightTaken{MilesAdded: 3000, TierPointsAdded: 3}, MetaData: metaData},
	}
	return history
}

func testEvents() []eventsourcing.Event {
	return testEventsWithID(aggregateID)
}

func testEventsPartTwo() []eventsourcing.Event {
	history := []eventsourcing.Event{
		{AggregateID: aggregateID, Version: 7, Reason: "FlightTaken", AggregateType: aggregateType, Timestamp: timestamp, Data: &FlightTaken{MilesAdded: 5600, TierPointsAdded: 5}},
		{AggregateID: aggregateID, Version: 8, Reason: "FlightTaken", AggregateType: aggregateType, Timestamp: timestamp, Data: &FlightTaken{MilesAdded: 3000, TierPointsAdded: 3}},
	}
	return history
}

func testEventOtherAggregate() eventsourcing.Event {
	return eventsourcing.Event{AggregateID: aggregateIDOther, Version: 1, Reason: "FrequentFlierAccountCreated", AggregateType: aggregateType, Timestamp: timestamp, Data: &FrequentFlierAccountCreated{AccountId: "1234567", OpeningMiles: 10000, OpeningTierPoints: 0}}
}

func saveAndGetEvents(t *testing.T, es eventsourcing.EventStore) {
	err := es.Save(testEvents())
	if err != nil {
		t.Fatal(err)
	}

	fetchedEvents, err := es.Get(aggregateID, aggregateType, 0)
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

	fetchedEventsIncludingPartTwo, err := es.Get(aggregateID, aggregateType, 0)
	if err != nil {
		t.Fatalf("repository Get returned error: %v", err)
	}

	if len(fetchedEventsIncludingPartTwo) != len(append(testEvents(), testEventsPartTwo()...)) {
		t.Error("wrong number of events returned")
	}

	if fetchedEventsIncludingPartTwo[0].Version != testEvents()[0].Version {
		t.Error("wrong event version returned")
	}

	if fetchedEventsIncludingPartTwo[0].AggregateID != testEvents()[0].AggregateID {
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

	if fetchedEventsIncludingPartTwo[0].Timestamp.Format(time.RFC3339) != timestamp.Format(time.RFC3339) {
		t.Errorf("wrong timestamp exp: %s got: %s", fetchedEventsIncludingPartTwo[0].Timestamp.Format(time.RFC3339), timestamp.Format(time.RFC3339))
	}

	data, ok := fetchedEventsIncludingPartTwo[0].Data.(*FrequentFlierAccountCreated)
	if !ok {
		t.Errorf("wrong type in Data")
	}

	if data.OpeningMiles != 10000 {
		t.Errorf("wrong OpeningMiles %d", data.OpeningMiles)
	}
}

func getEventsAfterVersion(t *testing.T, es eventsourcing.EventStore) {
	err := es.Save(testEvents())
	if err != nil {
		t.Fatal(err)
	}

	fetchedEvents, err := es.Get(aggregateID, aggregateType, 1)
	if err != nil {
		t.Fatal(err)
	}

	// Should return one less event
	if len(fetchedEvents) != len(testEvents())-1 {
		t.Fatalf("wrong number of events returned exp: %d, got:%d", len(fetchedEvents), len(testEvents())-1)
	}
	// first event version should be 2
	if fetchedEvents[0].Version != 2 {
		t.Fatal("wrong events returned")
	}
}

func saveEventsFromMoreThanOneAggregate(t *testing.T, es eventsourcing.EventStore) {
	invalidEvent := append(testEvents(), testEventOtherAggregate())
	err := es.Save(invalidEvent)
	if err == nil {
		t.Error("should not be able to save events that belongs to more than one aggregate")
	}
}

func saveEventsFromMoreThanOneAggregateType(t *testing.T, es eventsourcing.EventStore) {
	events := testEvents()
	events[1].AggregateType = "OtherAggregateType"

	err := es.Save(events)
	if err == nil {
		t.Error("should not be able to save events that belongs to other aggregate type")
	}
}

func saveEventsInWrongOrder(t *testing.T, es eventsourcing.EventStore) {
	events := append(testEvents(), testEvents()[0])
	err := es.Save(events)
	if err == nil {
		t.Error("should not be able to save events that are in wrong version order")
	}
}

func saveEventsInWrongVersion(t *testing.T, es eventsourcing.EventStore) {
	events := testEventsPartTwo()
	err := es.Save(events)
	if err == nil {
		t.Error("should not be able to save events that are out of sync compared to the storage order")
	}
}

func saveEventsWithEmptyReason(t *testing.T, es eventsourcing.EventStore) {
	events := testEvents()
	events[2].Reason = ""
	err := es.Save(events)
	if err == nil {
		t.Error("should not be able to save events with empty reason")
	}
}

func saveAndGetEventsConcurrently(t *testing.T, es eventsourcing.EventStore) {
	wg := sync.WaitGroup{}

	wg.Add(10)
	for i := 0; i < 10; i++ {
		events := testEventsWithID(fmt.Sprintf("id-%d", i))
		go func() {
			err := es.Save(events)
			if err != nil {
				t.Fatal(err)
			}
			wg.Done()
		}()
	}
	wg.Wait()
	wg.Add(10)
	for i := 0; i < 10; i++ {
		eventID := fmt.Sprintf("id-%d", i)
		go func() {
			events, err := es.Get(eventID, aggregateType, 0)
			if err != nil {
				t.Fatal(err)
			}
			if len(events) != 6 {
				t.Fatalf("wrong number of events fetched, expecting 6 got %d", len(events))
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func getErrWhenNoEvents(t *testing.T, es eventsourcing.EventStore) {
	_, err := es.Get(aggregateID, aggregateType, 0)
	if !errors.Is(err, eventsourcing.ErrNoEvents) {
		t.Fatal("expect error when no events are saved for aggregate")
	}
}
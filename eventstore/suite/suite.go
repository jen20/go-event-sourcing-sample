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
		run   func(t *testing.T, es eventsourcing.EventStore) error
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
		{"should get global event order from save", saveReturnGlobalEventOrder},
		{"should set global event on event when saved", setGlobalVersionOnSavedEvents},
	}
	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			es, closeFunc, err := esFunc()
			if err != nil {
				t.Fatal(err)
			}
			err = test.run(t, es)
			if err != nil {
				// make use of t.Error instead of t.Fatal to make sure the closeFunc is executed
				t.Error(err)
			}
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

func saveAndGetEvents(t *testing.T, es eventsourcing.EventStore) error {
	_, err := es.Save(testEvents())
	if err != nil {
		return err
	}

	fetchedEvents, err := es.Get(aggregateID, aggregateType, 0)
	if err != nil {
		return err
	}

	if len(fetchedEvents) != len(testEvents()) {
		return errors.New("wrong number of events returned")

	}

	if fetchedEvents[0].Version != testEvents()[0].Version {
		return errors.New("wrong events returned")
	}

	// Add more events to the same aggregate event stream
	_, err = es.Save(testEventsPartTwo())
	if err != nil {
		return err
	}

	fetchedEventsIncludingPartTwo, err := es.Get(aggregateID, aggregateType, 0)
	if err != nil {
		return err
	}

	if len(fetchedEventsIncludingPartTwo) != len(append(testEvents(), testEventsPartTwo()...)) {
		return errors.New("wrong number of events returned")
	}

	if fetchedEventsIncludingPartTwo[0].Version != testEvents()[0].Version {
		return errors.New("wrong event version returned")
	}

	if fetchedEventsIncludingPartTwo[0].AggregateID != testEvents()[0].AggregateID {
		return errors.New("wrong event aggregateID returned")
	}

	if fetchedEventsIncludingPartTwo[0].AggregateType != testEvents()[0].AggregateType {
		return errors.New("wrong event aggregateType returned")
	}

	if fetchedEventsIncludingPartTwo[0].Reason != testEvents()[0].Reason {
		return errors.New("wrong event aggregateType returned")
	}

	if fetchedEventsIncludingPartTwo[0].MetaData["test"] != "hello" {
		return errors.New("wrong event meta data returned")
	}

	if fetchedEventsIncludingPartTwo[0].Timestamp.Format(time.RFC3339) != timestamp.Format(time.RFC3339) {
		return fmt.Errorf("wrong timestamp exp: %s got: %s", fetchedEventsIncludingPartTwo[0].Timestamp.Format(time.RFC3339), timestamp.Format(time.RFC3339))
	}

	data, ok := fetchedEventsIncludingPartTwo[0].Data.(*FrequentFlierAccountCreated)
	if !ok {
		return errors.New("wrong type in Data")
	}

	if data.OpeningMiles != 10000 {
		return fmt.Errorf("wrong OpeningMiles %d", data.OpeningMiles)
	}
	return nil
}

func getEventsAfterVersion(t *testing.T, es eventsourcing.EventStore) error {
	_, err := es.Save(testEvents())
	if err != nil {
		return err
	}

	fetchedEvents, err := es.Get(aggregateID, aggregateType, 1)
	if err != nil {
		return err
	}

	// Should return one less event
	if len(fetchedEvents) != len(testEvents())-1 {
		return fmt.Errorf("wrong number of events returned exp: %d, got:%d", len(fetchedEvents), len(testEvents())-1)
	}
	// first event version should be 2
	if fetchedEvents[0].Version != 2 {
		return fmt.Errorf("wrong events returned")
	}
	return nil
}

func saveEventsFromMoreThanOneAggregate(t *testing.T, es eventsourcing.EventStore) error {
	invalidEvent := append(testEvents(), testEventOtherAggregate())
	_, err := es.Save(invalidEvent)
	if err == nil {
		return errors.New("should not be able to save events that belongs to more than one aggregate")
	}
	return nil
}

func saveEventsFromMoreThanOneAggregateType(t *testing.T, es eventsourcing.EventStore) error {
	events := testEvents()
	events[1].AggregateType = "OtherAggregateType"

	_, err := es.Save(events)
	if err == nil {
		return errors.New("should not be able to save events that belongs to other aggregate type")
	}
	return nil
}

func saveEventsInWrongOrder(t *testing.T, es eventsourcing.EventStore) error {
	events := append(testEvents(), testEvents()[0])
	_, err := es.Save(events)
	if err == nil {
		return errors.New("should not be able to save events that are in wrong version order")
	}
	return nil
}

func saveEventsInWrongVersion(t *testing.T, es eventsourcing.EventStore) error {
	events := testEventsPartTwo()
	_, err := es.Save(events)
	if err == nil {
		return errors.New("should not be able to save events that are out of sync compared to the storage order")
	}
	return nil
}

func saveEventsWithEmptyReason(t *testing.T, es eventsourcing.EventStore) error {
	events := testEvents()
	events[2].Reason = ""
	_, err := es.Save(events)
	if err == nil {
		return errors.New("should not be able to save events with empty reason")
	}
	return nil
}

func saveAndGetEventsConcurrently(t *testing.T, es eventsourcing.EventStore) error {
	wg := sync.WaitGroup{}
	var err error

	wg.Add(10)
	for i := 0; i < 10; i++ {
		events := testEventsWithID(fmt.Sprintf("id-%d", i))
		go func() {
			_, e := es.Save(events)
			if e != nil {
				err = e
			}
			wg.Done()
		}()
	}
	if err != nil {
		return err
	}

	wg.Wait()
	wg.Add(10)
	for i := 0; i < 10; i++ {
		eventID := fmt.Sprintf("id-%d", i)
		go func() {
			defer wg.Done()
			events, e := es.Get(eventID, aggregateType, 0)
			if e != nil {
				err = e
				return
			}
			if len(events) != 6 {
				err = fmt.Errorf("wrong number of events fetched, expecting 6 got %d", len(events))
				return
			}
		}()
	}
	if err != nil {
		return err
	}
	wg.Wait()
	return nil
}

func getErrWhenNoEvents(t *testing.T, es eventsourcing.EventStore) error {
	_, err := es.Get(aggregateID, aggregateType, 0)
	if !errors.Is(err, eventsourcing.ErrNoEvents) {
		return fmt.Errorf("expect error when no events are saved for aggregate")
	}
	return nil
}
func saveReturnGlobalEventOrder(t *testing.T, es eventsourcing.EventStore) error {
	events := testEvents()
	g, err := es.Save(events)
	if err != nil {
		return err
	}
	if g != 6 {
		return fmt.Errorf("expected global event order 6 got %d", g)
	}
	events2 := testEventOtherAggregate()
	g, err = es.Save([]eventsourcing.Event{events2})
	if err != nil {
		return err
	}
	if g != 7 {
		return fmt.Errorf("expected global event order 7 got %d", g)
	}
	return nil
}

func setGlobalVersionOnSavedEvents(t *testing.T, es eventsourcing.EventStore) error {
	events := testEvents()
	_, err := es.Save(events)
	if err != nil {
		return err
	}
	eventsGet, err := es.Get(events[0].AggregateID, events[0].AggregateType, 0)
	if err != nil {
		return err
	}
	var g eventsourcing.GlobalVersion
	for _, e := range eventsGet {
		g++
		if e.GlobalVersion != g {
			return fmt.Errorf("expected global version to be in sequens exp: %d, was: %d", g, e.GlobalVersion)
		}
	}
	return nil
}

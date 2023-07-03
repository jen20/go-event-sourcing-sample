package suite

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/hallgren/eventsourcing/base"
)

var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

func AggregateID() string {
	r := seededRand.Intn(999999999999)
	return fmt.Sprintf("%d", r)
}

type eventstoreFunc = func(ser base.Serializer) (base.EventStore, func(), error)

func Test(t *testing.T, esFunc eventstoreFunc) {
	tests := []struct {
		title string
		run   func(es base.EventStore) error
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
	}
	ser := base.NewSerializer(json.Marshal, json.Unmarshal)

	ser.Register(&FrequentFlierAccount{},
		ser.Events(
			&FrequentFlierAccountCreated{},
			&FlightTaken{},
			&StatusMatched{},
		),
	)

	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			es, closeFunc, err := esFunc(*ser)
			if err != nil {
				t.Fatal(err)
			}
			err = test.run(es)
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
	//eventsourcing.AggregateRoot
}

func (f *FrequentFlierAccount) Transition(e base.Event) {}

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

var aggregateType = "FrequentFlierAccount"
var timestamp = time.Now()

func testEventsWithID(aggregateID string) []base.Event {
	metadata := make(map[string]interface{})
	metadata["test"] = "hello"
	history := []base.Event{
		{AggregateID: aggregateID, Version: 1, AggregateType: aggregateType, Timestamp: timestamp, Data: &FrequentFlierAccountCreated{AccountId: "1234567", OpeningMiles: 10000, OpeningTierPoints: 0}, Metadata: metadata},
		{AggregateID: aggregateID, Version: 2, AggregateType: aggregateType, Timestamp: timestamp, Data: &StatusMatched{NewStatus: StatusSilver}, Metadata: metadata},
		{AggregateID: aggregateID, Version: 3, AggregateType: aggregateType, Timestamp: timestamp, Data: &FlightTaken{MilesAdded: 2525, TierPointsAdded: 5}, Metadata: metadata},
		{AggregateID: aggregateID, Version: 4, AggregateType: aggregateType, Timestamp: timestamp, Data: &FlightTaken{MilesAdded: 2512, TierPointsAdded: 5}, Metadata: metadata},
		{AggregateID: aggregateID, Version: 5, AggregateType: aggregateType, Timestamp: timestamp, Data: &FlightTaken{MilesAdded: 5600, TierPointsAdded: 5}, Metadata: metadata},
		{AggregateID: aggregateID, Version: 6, AggregateType: aggregateType, Timestamp: timestamp, Data: &FlightTaken{MilesAdded: 3000, TierPointsAdded: 3}, Metadata: metadata},
	}
	return history
}

func testEvents(aggregateID string) []base.Event {
	return testEventsWithID(aggregateID)
}

func testEventsPartTwo(aggregateID string) []base.Event {
	history := []base.Event{
		{AggregateID: aggregateID, Version: 7, AggregateType: aggregateType, Timestamp: timestamp, Data: &FlightTaken{MilesAdded: 5600, TierPointsAdded: 5}},
		{AggregateID: aggregateID, Version: 8, AggregateType: aggregateType, Timestamp: timestamp, Data: &FlightTaken{MilesAdded: 3000, TierPointsAdded: 3}},
	}
	return history
}

func testEventOtherAggregate(aggregateID string) base.Event {
	return base.Event{AggregateID: aggregateID, Version: 1, AggregateType: aggregateType, Timestamp: timestamp, Data: &FrequentFlierAccountCreated{AccountId: "1234567", OpeningMiles: 10000, OpeningTierPoints: 0}}
}

func saveAndGetEvents(es base.EventStore) error {
	aggregateID := AggregateID()
	events := testEvents(aggregateID)
	fetchedEvents := []base.Event{}
	err := es.Save(events)
	if err != nil {
		return err
	}
	iterator, err := es.Get(context.Background(), aggregateID, aggregateType, 0)
	if err != nil {
		return err
	}
	for {
		event, err := iterator.Next()
		if errors.Is(err, base.ErrNoMoreEvents) {
			break
		}
		if err != nil {
			return err
		}
		fetchedEvents = append(fetchedEvents, event)
	}
	iterator.Close()
	if len(fetchedEvents) != len(testEvents(aggregateID)) {
		return errors.New("wrong number of events returned")
	}

	if fetchedEvents[0].Version != testEvents(aggregateID)[0].Version {
		return errors.New("wrong events returned")
	}

	// Add more events to the same aggregate event stream
	err = es.Save(testEventsPartTwo(aggregateID))
	if err != nil {
		return err
	}
	fetchedEventsIncludingPartTwo := []base.Event{}
	iterator, err = es.Get(context.Background(), aggregateID, aggregateType, 0)
	if err != nil {
		return err
	}
	for {
		event, err := iterator.Next()
		if err != nil {
			break
		}
		fetchedEventsIncludingPartTwo = append(fetchedEventsIncludingPartTwo, event)
	}
	iterator.Close()

	if len(fetchedEventsIncludingPartTwo) != len(append(testEvents(aggregateID), testEventsPartTwo(aggregateID)...)) {
		return errors.New("wrong number of events returned")
	}

	if fetchedEventsIncludingPartTwo[0].Version != testEvents(aggregateID)[0].Version {
		return errors.New("wrong event version returned")
	}

	if fetchedEventsIncludingPartTwo[0].AggregateID != testEvents(aggregateID)[0].AggregateID {
		return errors.New("wrong event aggregateID returned")
	}

	if fetchedEventsIncludingPartTwo[0].AggregateType != testEvents(aggregateID)[0].AggregateType {
		return errors.New("wrong event aggregateType returned")
	}

	if fetchedEventsIncludingPartTwo[0].Reason() != testEvents(aggregateID)[0].Reason() {
		return errors.New("wrong event aggregateType returned")
	}

	if fetchedEventsIncludingPartTwo[0].Metadata["test"] != "hello" {
		return errors.New("wrong event meta data returned")
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

func getEventsAfterVersion(es base.EventStore) error {
	var fetchedEvents []base.Event
	aggregateID := AggregateID()
	err := es.Save(testEvents(aggregateID))
	if err != nil {
		return err
	}

	iterator, err := es.Get(context.Background(), aggregateID, aggregateType, 1)
	if err != nil {
		return err
	}

	for {
		event, err := iterator.Next()
		if err != nil {
			break
		}
		fetchedEvents = append(fetchedEvents, event)
	}
	iterator.Close()
	// Should return one less event
	if len(fetchedEvents) != len(testEvents(aggregateID))-1 {
		return fmt.Errorf("wrong number of events returned exp: %d, got:%d", len(fetchedEvents), len(testEvents(aggregateID))-1)
	}
	// first event version should be 2
	if fetchedEvents[0].Version != 2 {
		return fmt.Errorf("wrong events returned")
	}
	return nil
}

func saveEventsFromMoreThanOneAggregate(es base.EventStore) error {
	aggregateID := AggregateID()
	aggregateIDOther := AggregateID()
	invalidEvent := append(testEvents(aggregateID), testEventOtherAggregate(aggregateIDOther))
	err := es.Save(invalidEvent)
	if err == nil {
		return errors.New("should not be able to save events that belongs to more than one aggregate")
	}
	return nil
}

func saveEventsFromMoreThanOneAggregateType(es base.EventStore) error {
	aggregateID := AggregateID()
	events := testEvents(aggregateID)
	events[1].AggregateType = "OtherAggregateType"

	err := es.Save(events)
	if err == nil {
		return errors.New("should not be able to save events that belongs to other aggregate type")
	}
	return nil
}

func saveEventsInWrongOrder(es base.EventStore) error {
	aggregateID := AggregateID()
	events := append(testEvents(aggregateID), testEvents(aggregateID)[0])
	err := es.Save(events)
	if err == nil {
		return errors.New("should not be able to save events that are in wrong version order")
	}
	return nil
}

func saveEventsInWrongVersion(es base.EventStore) error {
	aggregateID := AggregateID()
	events := testEventsPartTwo(aggregateID)
	err := es.Save(events)
	if err == nil {
		return errors.New("should not be able to save events that are out of sync compared to the storage order")
	}
	return nil
}

func saveEventsWithEmptyReason(es base.EventStore) error {
	aggregateID := AggregateID()
	events := testEvents(aggregateID)
	events[2].Data = nil
	err := es.Save(events)
	if err == nil {
		return errors.New("should not be able to save events with empty reason")
	}
	return nil
}

func saveAndGetEventsConcurrently(es base.EventStore) error {
	wg := sync.WaitGroup{}
	var err error
	aggregateID := AggregateID()

	wg.Add(10)
	for i := 0; i < 10; i++ {
		events := testEventsWithID(fmt.Sprintf("%s-%d", aggregateID, i))
		go func() {
			e := es.Save(events)
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
		eventID := fmt.Sprintf("%s-%d", aggregateID, i)
		go func() {
			defer wg.Done()
			iterator, e := es.Get(context.Background(), eventID, aggregateType, 0)
			if e != nil {
				err = e
				return
			}
			events := make([]base.Event, 0)
			for {
				event, err := iterator.Next()
				if err != nil {
					break
				}
				events = append(events, event)
			}
			iterator.Close()
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

func getErrWhenNoEvents(es base.EventStore) error {
	aggregateID := AggregateID()
	iterator, err := es.Get(context.Background(), aggregateID, aggregateType, 0)
	if err != nil {
		if err != base.ErrNoEvents {
			return err
		}
		return nil
	}
	defer iterator.Close()
	_, err = iterator.Next()
	if !errors.Is(err, base.ErrNoMoreEvents) {
		return fmt.Errorf("expect error when no events are saved for aggregate")
	}
	return nil
}

func saveReturnGlobalEventOrder(es base.EventStore) error {
	aggregateID := AggregateID()
	aggregateID2 := AggregateID()
	events := testEvents(aggregateID)
	err := es.Save(events)
	if err != nil {
		return err
	}
	if events[len(events)-1].GlobalVersion == 0 {
		return fmt.Errorf("expected global event order > 0 on last event got %d", events[len(events)-1].GlobalVersion)
	}
	events2 := []base.Event{testEventOtherAggregate(aggregateID2)}
	err = es.Save(events2)
	if err != nil {
		return err
	}
	if events2[0].GlobalVersion <= events[len(events)-1].GlobalVersion {
		return fmt.Errorf("expected larger global event order got %d", events2[0].GlobalVersion)
	}
	return nil
}

/* re-activate when esdb eventstore have global event order on each stream
func setGlobalVersionOnSavedEvents(es eventsourcing.EventStore) error {
	events := testEvents()
	err := es.Save(events)
	if err != nil {
		return err
	}
	eventsGet, err := es.Get(events[0].AggregateID, events[0].AggregateType, 0)
	if err != nil {
		return err
	}
	var g eventsourcing.Version
	for _, e := range eventsGet {
		g++
		if e.GlobalVersion != g {
			return fmt.Errorf("expected global version to be in sequens exp: %d, was: %d", g, e.GlobalVersion)
		}
	}
	return nil
}
*/

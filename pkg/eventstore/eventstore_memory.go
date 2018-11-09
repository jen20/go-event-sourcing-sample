package eventstore

import (
	"fmt"
	"go-event-sourcing-sample/pkg/eventsourcing"
)

// Memory is a handler for event streaming
type Memory struct {
	aggregateEvents map[string][]eventsourcing.Event // The memory structure where we store aggregate events
	eventsInOrder   []eventsourcing.Event
}

// CreateMemory in memory event store
func CreateMemory() *Memory {
	return &Memory{
		aggregateEvents: make(map[string][]eventsourcing.Event),
		eventsInOrder:   make([]eventsourcing.Event, 0),
	}
}

// Save an aggregate (its events)
func (e *Memory) Save(events []eventsourcing.Event) error {
	// Return if there is no events to save
	if len(events) == 0 {
		return nil
	}

	// get bucket name from first event
	aggregateType := events[0].AggregateType
	aggregateID := events[0].AggregateRootID
	bucketName := bucketName(aggregateType, string(aggregateID))

	evBucket := e.aggregateEvents[bucketName]
	currentVersion := eventsourcing.Version(0)

	if len(evBucket) > 0 {
		// Last version in the list
		currentVersion = evBucket[len(evBucket)-1].Version
	}

	//Validate events
	ok, err := e.validateEvents(aggregateID, currentVersion, events)
	if !ok {
		//TODO created describing errors
		return err
	}

	eventsInOrder := e.eventsInOrder

	for _, event := range events {
		evBucket = append(evBucket, event)
		eventsInOrder = append(eventsInOrder, event)
	}

	e.aggregateEvents[bucketName] = evBucket
	e.eventsInOrder = eventsInOrder

	return nil
}

// Get aggregate events
func (e *Memory) Get(id string, aggregateType string) []eventsourcing.Event {
	return e.aggregateEvents[bucketName(aggregateType, id)]
}

// GlobalGet returns events from the global order
func (e *Memory) GlobalGet(start int, count int) []eventsourcing.Event {

	events := make([]eventsourcing.Event, 0)

	for i, event := range e.eventsInOrder {
		if i >= start+count {
			break
		}
		if i >= start {
			events = append(events, event)
		}

	}

	return events
}

// Close does nothing
func (e *Memory) Close() {

}

func bucketName(aggregateType, aggregateID string) string {
	return aggregateType + "_" + aggregateID
}

func (e *Memory) validateEvents(aggregateID eventsourcing.AggregateRootID, currentVersion eventsourcing.Version, events []eventsourcing.Event) (bool, error) {
	aggregateType := events[0].AggregateType

	for _, event := range events {
		if event.AggregateRootID != aggregateID {
			return false, fmt.Errorf("events holds events for more than one aggregate")
		}

		if event.AggregateType != aggregateType {
			return false, fmt.Errorf("events holds events for more than one aggregate type")
		}

		if currentVersion+1 != event.Version {
			return false, fmt.Errorf("concurrency error")
		}

		if event.Reason == "" {
			return false, fmt.Errorf("event holds no reason")
		}

		currentVersion = event.Version
	}
	return true, nil
}

package memory

import (
	"eventsourcing"
	"eventsourcing/eventstore"
)

// Memory is a handler for event streaming
type Memory struct {
	aggregateEvents map[string][]eventsourcing.Event // The memory structure where we store aggregate events
	eventsInOrder   []eventsourcing.Event            // The global event order
}

// Create in memory event store
func Create() *Memory {
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
	bucketName := aggregateKey(aggregateType, string(aggregateID))

	evBucket := e.aggregateEvents[bucketName]
	currentVersion := eventsourcing.Version(0)

	if len(evBucket) > 0 {
		// Last version in the list
		currentVersion = evBucket[len(evBucket)-1].Version
	}

	//Validate events
	err := eventstore.ValidateEvents(aggregateID, currentVersion, events)
	if err != nil {
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
func (e *Memory) Get(id string, aggregateType string) ([]eventsourcing.Event, error) {
	return e.aggregateEvents[aggregateKey(aggregateType, id)], nil
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

// aggregateKey generate a aggregate key to store events against from aggregateType and aggregateID
func aggregateKey(aggregateType, aggregateID string) string {
	return aggregateType + "_" + aggregateID
}

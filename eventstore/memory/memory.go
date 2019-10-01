package memory

import (
	"github.com/hallgren/eventsourcing"
	"github.com/hallgren/eventsourcing/eventstore"
)

// Memory is a handler for event streaming
type Memory struct {
	aggregateEvents map[string][][]byte // The memory structure where we store aggregate events
	eventsInOrder   [][]byte            // The global event order
	serializer      eventstore.EventSerializer
}

// Create in memory event store
func Create(serializer eventstore.EventSerializer) *Memory {
	return &Memory{
		aggregateEvents: make(map[string][][]byte),
		eventsInOrder:   make([][]byte, 0),
		serializer:      serializer,
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
		lastEvent, err := e.serializer.DeserializeEvent(evBucket[len(evBucket)-1])
		if err != nil {
			return err
		}
		currentVersion = lastEvent.Version
	}

	//Validate events
	err := eventstore.ValidateEvents(aggregateID, currentVersion, events)
	if err != nil {
		return err
	}

	eventsInOrder := e.eventsInOrder

	for _, event := range events {
		eventSerialized, err := e.serializer.SerializeEvent(event)
		if err != nil {
			return err
		}
		evBucket = append(evBucket, eventSerialized)
		eventsInOrder = append(eventsInOrder, eventSerialized)
	}

	e.aggregateEvents[bucketName] = evBucket
	e.eventsInOrder = eventsInOrder

	return nil
}

// Get aggregate events
func (e *Memory) Get(id string, aggregateType string, afterVersion eventsourcing.Version) ([]eventsourcing.Event, error) {
	var events []eventsourcing.Event
	eventsSerialized := e.aggregateEvents[aggregateKey(aggregateType, id)]
	for _, eventSerialized := range eventsSerialized {
		event, err := e.serializer.DeserializeEvent(eventSerialized)
		if err != nil {
			return nil, err
		}
		if event.Version > afterVersion {
			events = append(events, event)
		}
	}
	return events, nil
}

// GlobalGet returns events from the global order
func (e *Memory) GlobalGet(start int, count int) []eventsourcing.Event {
	events := make([]eventsourcing.Event, 0)
	var i int
	for id, eventSerialized := range e.eventsInOrder {
		event, err := e.serializer.DeserializeEvent(eventSerialized)
		if err != nil {
			return nil
		}
		if id >= start-1 {
			events = append(events, event)
			i++
			if i == count {
				break
			}
		}
	}
	return events
}

// Close does nothing
func (e *Memory) Close() {}

// aggregateKey generate a aggregate key to store events against from aggregateType and aggregateID
func aggregateKey(aggregateType, aggregateID string) string {
	return aggregateType + "_" + aggregateID
}

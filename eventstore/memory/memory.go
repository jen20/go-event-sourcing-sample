package memory

import (
	"sync"

	"github.com/hallgren/eventsourcing"
	"github.com/hallgren/eventsourcing/eventstore"
)

// Memory is a handler for event streaming
type Memory struct {
	aggregateEvents map[string][]eventsourcing.Event // The memory structure where we store aggregate events
	eventsInOrder   []eventsourcing.Event            // The global event order
	lock            sync.Mutex
}

// Create in memory event store
func Create() *Memory {
	return &Memory{
		aggregateEvents: make(map[string][]eventsourcing.Event),
		eventsInOrder:   make([]eventsourcing.Event, 0),
	}
}

// Save an aggregate (its events)
func (e *Memory) Save(events []eventsourcing.Event) (uint64, error) {
	// Return if there is no events to save
	if len(events) == 0 {
		return 0, nil
	}

	// make sure its thread safe
	e.lock.Lock()
	defer e.lock.Unlock()

	// get bucket name from first event
	aggregateType := events[0].AggregateType
	aggregateID := events[0].AggregateID
	bucketName := aggregateKey(aggregateType, aggregateID)

	evBucket := e.aggregateEvents[bucketName]
	currentVersion := eventsourcing.Version(0)

	if len(evBucket) > 0 {
		// Last version in the list
		lastEvent := evBucket[len(evBucket)-1]
		currentVersion = lastEvent.Version
	}

	//Validate events
	err := eventstore.ValidateEvents(aggregateID, currentVersion, events)
	if err != nil {
		return 0, err
	}

	eventsInOrder := e.eventsInOrder

	for _, event := range events {
		if err != nil {
			return 0, err
		}
		evBucket = append(evBucket, event)
		eventsInOrder = append(eventsInOrder, event)
	}

	e.aggregateEvents[bucketName] = evBucket
	e.eventsInOrder = eventsInOrder

	return uint64(len(e.eventsInOrder)), nil
}

// Get aggregate events
func (e *Memory) Get(id string, aggregateType string, afterVersion eventsourcing.Version) ([]eventsourcing.Event, error) {
	var events []eventsourcing.Event
	// make sure its thread safe
	e.lock.Lock()
	defer e.lock.Unlock()

	for _, e := range e.aggregateEvents[aggregateKey(aggregateType, id)] {
		if e.Version > afterVersion {
			events = append(events, e)
		}
	}
	if len(events) == 0 {
		return nil, eventsourcing.ErrNoEvents
	}
	return events, nil
}

// Close does nothing
func (e *Memory) Close() {}

// aggregateKey generate a aggregate key to store events against from aggregateType and aggregateID
func aggregateKey(aggregateType, aggregateID string) string {
	return aggregateType + "_" + aggregateID
}

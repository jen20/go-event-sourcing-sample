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

type iterator struct {
	events   []eventsourcing.Event
	position int
}

func (i *iterator) Next() (eventsourcing.Event, error) {
	if len(i.events) <= i.position {
		return eventsourcing.Event{}, eventsourcing.ErrNoMoreEvents
	}
	event := i.events[i.position]
	i.position++
	return event, nil
}

func (i *iterator) Close() {
	i.events = nil
	i.position = 0
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
		return err
	}

	for i, event := range events {
		// set the global version on the event +1 as if the event was already on the eventsInOrder slice
		event.GlobalVersion = eventsourcing.Version(len(e.eventsInOrder) + 1)
		evBucket = append(evBucket, event)
		e.eventsInOrder = append(e.eventsInOrder, event)
		// override the event in the slice exposing the GlobalVersion to the caller
		events[i].GlobalVersion = event.GlobalVersion
	}

	e.aggregateEvents[bucketName] = evBucket
	return nil
}

// Get aggregate events
func (e *Memory) Get(id string, aggregateType string, afterVersion eventsourcing.Version) (eventsourcing.EventIterator, error) {
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
	return &iterator{events: events}, nil
}

// GlobalEvents will return count events in order globaly from the start posistion
func (e *Memory) GlobalEvents(start, count uint64) ([]eventsourcing.Event, error) {
	var events []eventsourcing.Event
	// make sure its thread safe
	e.lock.Lock()
	defer e.lock.Unlock()

	for _, e := range e.eventsInOrder {
		// find start position and append until counter is 0
		if uint64(e.GlobalVersion) >= start {
			events = append(events, e)
			count--
			if count == 0 {
				break
			}
		}
	}
	return events, nil
}

// Close does nothing
func (e *Memory) Close() {}

// aggregateKey generate a aggregate key to store events against from aggregateType and aggregateID
func aggregateKey(aggregateType, aggregateID string) string {
	return aggregateType + "_" + aggregateID
}

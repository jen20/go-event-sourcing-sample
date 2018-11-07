package eventsourcing

import (
	"reflect"

	uuid "github.com/satori/go.uuid"
)

// AggregateRoot to be included into aggregates
type AggregateRoot struct {
	ID      string
	Version int
	Changes []Event
}

// Event holding meta data and
type Event struct {
	AggregateRootID string
	Version         int
	Reason          string
	AggregateType   string
	Data            interface{}
	MetaData        map[string]interface{}
}

// transition function to apply events on aggregates
type transition func(Event)

// TrackChange is used internally by bevhavious methods to apply a state change to
// the current instance and also track it in order that it can be persisted later.
func (state *AggregateRoot) TrackChange(aggregate interface{}, eventData interface{}, fn transition) {

	// This can be overwritten in the constructor of the aggregate
	if state.ID == "" {
		state.ID = uuid.Must(uuid.NewV4()).String()
	}

	eventVersion := state.CurrentVersion() + 1
	reason := reflect.TypeOf(eventData).Name()
	aggregateType := reflect.TypeOf(aggregate).Name()

	event := Event{AggregateRootID: state.ID, Version: eventVersion, Reason: reason, AggregateType: aggregateType, Data: eventData}
	state.Changes = append(state.Changes, event)

	fn(event)
}

// BuildFromHistory builds the aggregate state from events
func (state *AggregateRoot) BuildFromHistory(events []Event, fn transition) {
	for _, event := range events {
		fn(event)
		//Set the aggregate id
		state.ID = event.AggregateRootID
		// Make sure the aggregate is in the correct version (the last event)
		state.Version = event.Version
	}
}

// CurrentVersion get the current version including the pending changes
func (state *AggregateRoot) CurrentVersion() int {
	if len(state.Changes) > 0 {
		lastEvent := state.Changes[len(state.Changes)-1]
		return lastEvent.Version
	}
	return state.Version
}

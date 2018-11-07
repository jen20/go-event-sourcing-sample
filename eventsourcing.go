package main

import (
	"reflect"

	uuid "github.com/satori/go.uuid"
)

type AggregateRoot struct {
	id      string
	version int
	changes []Event
}

// Event holding meta data and
type Event struct {
	aggregateRootID string
	version         int
	reason          string
	aggregateType   string
	data            interface{}
	metaData        map[string]interface{}
}

// transition function to apply events on aggregates
type transition func(Event)

// trackChange is used internally by bevhavious methods to apply a state change to
// the current instance and also track it in order that it can be persisted later.
func (state *AggregateRoot) trackChange(aggregate interface{}, eventData interface{}, fn transition) {

	// This can be overwritten in the constructor of the aggregate
	if state.id == "" {
		state.id = uuid.Must(uuid.NewV4()).String()
	}

	eventVersion := state.currentVersion() + 1
	reason := reflect.TypeOf(eventData).Name()
	aggregateType := reflect.TypeOf(aggregate).Name()

	event := Event{aggregateRootID: state.id, version: eventVersion, reason: reason, aggregateType: aggregateType, data: eventData}
	state.changes = append(state.changes, event)

	fn(event)
}

func (state *AggregateRoot) buildFromHistory(events []Event, fn transition) {
	for _, event := range events {
		fn(event)
		//Set the aggregate id
		state.id = event.aggregateRootID
		// Make sure the aggregate is in the correct version (the last event)
		state.version = event.version
	}
}

func (state *AggregateRoot) currentVersion() int {
	if len(state.changes) > 0 {
		lastEvent := state.changes[len(state.changes)-1]
		return lastEvent.version
	}
	return state.version
}

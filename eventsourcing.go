package eventsourcing

import (
	"errors"
	"reflect"

	uuid "github.com/satori/go.uuid"
)

// Version is the event version used in event and aggregateRoot
type Version int

// AggregateRootID is the identifier on the aggregate
type AggregateRootID string

// AggregateRoot to be included into aggregates
type AggregateRoot struct {
	id      AggregateRootID
	version Version
	changes []Event
}

// Event holding meta data and the application specific event in the Data property
type Event struct {
	AggregateRootID AggregateRootID
	Version         Version
	Reason          string
	AggregateType   string
	Data            interface{}
	MetaData        map[string]interface{}
}

// Transition function to apply events on aggregates to build its current state
type transition func(Event)

// ErrAggregateAlreadyExists returned if the ID is set more than one time
var ErrAggregateAlreadyExists = errors.New("its not possible to set id on already existing aggregate")

var emptyAggregateID = AggregateRootID("")

// TrackChange is used internally by behaviour methods to apply a state change to
// the current instance and also track it in order that it can be persisted later.
func (state *AggregateRoot) TrackChange(aggregate interface{}, eventData interface{}, fn transition) {
	// This can be overwritten in the constructor of the aggregate
	if state.id == emptyAggregateID {
		state.setID(uuid.Must(uuid.NewV4()).String())
	}

	reason := reflect.TypeOf(eventData).Name()
	aggregateType := reflect.TypeOf(aggregate).Elem().Name()
	event := Event{
		AggregateRootID: state.id,
		Version:         state.nextVersion(),
		Reason:          reason,
		AggregateType:   aggregateType,
		Data:            eventData,
	}
	state.changes = append(state.changes, event)
	fn(event)
}

// BuildFromHistory builds the aggregate state from events
func (state *AggregateRoot) BuildFromHistory(events []Event, fn transition) {
	for _, event := range events {
		fn(event)
		//Set the aggregate id
		state.id = event.AggregateRootID
		// Make sure the aggregate is in the correct version (the last event)
		state.version = event.Version
	}
}

func (state *AggregateRoot) nextVersion() Version {
	return state.currentVersion() + 1
}

func (state *AggregateRoot) currentVersion() Version {
	if len(state.changes) > 0 {
		return state.changes[len(state.changes)-1].Version
	}
	return state.version
}

// setID is the internal method to set the aggregate id
func (state *AggregateRoot) setID(id string) {
	state.id = AggregateRootID(id)
}

//Public accessors for aggregate root properties

// Setters

// SetID opens up the possibility to set manual aggregate id from the outside
func (state *AggregateRoot) SetID(id string) error {
	//TODO: Validate id structure

	if state.id != emptyAggregateID {
		return ErrAggregateAlreadyExists
	}

	state.setID(id)
	return nil
}

// Getters

// ID exposes the internal id
func (state *AggregateRoot) ID() string {
	return string(state.id)
}

// Changes exposes the internal changes property on the aggregateRoot
func (state *AggregateRoot) Changes() []Event {
	return state.changes
}

// Version get the current version including the pending changes
func (state *AggregateRoot) Version() int {
	return int(state.currentVersion())
}

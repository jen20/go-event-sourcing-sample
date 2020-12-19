package eventsourcing

import (
	"errors"
	"reflect"
	"time"

	uuid "github.com/google/uuid"
)

// Version is the event version used in event and aggregateRoot
type Version int

// AggregateRoot to be included into aggregates
type AggregateRoot struct {
	aggregateID      string
	aggregateVersion Version
	aggregateEvents  []Event
}

// Event holding meta data and the application specific event in the Data property
type Event struct {
	AggregateRootID string
	Version         Version
	Reason          string
	AggregateType   string
	Timestamp       time.Time
	Data            interface{}
	MetaData        map[string]interface{}
}

var (
	// ErrAggregateAlreadyExists returned if the aggregateID is set more than one time
	ErrAggregateAlreadyExists = errors.New("its not possible to set ID on already existing aggregate")
)

const (
	emptyAggregateID = ""
)

// TrackChange is used internally by behaviour methods to apply a state change to
// the current instance and also track it in order that it can be persisted later.
func (state *AggregateRoot) TrackChange(a aggregate, data interface{}) {
	state.TrackChangeWithMetaData(a, data, nil)
}

// TrackChangeWithMetaData is used internally by behaviour methods to apply a state change to
// the current instance and also track it in order that it can be persisted later.
// meta data is handled by this func to store none related application state
func (state *AggregateRoot) TrackChangeWithMetaData(a aggregate, data interface{}, metaData map[string]interface{}) {
	// This can be overwritten in the constructor of the aggregate
	if state.aggregateID == emptyAggregateID {
		state.setID(uuid.New().String())
	}

	reason := reflect.TypeOf(data).Elem().Name()
	name := reflect.TypeOf(a).Elem().Name()
	event := Event{
		AggregateRootID: state.aggregateID,
		Version:         state.nextVersion(),
		Reason:          reason,
		AggregateType:   name,
		Timestamp:       time.Now().UTC(),
		Data:            data,
		MetaData:        metaData,
	}
	state.aggregateEvents = append(state.aggregateEvents, event)
	a.Transition(event)
}

// BuildFromHistory builds the aggregate state from events
func (state *AggregateRoot) BuildFromHistory(a aggregate, events []Event) {
	for _, event := range events {
		a.Transition(event)
		//Set the aggregate ID
		state.aggregateID = event.AggregateRootID
		// Make sure the aggregate is in the correct version (the last event)
		state.aggregateVersion = event.Version
	}
}

func (state *AggregateRoot) nextVersion() Version {
	return state.Version() + 1
}

// updateVersion sets the aggregateVersion to the aggregateVersion in the last event if reset the events
// called by the Save func in the repository after the events are stored
func (state *AggregateRoot) updateVersion() {
	if len(state.aggregateEvents) > 0 {
		state.aggregateVersion = state.aggregateEvents[len(state.aggregateEvents)-1].Version
		state.aggregateEvents = []Event{}
	}
}

func (state *AggregateRoot) changes() []Event {
	return state.aggregateEvents
}

// setID is the internal method to set the aggregate ID
func (state *AggregateRoot) setID(id string) {
	state.aggregateID = id
}

//Public accessors for aggregate root properties

// SetID opens up the possibility to set manual aggregate ID from the outside
func (state *AggregateRoot) SetID(id string) error {
	if state.aggregateID != emptyAggregateID {
		return ErrAggregateAlreadyExists
	}
	state.setID(id)
	return nil
}

// ID returns the aggregate ID as a string
func (state *AggregateRoot) ID() string {
	return state.aggregateID
}

// path return the full name of the aggregate making it unique to other aggregates with
// the same name but placed in other packages.
func (state *AggregateRoot) path() string {
	return reflect.TypeOf(state).Elem().PkgPath()
}

// Version return the version based on events that are not stored
func (state *AggregateRoot) Version() Version {
	if len(state.aggregateEvents) > 0 {
		return state.aggregateEvents[len(state.aggregateEvents)-1].Version
	}
	return state.aggregateVersion
}

// Events return the aggregate events from the aggregate
func (state *AggregateRoot) Events() []Event {
	return state.aggregateEvents
}

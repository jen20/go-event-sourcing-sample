package eventsourcing

import (
	"errors"
	"reflect"
	"time"
)

// Version is the event version used in event and aggregateRoot
type Version int

// AggregateRoot to be included into aggregates
type AggregateRoot struct {
	AggregateID      string
	AggregateVersion Version
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
	// ErrAggregateAlreadyExists returned if the AggregateID is set more than one time
	ErrAggregateAlreadyExists = errors.New("its not possible to set ID on already existing aggregate")
)

const (
	emptyAggregateID = ""
)

// TrackChange is used internally by behaviour methods to apply a state change to
// the current instance and also track it in order that it can be persisted later.
func (state *AggregateRoot) TrackChange(a Aggregate, data interface{}) {
	state.TrackChangeWithMetaData(a, data, nil)
}

// TrackChangeWithMetaData is used internally by behaviour methods to apply a state change to
// the current instance and also track it in order that it can be persisted later.
// meta data is handled by this func to store none related application state
func (state *AggregateRoot) TrackChangeWithMetaData(a Aggregate, data interface{}, metaData map[string]interface{}) {
	// This can be overwritten in the constructor of the aggregate
	if state.AggregateID == emptyAggregateID {
		state.AggregateID = idFunc()
	}

	reason := reflect.TypeOf(data).Elem().Name()
	name := reflect.TypeOf(a).Elem().Name()
	event := Event{
		AggregateRootID: state.AggregateID,
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
func (state *AggregateRoot) BuildFromHistory(a Aggregate, events []Event) {
	for _, event := range events {
		a.Transition(event)
		//Set the aggregate ID
		state.AggregateID = event.AggregateRootID
		// Make sure the aggregate is in the correct version (the last event)
		state.AggregateVersion = event.Version
	}
}

func (state *AggregateRoot) nextVersion() Version {
	return state.Version() + 1
}

// updateVersion sets the AggregateVersion to the AggregateVersion in the last event if reset the events
// called by the Save func in the repository after the events are stored
func (state *AggregateRoot) updateVersion() {
	if len(state.aggregateEvents) > 0 {
		state.AggregateVersion = state.aggregateEvents[len(state.aggregateEvents)-1].Version
		state.aggregateEvents = []Event{}
	}
}

// path return the full name of the aggregate making it unique to other aggregates with
// the same name but placed in other packages.
func (state *AggregateRoot) path() string {
	return reflect.TypeOf(state).Elem().PkgPath()
}

//Public accessors for aggregate Root properties

// SetID opens up the possibility to set manual aggregate ID from the outside
func (state *AggregateRoot) SetID(id string) error {
	if state.AggregateID != emptyAggregateID {
		return ErrAggregateAlreadyExists
	}
	state.AggregateID = id
	return nil
}

// ID returns the aggregate ID as a string
func (state *AggregateRoot) ID() string {
	return state.AggregateID
}

// Root returns the included Aggregate Root state, and is used from the interface Aggregate.
func (state *AggregateRoot) Root() *AggregateRoot {
	return state
}

// Version return the version based on events that are not stored
func (state *AggregateRoot) Version() Version {
	if len(state.aggregateEvents) > 0 {
		return state.aggregateEvents[len(state.aggregateEvents)-1].Version
	}
	return state.AggregateVersion
}

// Events return the aggregate events from the aggregate
func (state *AggregateRoot) Events() []Event {
	return state.aggregateEvents
}

// UnsavedEvents return true if there's unsaved events on the aggregate
func (state *AggregateRoot) UnsavedEvents() bool {
	return len(state.aggregateEvents) > 0
}

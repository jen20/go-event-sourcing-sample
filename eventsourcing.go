package eventsourcing

import (
	"errors"
	"fmt"
	"reflect"

	uuid "github.com/satori/go.uuid"
)

// AggregateVersion is the event version used in event and aggregateRoot
type Version int

// AggregateRootID is the identifier on the aggregate
type AggregateRootID string

// AggregateRoot to be included into aggregates
type AggregateRoot struct {
	AggregateID      AggregateRootID
	AggregateVersion Version
	AggregateEvents  []Event
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

// ErrAggregateAlreadyExists returned if the AggregateID is set more than one time
var ErrAggregateAlreadyExists = errors.New("its not possible to set id on already existing aggregate")

var emptyAggregateID = AggregateRootID("")
var AggregateNotPointerTypeError = fmt.Errorf("aggregate is not a pointer type")
var EventDataNotPointerTypeError = fmt.Errorf("eventData is not a pointer type")

// TrackChange is used internally by behaviour methods to apply a state change to
// the current instance and also track it in order that it can be persisted later.
func (state *AggregateRoot) TrackChange(a aggregate, eventData interface{}) error {
	// Make sure the aggregate and eventData is a pointer type
	if reflect.ValueOf(a).Kind() != reflect.Ptr {
		return AggregateNotPointerTypeError
	}
	if reflect.ValueOf(eventData).Kind() != reflect.Ptr {
		return EventDataNotPointerTypeError
	}

	// This can be overwritten in the constructor of the aggregate
	if state.AggregateID == emptyAggregateID {
		state.setID(uuid.Must(uuid.NewV4()).String())
	}

	reason := reflect.TypeOf(eventData).Elem().Name()
	aggregateType := reflect.TypeOf(a).Elem().Name()
	event := Event{
		AggregateRootID: state.AggregateID,
		Version:         state.nextVersion(),
		Reason:          reason,
		AggregateType:   aggregateType,
		Data:            eventData,
	}
	state.AggregateEvents = append(state.AggregateEvents, event)
	a.Transition(event)
	return nil
}

// BuildFromHistory builds the aggregate state from events
func (state *AggregateRoot) BuildFromHistory(a aggregate, events []Event) {
	for _, event := range events {
		a.Transition(event)
		//Set the aggregate id
		state.AggregateID = event.AggregateRootID
		// Make sure the aggregate is in the correct version (the last event)
		state.AggregateVersion = event.Version
	}
}

func (state *AggregateRoot) nextVersion() Version {
	return state.CurrentVersion() + 1
}

// updateVersion sets the AggregateVersion to the AggregateVersion in the last event if reset the events
// called by the Save func in the repository after the events are stored
func (state *AggregateRoot) updateVersion() {
	if len(state.AggregateEvents) > 0 {
		state.AggregateVersion = state.AggregateEvents[len(state.AggregateEvents)-1].Version
		state.AggregateEvents = []Event{}
	}
}

func (state *AggregateRoot) changes() []Event {
	return state.AggregateEvents
}

// setID is the internal method to set the aggregate id
func (state *AggregateRoot) setID(id string) {
	state.AggregateID = AggregateRootID(id)
}

func (state *AggregateRoot) version() Version {
	return state.AggregateVersion
}

//Public accessors for aggregate root properties

// SetID opens up the possibility to set manual aggregate id from the outside
func (state *AggregateRoot) SetID(id string) error {
	if state.AggregateID != emptyAggregateID {
		return ErrAggregateAlreadyExists
	}
	state.setID(id)
	return nil
}

// ID returns the aggregate id as a string
func (state *AggregateRoot) id() string {
	return state.AggregateID.String()
}

func (id AggregateRootID) String() string {
	return string(id)
}

// CurrentVersion return the version based on events that are not stored
func (state *AggregateRoot) CurrentVersion() Version {
	if len(state.AggregateEvents) > 0 {
		return state.AggregateEvents[len(state.AggregateEvents)-1].Version
	}
	return state.AggregateVersion
}

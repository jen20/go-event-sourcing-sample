package json

import (
	"encoding/json"
	"errors"
	"reflect"
	"time"

	"github.com/hallgren/eventsourcing"
)

// Handler for json serializes
type Handler struct {
	eventRegister map[string]interface{}
}

// New returns a json Handle
func New() *Handler {
	return &Handler{eventRegister: make(map[string]interface{})}
}

type jsonEvent struct {
	AggregateType   string
	Reason          string
	Version         int
	AggregateRootID string
	Timestamp 		time.Time
	Data            json.RawMessage
	MetaData        map[string]interface{}
}

type aggregate interface {
	Transition(event eventsourcing.Event)
}

var (
	// ErrAggregateNameMissing return if aggregate name is missing
	ErrAggregateNameMissing = errors.New("missing aggregate name")

	// ErrNoEventsToRegister return if no events to register
	ErrNoEventsToRegister = errors.New("no events to register")

	// ErrEventNameMissing return if event name is missing
	ErrEventNameMissing = errors.New("missing event name")
)

// Register events aggregate
func (h *Handler) Register(aggregate aggregate, events ...interface{}) error {
	aggregateName := reflect.TypeOf(aggregate).Elem().Name()
	if aggregateName == "" {
		return ErrAggregateNameMissing
	}
	if len(events) == 0 {
		return ErrNoEventsToRegister
	}
	for _, event := range events {
		eventName := reflect.TypeOf(event).Elem().Name()
		if eventName == "" {
			return ErrEventNameMissing
		}
		h.eventRegister[aggregateName+"_"+eventName] = event
	}
	return nil
}

// SerializeEvent marshals an event into a json byte array
func (h *Handler) SerializeEvent(event eventsourcing.Event) ([]byte, error) {
	e := jsonEvent{}
	// Marshal the event data by itself
	data, _ := json.Marshal(event.Data)
	e.Data = data
	e.AggregateType = event.AggregateType
	e.Version = int(event.Version)
	e.AggregateRootID = string(event.AggregateRootID)
	e.Timestamp = event.Timestamp
	e.Reason = event.Reason
	e.MetaData = event.MetaData

	b, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// DeserializeEvent un marshals an byte array into an event
func (h *Handler) DeserializeEvent(v []byte) (event eventsourcing.Event, err error) {
	jsonEvent := jsonEvent{}
	err = json.Unmarshal(v, &jsonEvent)
	if err != nil {
		return
	}
	data := h.eventRegister[jsonEvent.AggregateType+"_"+jsonEvent.Reason]
	err = json.Unmarshal(jsonEvent.Data, &data)
	if err != nil {
		return
	}
	event.Data = data
	event.MetaData = jsonEvent.MetaData
	event.Reason = jsonEvent.Reason
	event.AggregateRootID = eventsourcing.AggregateRootID(jsonEvent.AggregateRootID)
	event.Timestamp = jsonEvent.Timestamp
	event.Version = eventsourcing.Version(jsonEvent.Version)
	event.AggregateType = jsonEvent.AggregateType
	return
}

// SerializeSnapshot marshals an aggregate as interface{} to []byte
func (h *Handler) SerializeSnapshot(aggregate interface{}) ([]byte, error) {
	data, err := json.Marshal(aggregate)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// DeserializeSnapshot unmarshals []byte to an aggregate
func (h *Handler) DeserializeSnapshot(data []byte, a interface{}) error {
	return json.Unmarshal(data, a)
}

package json

import (
	"encoding/json"
	"github.com/hallgren/eventsourcing"
)

type Handler struct {}

// New returns a json Handle
func New() *Handler {
	return &Handler{}
}

// Serialize marshals an event into a json byte array
func (h *Handler) Serialize(event eventsourcing.Event) ([]byte, error) {
	b, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}
	return b,nil
}

// Deserialize un marshals an byte array into an event
func (h *Handler) Deserialize(v []byte) (event eventsourcing.Event, err error)  {
	err = json.Unmarshal(v, &event)
	return
}

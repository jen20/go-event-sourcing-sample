package memory

import (
	"fmt"
	"github.com/hallgren/eventsourcing"
)

type Handler struct {
	store map[eventsourcing.AggregateRootID]interface{}
}

func New() *Handler {
	return &Handler{
		store: make(map[eventsourcing.AggregateRootID]interface{}),
	}
}

func (h *Handler) Get(id eventsourcing.AggregateRootID) (interface{}, error) {
	aggregate, ok := h.store[id]
	if !ok {
		return nil, fmt.Errorf("aggreagate not found")
	}
	return aggregate, nil
}

func (h *Handler) Save(id eventsourcing.AggregateRootID, a interface{}) error  {
	if id == "" {
		return fmt.Errorf("aggregate id is empty")
	}
	h.store[id] = a
	return nil
}
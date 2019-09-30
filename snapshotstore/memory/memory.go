package memory

import (
	"fmt"
	"github.com/hallgren/eventsourcing"
)

type Handler struct {
	store map[eventsourcing.AggregateRootID][]byte
	serializer snapshotSerializer
}

type snapshotSerializer interface {
	SerializeSnapshot(interface{})  ([]byte, error)
	DeserializeSnapshot(data []byte, a interface{}) error
}

func New(serializer snapshotSerializer) *Handler {
	return &Handler{
		store: make(map[eventsourcing.AggregateRootID][]byte),
		serializer: serializer,
	}
}

func (h *Handler) Get(id eventsourcing.AggregateRootID, a interface{})  error {
	data, ok := h.store[id]
	if !ok {
		return fmt.Errorf("aggreagate not found")
	}
	err := h.serializer.DeserializeSnapshot(data, a)
	if err != nil {
		return err
	}
	return nil
}

func (h *Handler) Save(id eventsourcing.AggregateRootID, a interface{}) error  {
	if id == "" {
		return fmt.Errorf("aggregate id is empty")
	}
	data, err := h.serializer.SerializeSnapshot(a)
	if err != nil {
		return err
	}
	h.store[id] = data
	return nil
}
package snapshotstore

import (
	"fmt"
)

type Handler struct {
	store      map[string][]byte
	serializer snapshotSerializer
}

type snapshotSerializer interface {
	SerializeSnapshot(interface{}) ([]byte, error)
	DeserializeSnapshot(data []byte, a interface{}) error
}

var SnapshotNotFoundError = fmt.Errorf("snapshot not found")

func New(serializer snapshotSerializer) *Handler {
	return &Handler{
		store:      make(map[string][]byte),
		serializer: serializer,
	}
}

func (h *Handler) Get(id string, a interface{}) error {
	data, ok := h.store[id]
	if !ok {
		return SnapshotNotFoundError
	}
	err := h.serializer.DeserializeSnapshot(data, a)
	if err != nil {
		return err
	}
	return nil
}

func (h *Handler) Save(id string, a interface{}) error {
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

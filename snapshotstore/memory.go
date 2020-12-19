package snapshotstore

import (
	"errors"
)

type value struct {
	data []byte
	version int
}

// Handler of snapshot store
type Handler struct {
	store      map[string]value
	serializer snapshotSerializer
}

// Snapshot interface
type Snapshot interface {
	ID() string
	SetID(id string) error
	Version() int
	SetVersion(version int)
}

type snapshotSerializer interface {
	SerializeSnapshot(s Snapshot) ([]byte, error)
	DeserializeSnapshot(data []byte, s Snapshot) error
}

// ErrSnapshotNotFound returns if snapshot not found
var ErrSnapshotNotFound = errors.New("snapshot not found")

// New handler for the snapshot service
func New(serializer snapshotSerializer) *Handler {
	return &Handler{
		store:      make(map[string]value),
		serializer: serializer,
	}
}

// Get returns the deserialize snapshot
func (h *Handler) Get(id string, s Snapshot) error {
	v, ok := h.store[id]
	if !ok {
		return ErrSnapshotNotFound
	}
	err := h.serializer.DeserializeSnapshot(v.data, s)
	if err != nil {
		return err
	}

	s.SetID(id)
	s.SetVersion(v.version)
	return nil
}

// Save persists the snapshot
func (h *Handler) Save(s Snapshot) error {
	if s.ID() == "" {
		return errors.New("aggregate id is empty")
	}
	data, err := h.serializer.SerializeSnapshot(s)
	if err != nil {
		return err
	}
	h.store[s.ID()] = value{data: data, version: s.Version()}
	return nil
}

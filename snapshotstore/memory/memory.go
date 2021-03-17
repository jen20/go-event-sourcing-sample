package memory

import (
	"github.com/hallgren/eventsourcing"
	"github.com/hallgren/eventsourcing/snapshotstore"
)

// Handler of snapshot store
type Handler struct {
	store      map[string]snapshot
	serializer eventsourcing.Serializer
}

type snapshot struct {
	State         []byte
	Version       eventsourcing.Version
	GlobalVersion eventsourcing.Version
}

// New handler for the snapshot service
func New(serializer eventsourcing.Serializer) *Handler {
	return &Handler{
		store:      make(map[string]snapshot),
		serializer: serializer,
	}
}

// Get returns the deserialize snapshot
func (h *Handler) Get(id string, s eventsourcing.Aggregate) error {
	v, ok := h.store[id]
	if !ok {
		return eventsourcing.ErrSnapshotNotFound
	}

	err := h.serializer.Unmarshal(v.State, s)
	if err != nil {
		return err
	}
	// TODO: include when version and globalVersion in not exported on the AggregateRoot
	// r := s.Root()
	// r.AggregateVersion = v.Version
	// r.AggregateGlobalVersion = v.GlobalVersion
	return nil
}

// Save persists the snapshot
func (h *Handler) Save(a eventsourcing.Aggregate) error {
	root := a.Root()
	err := snapshotstore.Validate(*root)
	if err != nil {
		return err
	}

	data, err := h.serializer.Marshal(a)
	if err != nil {
		return err
	}

	s := snapshot{
		State:         data,
		Version:       root.AggregateVersion,
		GlobalVersion: root.AggregateGlobalVersion,
	}

	h.store[root.ID()] = s
	return nil
}

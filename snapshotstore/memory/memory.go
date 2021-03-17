package memory

import (
	"github.com/hallgren/eventsourcing"
	"github.com/hallgren/eventsourcing/snapshotstore"
)

// Handler of snapshot store
type Handler struct {
	store      map[string]eventsourcing.Snapshot
	serializer eventsourcing.Serializer
}

// New handler for the snapshot service
func New(serializer eventsourcing.Serializer) *Handler {
	return &Handler{
		store:      make(map[string]eventsourcing.Snapshot),
		serializer: serializer,
	}
}

// Get returns the deserialize snapshot
func (h *Handler) Get(id string, s eventsourcing.Aggregate) error {
	v, ok := h.store[id]
	if !ok {
		return eventsourcing.ErrSnapshotNotFound
	}

	// unmarhal data into aggregate
	err := h.serializer.Unmarshal(v.State, s)
	if err != nil {
		return err
	}
	r := s.Root()
	r.BuildFromSnapshot(s, v)
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

	s := eventsourcing.Snapshot{
		State:         data,
		Version:       root.Version(),
		GlobalVersion: root.GlobalVersion(),
	}

	h.store[root.ID()] = s
	return nil
}

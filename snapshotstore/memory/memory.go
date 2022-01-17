package memory

import (
	"context"
	"fmt"

	"github.com/hallgren/eventsourcing"
)

// Handler of snapshot store
type Handler struct {
	store map[string]eventsourcing.Snapshot
}

// New handler for the snapshot service
func New() *Handler {
	return &Handler{
		store: make(map[string]eventsourcing.Snapshot),
	}
}

// Get returns the deserialize snapshot
func (h *Handler) Get(ctx context.Context, id, typ string) (eventsourcing.Snapshot, error) {
	v, ok := h.store[fmt.Sprintf("%s_%s", id, typ)]
	if !ok {
		return eventsourcing.Snapshot{}, eventsourcing.ErrSnapshotNotFound
	}
	return v, nil
}

// Save persists the snapshot
func (h *Handler) Save(s eventsourcing.Snapshot) error {
	h.store[fmt.Sprintf("%s_%s", s.ID, s.Type)] = s
	return nil
}

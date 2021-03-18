package eventsourcing

import (
	"fmt"
	"reflect"
)

// Snapshot holds current state of an aggregate
type Snapshot struct {
	ID            string
	Type          string
	State         []byte
	Version       Version
	GlobalVersion Version
}

// SnapshotHandler gets and saves snapshots
type SnapshotHandler struct {
	snapshotStore SnapshotStore
	serializer    Serializer
}

// SnapshotNew constructs a SnapshotHandler
func SnapshotNew(ss SnapshotStore, ser Serializer) *SnapshotHandler {
	return &SnapshotHandler{
		snapshotStore: ss,
		serializer:    ser,
	}
}

// Save transform an aggregate to a snapshot
func (s *SnapshotHandler) Save(a Aggregate) error {
	root := a.Root()
	err := validate(*root)
	if err != nil {
		return err
	}
	typ := reflect.TypeOf(a).Elem().Name()
	b, err := s.serializer.Marshal(a)
	if err != nil {
		return err
	}
	snap := Snapshot{
		ID:            root.ID(),
		Type:          typ,
		Version:       root.Version(),
		GlobalVersion: root.GlobalVersion(),
		State:         b,
	}
	return s.snapshotStore.Save(snap)
}

// Get fetch a snapshot and reconstruct an aggregate
func (s *SnapshotHandler) Get(id string, a Aggregate) error {
	typ := reflect.TypeOf(a).Elem().Name()
	snap, err := s.snapshotStore.Get(id, typ)
	if err != nil {
		return err
	}
	err = s.serializer.Unmarshal(snap.State, a)
	if err != nil {
		return err
	}
	root := a.Root()
	root.setInternals(snap.ID, snap.Version, snap.GlobalVersion)
	return nil
}

// validate make sure the aggregate is valid to be saved
func validate(root AggregateRoot) error {
	if root.ID() == "" {
		return fmt.Errorf("empty id")
	}
	if root.UnsavedEvents() {
		return fmt.Errorf("unsaved events")
	}
	return nil
}

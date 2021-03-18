package eventsourcing

import (
	"fmt"
	"reflect"
)

type Snapshot struct {
	ID            string
	Type          string
	State         []byte
	Version       Version
	GlobalVersion Version
}

type S struct {
	snapshotStore SnapshotStore
	serializer    Serializer
}

func SnapshotNew(ss SnapshotStore, ser Serializer) *S {
	return &S{
		snapshotStore: ss,
		serializer:    ser,
	}
}

func (s *S) Save(a Aggregate) error {
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

func (s *S) Get(id string, a Aggregate) error {
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

func validate(root AggregateRoot) error {
	if root.ID() == "" {
		return fmt.Errorf("empty id")
	}
	if root.UnsavedEvents() {
		return fmt.Errorf("unsaved events")
	}
	return nil
}

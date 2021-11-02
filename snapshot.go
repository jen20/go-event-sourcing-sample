package eventsourcing

import (
	"errors"
	"reflect"
)

// ErrEmptyID indicates that the aggregate ID was empty
var ErrEmptyID = errors.New("aggregate id is empty")

// ErrUnsavedEvents aggregate events must be saved before creating snapshot
var ErrUnsavedEvents = errors.New("aggregate holds unsaved events")

// Snapshot holds current state of an aggregate
type Snapshot struct {
	ID            string
	Type          string
	State         []byte
	Version       Version
	GlobalVersion Version
}

// SnapshotAggregate is an Aggregate plus extra methods to help serialize into a snapshot
type SnapshotAggregate interface {
	Aggregate
	Marshal() (interface{}, error)
	UnMarshal(i interface{}) error
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
func (s *SnapshotHandler) Save(i interface{}) error {
	sa, ok := i.(SnapshotAggregate)
	if ok {
		return s.saveSnapshotAggregate(sa)
	}
	a, ok := i.(Aggregate)
	if ok {
		return s.saveAggregate(a)
	}
	return errors.New("not an aggregate")
}

func (s *SnapshotHandler) saveSnapshotAggregate(sa SnapshotAggregate) error {
	return nil
}

func (s *SnapshotHandler) saveAggregate(sa Aggregate) error {
	root := sa.Root()
	err := validate(*root)
	if err != nil {
		return err
	}
	typ := reflect.TypeOf(sa).Elem().Name()
	b, err := s.serializer.Marshal(sa)
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
		return ErrEmptyID
	}
	if root.UnsavedEvents() {
		return ErrUnsavedEvents
	}
	return nil
}

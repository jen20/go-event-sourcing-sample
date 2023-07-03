package eventsourcing

import (
	"context"
	"errors"
	"reflect"

	"github.com/hallgren/eventsourcing/base"
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
	Version       base.Version
	GlobalVersion base.Version
}

// SnapshotAggregate is an Aggregate plus extra methods to help serialize into a snapshot
type SnapshotAggregate interface {
	Aggregate
	Marshal(m base.MarshalSnapshotFunc) ([]byte, error)
	Unmarshal(m base.UnmarshalSnapshotFunc, b []byte) error
}

// SnapshotHandler gets and saves snapshots
type SnapshotHandler struct {
	snapshotStore SnapshotStore
	serializer    base.Serializer
}

// SnapshotNew constructs a SnapshotHandler
func SnapshotNew(ss SnapshotStore, ser base.Serializer) *SnapshotHandler {
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
	root := sa.Root()
	err := validate(*root)
	if err != nil {
		return err
	}
	typ := reflect.TypeOf(sa).Elem().Name()
	b, err := sa.Marshal(s.serializer.Marshal)
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
func (s *SnapshotHandler) Get(ctx context.Context, id string, i interface{}) error {
	typ := reflect.TypeOf(i).Elem().Name()
	snap, err := s.snapshotStore.Get(ctx, id, typ)
	if err != nil {
		return err
	}
	switch a := i.(type) {
	case SnapshotAggregate:
		err := a.Unmarshal(s.serializer.Unmarshal, snap.State)
		if err != nil {
			return err
		}
		root := a.Root()
		root.setInternals(snap.ID, snap.Version, snap.GlobalVersion)
	case Aggregate:
		err = s.serializer.Unmarshal(snap.State, a)
		if err != nil {
			return err
		}
		root := a.Root()
		root.setInternals(snap.ID, snap.Version, snap.GlobalVersion)
	default:
		return errors.New("not an aggregate")
	}
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

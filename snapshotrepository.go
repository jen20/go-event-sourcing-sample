package eventsourcing

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"

	"github.com/hallgren/eventsourcing/core"
)

// ErrUnsavedEvents aggregate events must be saved before creating snapshot
var ErrUnsavedEvents = errors.New("aggregate holds unsaved events")

type SnapshotRepository struct {
	eventRepository *EventRepository
	snapshotStore   core.SnapshotStore
	Serializer      MarshalFunc
	Deserializer    UnmarshalFunc
}

// NewSnapshotRepository factory function
func NewSnapshotRepository(snapshotStore core.SnapshotStore, eventRepo *EventRepository) *SnapshotRepository {
	return &SnapshotRepository{
		snapshotStore:   snapshotStore,
		eventRepository: eventRepo,
		Serializer:      json.Marshal,
		Deserializer:    json.Unmarshal,
	}
}

// EventRepository return the underlaying event repository. If the user wants to operate on the event repository
// and not use snapshot
func (s *SnapshotRepository) EventRepository() *EventRepository {
	return s.eventRepository
}

func (s *SnapshotRepository) GetWithContext(ctx context.Context, id string, a aggregate) error {
	err := s.GetSnapshot(ctx, id, a)
	if err != nil {
		return err
	}

	// Append events that could have been saved after the snapshot
	return s.eventRepository.GetWithContext(ctx, id, a)
}

// GetSnapshot return aggregate that is based on the snapshot data
// Beware that it could be more events that has happened after the snapshot was taken
func (s *SnapshotRepository) GetSnapshot(ctx context.Context, id string, a aggregate) error {
	if reflect.ValueOf(a).Kind() != reflect.Ptr {
		return ErrAggregateNeedsToBeAPointer
	}

	snapshot, err := s.snapshotStore.Get(ctx, id, aggregateType(a))
	if err != nil && !errors.Is(err, core.ErrSnapshotNotFound) {
		return err
	}
	// Snapshot found
	if err == nil {
		err = s.Deserializer(snapshot.State, a)
		if err != nil {
			return err
		}

		// set the internal aggregate properties
		root := a.Root()
		root.aggregateGlobalVersion = Version(snapshot.GlobalVersion)
		root.aggregateVersion = Version(snapshot.Version)
		root.aggregateID = snapshot.ID
	}
	return nil
}

// Save will save aggregate events and snapshot
func (s *SnapshotRepository) Save(a aggregate) error {
	// make sure events are stored
	err := s.eventRepository.Save(a)
	if err != nil {
		return err
	}

	return s.SaveSnapshot(a)
}

// SaveSnapshot will only store the snapshot and will return error if there is events that is not stored
func (s *SnapshotRepository) SaveSnapshot(a aggregate) error {
	root := a.Root()
	if len(root.Events()) > 0 {
		return ErrUnsavedEvents
	}

	state, err := s.Serializer(a)
	if err != nil {
		return err
	}

	snapshot := core.Snapshot{
		ID:            root.ID(),
		Type:          aggregateType(a),
		Version:       core.Version(root.Version()),
		GlobalVersion: core.Version(root.GlobalVersion()),
		State:         state,
	}

	err = s.snapshotStore.Save(root.ID(), aggregateType(a), snapshot)
	if err != nil {
		return err
	}
	return nil
}

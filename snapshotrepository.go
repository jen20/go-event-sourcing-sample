package eventsourcing

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/hallgren/eventsourcing/core"
)

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

func (s *SnapshotRepository) GetWithContext(ctx context.Context, id string, a aggregate) error {
	if reflect.ValueOf(a).Kind() != reflect.Ptr {
		return errors.New("aggregate needs to be a pointer")
	}

	aggregateType := aggregateType(a)

	snapshotData, err := s.snapshotStore.Get(ctx, id, aggregateType)
	if err != nil {
		return err
	}

	err = s.Deserializer(snapshotData, a)
	if err != nil {
		return err
	}

	// Append events that could have been saved after the snapshot
	return s.eventRepository.GetWithContext(ctx, id, a)
}

// Save will save aggregate event and snapshot
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
		return fmt.Errorf("not ok to store snapshot with unsaved events")
	}

	snapshot, err := s.Serializer(a)
	if err != nil {
		return err
	}

	err = s.snapshotStore.Save(root.ID(), aggregateType(a), snapshot)
	if err != nil {
		return err
	}
	return nil
}

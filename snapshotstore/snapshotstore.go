package snapshotstore

import (
	"errors"
	"github.com/hallgren/eventsourcing"
)

// ErrEmptyID indicates that the aggregate ID was empty
var ErrEmptyID = errors.New("aggregate id is empty")

// ErrUnsavedEvents aggregate events must be saved before creating snapshot
var ErrUnsavedEvents = errors.New("aggregate holds unsaved events")

// Validate that the aggregate is valid
func Validate(root eventsourcing.AggregateRoot) error {
	if root.ID() == "" {
		return ErrEmptyID
	}
	if root.UnsavedEvents() {
		return ErrUnsavedEvents
	}
	return nil
}

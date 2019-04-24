package eventstore

import (
	"errors"
	"gitlab.se.axis.com/morganh/eventsourcing"
)

// ErrEventMultipleAggregates when events holds different id
var ErrEventMultipleAggregates = errors.New("events holds events for more than one aggregate")

// ErrEventMultipleAggregateTypes when events holds different aggregate types
var ErrEventMultipleAggregateTypes = errors.New("events holds events for more than one aggregate type")

// ErrConcurrency when the currently saved version of the aggregate differs from the new ones
var ErrConcurrency = errors.New("concurrency error")

// ErrReasonMissing when the reason is not present in the events
var ErrReasonMissing = errors.New("event holds no reason")

// ValidateEvents make sure the incoming events are valid
func ValidateEvents(aggregateID eventsourcing.AggregateRootID, currentVersion eventsourcing.Version, events []eventsourcing.Event) error {
	aggregateType := events[0].AggregateType

	for _, event := range events {
		if event.AggregateRootID != aggregateID {
			return ErrEventMultipleAggregates
		}

		if event.AggregateType != aggregateType {
			return ErrEventMultipleAggregateTypes
		}

		if currentVersion+1 != event.Version {
			return ErrConcurrency
		}

		if event.Reason == "" {
			return ErrReasonMissing
		}

		currentVersion = event.Version
	}
	return nil
}

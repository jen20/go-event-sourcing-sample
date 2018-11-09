package eventstore

import (
	"fmt"
	"go-event-sourcing-sample/pkg/eventsourcing"
)

// BucketName generate a bucketname from aggregateType and aggregateID
func BucketName(aggregateType, aggregateID string) string {
	return aggregateType + "_" + aggregateID
}

// ValidateEvents make sure the incoming events are valid
func ValidateEvents(aggregateID eventsourcing.AggregateRootID, currentVersion eventsourcing.Version, events []eventsourcing.Event) (bool, error) {
	aggregateType := events[0].AggregateType

	for _, event := range events {
		if event.AggregateRootID != aggregateID {
			return false, fmt.Errorf("events holds events for more than one aggregate")
		}

		if event.AggregateType != aggregateType {
			return false, fmt.Errorf("events holds events for more than one aggregate type")
		}

		if currentVersion+1 != event.Version {
			return false, fmt.Errorf("concurrency error")
		}

		if event.Reason == "" {
			return false, fmt.Errorf("event holds no reason")
		}

		currentVersion = event.Version
	}
	return true, nil
}

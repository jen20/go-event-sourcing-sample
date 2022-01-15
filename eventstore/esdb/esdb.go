package esdb

import (
	"context"
	"errors"

	"github.com/hallgren/eventsourcing/eventstore"

	"github.com/EventStore/EventStore-Client-Go/esdb"
	"github.com/hallgren/eventsourcing"
)

const streamSeparator = "_"

// ESDB is the event store handler
type ESDB struct {
	client      *esdb.Client
	serializer  eventsourcing.Serializer
	contentType esdb.ContentType
}

// Open binds the event store db client
func Open(client *esdb.Client, serializer eventsourcing.Serializer, jsonSerializer bool) *ESDB {
	// defaults to binary
	var contentType esdb.ContentType
	if jsonSerializer {
		contentType = esdb.JsonContentType
	}
	return &ESDB{
		client:      client,
		serializer:  serializer,
		contentType: contentType,
	}
}

// Save persists events to the database
func (es *ESDB) Save(events []eventsourcing.Event) error {
	// If no event return no error
	if len(events) == 0 {
		return nil
	}

	var streamOptions esdb.AppendToStreamOptions
	aggregateID := events[0].AggregateID
	aggregateType := events[0].AggregateType
	version := events[0].Version
	stream := stream(aggregateType, aggregateID)

	err := eventstore.ValidateEventsNoVersionCheck(aggregateID, events)
	if err != nil {
		return err
	}

	esdbEvents := make([]esdb.EventData, len(events))

	for i, event := range events {
		var e, m []byte

		e, err := es.serializer.Marshal(event.Data)
		if err != nil {
			return err
		}
		if event.Metadata != nil {
			m, err = es.serializer.Marshal(event.Metadata)
			if err != nil {
				return err
			}
		}
		eventData := esdb.EventData{
			ContentType: es.contentType,
			EventType:   event.Reason(),
			Data:        e,
			Metadata:    m,
		}

		esdbEvents[i] = eventData
	}

	if version > 1 {
		// StreamRevision value -2 due to version in the eventsourcing pkg start on 1 but in esdb on 0
		// and also the AppendToStream streamOptions expected revision is one version before the first appended event.
		streamOptions.ExpectedRevision = esdb.StreamRevision{Value: uint64(version) - 2}
	} else if version == 1 {
		streamOptions.ExpectedRevision = esdb.NoStream{}
	}
	wr, err := es.client.AppendToStream(context.Background(), stream, streamOptions, esdbEvents...)
	if err != nil {
		return err
	}
	for i := range events {
		// Set all events GlobalVersion to the last events commit position.
		events[i].GlobalVersion = eventsourcing.Version(wr.CommitPosition)
	}
	return nil
}

func (es *ESDB) Get(id string, aggregateType string, afterVersion eventsourcing.Version) (eventsourcing.EventIterator, error) {
	streamID := stream(aggregateType, id)

	from := esdb.StreamRevision{Value: uint64(afterVersion)}
	stream, err := es.client.ReadStream(context.Background(), streamID, esdb.ReadStreamOptions{From: from}, ^uint64(0))
	if err != nil {
		if errors.Is(err, esdb.ErrStreamNotFound) {
			return nil, eventsourcing.ErrNoEvents
		}
		return nil, err
	}
	return &iterator{stream: stream, serializer: es.serializer}, nil
}

func stream(aggregateType, aggregateID string) string {
	return aggregateType + streamSeparator + aggregateID
}

package esdb

import (
	"context"
	"errors"
	"io"
	"strings"

	"github.com/hallgren/eventsourcing/eventstore"

	"github.com/EventStore/EventStore-Client-Go/esdb"
	"github.com/hallgren/eventsourcing"
)

const streamSeparator = "_"

type ESDB struct {
	client      *esdb.Client
	serializer  eventsourcing.Serializer
	contentType esdb.ContentType
}

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

func (es *ESDB) Get(id string, aggregateType string, afterVersion eventsourcing.Version) ([]eventsourcing.Event, error) {
	var events []eventsourcing.Event
	streamID := stream(aggregateType, id)

	from := esdb.StreamRevision{Value: uint64(afterVersion)}
	stream, err := es.client.ReadStream(context.Background(), streamID, esdb.ReadStreamOptions{From: from}, ^uint64(0))
	if err != nil {
		if errors.Is(err, esdb.ErrStreamNotFound) {
			return nil, eventsourcing.ErrNoEvents
		}
		return nil, err
	}
	defer stream.Close()

	for {
		var eventMetadata map[string]interface{}
		event, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}

		stream := strings.Split(event.Event.StreamID, streamSeparator)
		f, ok := es.serializer.Type(stream[0], event.Event.EventType)
		if !ok {
			// if the typ/reason is not register jump over the event
			continue
		}
		eventData := f()
		err = es.serializer.Unmarshal(event.Event.Data, &eventData)
		if err != nil {
			return nil, err
		}
		if event.Event.UserMetadata != nil {
			err = es.serializer.Unmarshal(event.Event.UserMetadata, &eventMetadata)
			if err != nil {
				return nil, err
			}
		}
		events = append(events, eventsourcing.Event{
			AggregateID:   stream[1],
			Version:       eventsourcing.Version(event.Event.EventNumber) + 1, // +1 as the eventsourcing Version starts on 1 but the esdb event version starts on 0
			AggregateType: stream[0],
			Timestamp:     event.Event.CreatedDate,
			Data:          eventData,
			Metadata:      eventMetadata,
			// Can't get the global version when using the ReadStream method
			//GlobalVersion: eventsourcing.Version(event.Event.Position.Commit),
		})
	}
	return events, nil
}

func stream(aggregateType, aggregateID string) string {
	return aggregateType + streamSeparator + aggregateID
}

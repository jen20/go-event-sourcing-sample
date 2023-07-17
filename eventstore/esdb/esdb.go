package esdb

import (
	"context"

	"github.com/hallgren/eventsourcing/base"

	"github.com/EventStore/EventStore-Client-Go/v3/esdb"
)

const streamSeparator = "-"

// ESDB is the event store handler
type ESDB struct {
	client      *esdb.Client
	contentType esdb.ContentType
}

// Open binds the event store db client
func Open(client *esdb.Client, jsonSerializer bool) *ESDB {
	// defaults to binary
	var contentType esdb.ContentType
	if jsonSerializer {
		contentType = esdb.ContentTypeJson
	}
	return &ESDB{
		client:      client,
		contentType: contentType,
	}
}

// Save persists events to the database
func (es *ESDB) Save(events []base.Event) error {
	// If no event return no error
	if len(events) == 0 {
		return nil
	}

	var streamOptions esdb.AppendToStreamOptions
	aggregateID := events[0].AggregateID
	aggregateType := events[0].AggregateType
	version := events[0].Version
	stream := stream(aggregateType, aggregateID)

	err := base.ValidateEventsNoVersionCheck(aggregateID, events)
	if err != nil {
		return err
	}

	esdbEvents := make([]esdb.EventData, len(events))

	for i, event := range events {
		/*
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
		*/
		eventData := esdb.EventData{
			ContentType: es.contentType,
			EventType:   event.Reason,
			Data:        event.Data,
			Metadata:    event.Metadata,
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
		events[i].GlobalVersion = base.Version(wr.CommitPosition)
	}
	return nil
}

func (es *ESDB) Get(ctx context.Context, id string, aggregateType string, afterVersion base.Version) (base.EventIterator, error) {
	streamID := stream(aggregateType, id)

	from := esdb.StreamRevision{Value: uint64(afterVersion)}
	stream, err := es.client.ReadStream(ctx, streamID, esdb.ReadStreamOptions{From: from}, ^uint64(0))
	if err != nil {
		if err, ok := esdb.FromError(err); !ok {
			if err.Code() == esdb.ErrorCodeResourceNotFound {
				return nil, base.ErrNoEvents
			}
		}
		return nil, err
	} else if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	return &iterator{stream: stream}, nil
}

func stream(aggregateType, aggregateID string) string {
	return aggregateType + streamSeparator + aggregateID
}

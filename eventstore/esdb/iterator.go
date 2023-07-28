package esdb

import (
	"errors"
	"io"
	"strings"

	"github.com/EventStore/EventStore-Client-Go/v3/esdb"
	"github.com/hallgren/eventsourcing/core"
)

type iterator struct {
	stream *esdb.ReadStream
}

// Close closes the stream
func (i *iterator) Close() {
	i.stream.Close()
}

// Next returns next event from the stream
func (i *iterator) Next() (core.Event, error) {

	eventESDB, err := i.stream.Recv()
	if errors.Is(err, io.EOF) {
		return core.Event{}, core.ErrNoMoreEvents
	}
	if err, ok := esdb.FromError(err); !ok {
		if err.Code() == esdb.ErrorCodeResourceNotFound {
			return core.Event{}, core.ErrNoMoreEvents
		}
	}
	if err != nil {
		return core.Event{}, err
	}

	stream := strings.Split(eventESDB.Event.StreamID, streamSeparator)

	event := core.Event{
		AggregateID:   stream[1],
		Version:       core.Version(eventESDB.Event.EventNumber) + 1, // +1 as the eventsourcing Version starts on 1 but the esdb event version starts on 0
		AggregateType: stream[0],
		Timestamp:     eventESDB.Event.CreatedDate,
		Data:          eventESDB.Event.Data,
		Metadata:      eventESDB.Event.UserMetadata,
		Reason:        eventESDB.Event.EventType,
		// Can't get the global version when using the ReadStream method
		//GlobalVersion: core.Version(event.Event.Position.Commit),
	}
	return event, nil
}

package sql

import (
	"database/sql"
	"time"

	"github.com/hallgren/eventsourcing/base"
)

type iterator struct {
	rows       *sql.Rows
	serializer base.Serializer
}

// Next return the next event
func (i *iterator) Next() (base.Event, error) {
	var globalVersion base.Version
	var eventMetadata map[string]interface{}
	var version base.Version
	var id, reason, typ, timestamp string
	var data, metadata string
	if !i.rows.Next() {
		if err := i.rows.Err(); err != nil {
			return base.Event{}, err
		}
		return base.Event{}, base.ErrNoMoreEvents
	}
	if err := i.rows.Scan(&globalVersion, &id, &version, &reason, &typ, &timestamp, &data, &metadata); err != nil {
		return base.Event{}, err
	}

	t, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return base.Event{}, err
	}

	f, ok := i.serializer.Type(typ, reason)
	if !ok {
		// if the typ/reason is not register jump over the event
		return i.Next()
	}

	eventData := f()
	err = i.serializer.Unmarshal([]byte(data), &eventData)
	if err != nil {
		return base.Event{}, err
	}
	if metadata != "" {
		err = i.serializer.Unmarshal([]byte(metadata), &eventMetadata)
		if err != nil {
			return base.Event{}, err
		}
	}

	event := base.Event{
		AggregateID:   id,
		Version:       version,
		GlobalVersion: globalVersion,
		AggregateType: typ,
		Timestamp:     t,
		Data:          eventData,
		Metadata:      eventMetadata,
	}
	return event, nil
}

// Close closes the iterator
func (i *iterator) Close() {
	i.rows.Close()
}

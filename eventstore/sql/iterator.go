package sql

import (
	"database/sql"
	"time"

	"github.com/hallgren/eventsourcing/core"
)

type iterator struct {
	rows *sql.Rows
}

// Next return true if there are more data
func (i *iterator) Next() bool {
	return i.rows.Next()
}

// Value return the an event
func (i *iterator) Value() (core.Event, error) {
	var globalVersion core.Version
	var version core.Version
	var id, reason, typ, timestamp string
	var data, metadata []byte

	if err := i.rows.Scan(&globalVersion, &id, &version, &reason, &typ, &timestamp, &data, &metadata); err != nil {
		return core.Event{}, err
	}

	t, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return core.Event{}, err
	}

	event := core.Event{
		AggregateID:   id,
		Version:       version,
		GlobalVersion: globalVersion,
		AggregateType: typ,
		Timestamp:     t,
		Data:          data,
		Metadata:      metadata,
		Reason:        reason,
	}
	return event, nil
}

// Close closes the iterator
func (i *iterator) Close() {
	i.rows.Close()
}

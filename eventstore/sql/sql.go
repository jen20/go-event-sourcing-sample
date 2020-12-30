package sql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/hallgren/eventsourcing/serializer"
	"time"

	"github.com/hallgren/eventsourcing"
	"github.com/hallgren/eventsourcing/eventstore"
)

// SQL for store events
type SQL struct {
	db         sql.DB
	serializer *serializer.Handler
}

// Open connection to database
func Open(db sql.DB, serializer *serializer.Handler) *SQL {
	return &SQL{
		db:         db,
		serializer: serializer,
	}
}

// Close the connection
func (sql *SQL) Close() {
	sql.db.Close()
}

// Save persists events to the database
func (sql *SQL) Save(events []eventsourcing.Event) error {
	// If no event return no error
	if len(events) == 0 {
		return nil
	}
	aggregateID := events[0].AggregateRootID
	aggregateType := events[0].AggregateType

	tx, err := sql.db.BeginTx(context.Background(), nil)
	if err != nil {
		return errors.New(fmt.Sprintf("could not start a write transaction, %v", err))
	}
	defer tx.Rollback()

	// the current version of that is the last event saved
	selectStm := `Select version from events where id=? and type=? order by version desc limit 1`
	rows, err := tx.Query(selectStm, aggregateID, aggregateType)
	if err != nil {
		return err
	}
	defer rows.Close()

	currentVersion := eventsourcing.Version(0)
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return err
		}
		currentVersion = eventsourcing.Version(version)
	}

	//Validate events
	err = eventstore.ValidateEvents(aggregateID, currentVersion, events)
	if err != nil {
		return err
	}

	insert := `Insert into events (id, version, reason, type, timestamp, data, metadata) values ($1, $2, $3, $4, $5, $6, $7)`
	for _, event := range events {
		var e,m []byte

		e, err := sql.serializer.Marshal(event.Data)
		if err != nil {
			return err
		}
		if event.MetaData != nil {
			m, err = sql.serializer.Marshal(event.MetaData)
			if err != nil {
				return err
			}
		}
		_, err = tx.Exec(insert, event.AggregateRootID, event.Version, event.Reason, event.AggregateType, event.Timestamp.Format(time.RFC3339), string(e), string(m))
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

// Get the events from database
func (sql *SQL) Get(id string, aggregateType string, afterVersion eventsourcing.Version) (events []eventsourcing.Event, err error) {
	selectStm := `Select id, version, reason, type, timestamp, data, metadata from events where id=? and type=? and version>? order by version asc`
	rows, err := sql.db.Query(selectStm, id, aggregateType, afterVersion)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var eventMetaData map[string]interface{}
		var version eventsourcing.Version
		var id, reason, typ, timestamp string
		var data, metadata string
		if err := rows.Scan(&id, &version, &reason, &typ, &timestamp, &data, &metadata); err != nil {
			return nil, err
		}

		t, err := time.Parse(time.RFC3339, timestamp)
		if err != nil {
			return nil, err
		}

		f, ok := sql.serializer.Type(typ, reason)
		if !ok {
			// if the typ/reason is not register jump over the event
			continue
		}

		eventData := f()
		err = sql.serializer.Unmarshal([]byte(data), &eventData)
		if err != nil {
			return nil, err
		}
		if metadata != "" {
			err = sql.serializer.Unmarshal([]byte(metadata), &eventMetaData)
			if err != nil {
				return nil, err
			}
		}

		events = append(events, eventsourcing.Event{
			AggregateRootID: id,
			Version: version,
			AggregateType: typ,
			Reason: reason,
			Timestamp: t,
			Data: eventData,
			MetaData: eventMetaData,
		})
	}
	return
}
